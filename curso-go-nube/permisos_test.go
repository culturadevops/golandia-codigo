package main

// Las pruebas de los permisos. La mitad son de la evaluación pura —sin red, en microsegundos— y
// la otra mitad son de verdad: el SDK oficial de AWS pidiendo algo y recibiendo un `AccessDenied`
// con la forma correcta.
//
//	go test ./... -count=1 -race -run Permiso

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	dtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	smithy "github.com/aws/smithy-go"
)

func pol(t *testing.T, crudo string) Politica {
	t.Helper()
	var p Politica
	if err := json.Unmarshal([]byte(crudo), &p); err != nil {
		t.Fatalf("la politica no se entiende: %v\n%s", err, crudo)
	}
	return p
}

// ── la evaluación pura ───────────────────────────────────────────────────

// 🥇 REGLA 1: por defecto NO. Sin un Allow que te toque, no pasas.
func TestPermisoPorDefectoNo(t *testing.T) {
	p := pol(t, `{"Version":"2012-10-17","Statement":[
		{"Effect":"Allow","Action":"s3:GetObject","Resource":"arn:aws:s3:::fotos/*"}]}`)

	if v := Evaluar([]Politica{p}, "s3:PutObject", "arn:aws:s3:::fotos/gato.png"); v.Permitido {
		t.Fatal("dejo hacer algo que ninguna sentencia menciona")
	}
	// Y el motivo se puede leer: eso es media utilidad del error.
	v := Evaluar([]Politica{p}, "s3:PutObject", "arn:aws:s3:::fotos/gato.png")
	if !strings.Contains(v.Motivo, "ninguna politica permite") {
		t.Fatalf("el motivo no dice por que: %q", v.Motivo)
	}
}

// REGLA 2: un Allow explícito deja pasar.
func TestPermisoUnAllowExplicitoDejaPasar(t *testing.T) {
	p := pol(t, `{"Version":"2012-10-17","Statement":[
		{"Sid":"leer","Effect":"Allow","Action":"s3:GetObject","Resource":"arn:aws:s3:::fotos/*"}]}`)
	v := Evaluar([]Politica{p}, "s3:GetObject", "arn:aws:s3:::fotos/gato.png")
	if !v.Permitido {
		t.Fatalf("no dejo pasar algo permitido: %s", v.Motivo)
	}
	if v.Sid != "leer" {
		t.Fatalf("no dice que sentencia decidio: %q", v.Sid)
	}
}

// 🥇 REGLA 3: un Deny explícito GANA SIEMPRE, aunque haya mil Allow y aunque esté escrito
// después.
func TestPermisoUnDenyGanaSiempre(t *testing.T) {
	p := pol(t, `{"Version":"2012-10-17","Statement":[
		{"Sid":"todo","Effect":"Allow","Action":"s3:*","Resource":"*"},
		{"Sid":"menos-esto","Effect":"Deny","Action":"s3:DeleteBucket","Resource":"*"}]}`)

	if v := Evaluar([]Politica{p}, "s3:PutObject", "arn:aws:s3:::fotos/x"); !v.Permitido {
		t.Fatalf("el Allow general no funciono: %s", v.Motivo)
	}
	v := Evaluar([]Politica{p}, "s3:DeleteBucket", "arn:aws:s3:::fotos")
	if v.Permitido {
		t.Fatal("un Allow general se comio un Deny explicito")
	}
	if v.Sid != "menos-esto" {
		t.Fatalf("no dice que Deny decidio: %q", v.Sid)
	}
}

// 🥇 Y gana aunque esté escrito ANTES. El orden en el fichero no cambia el resultado: por eso el
// Deny se busca entero antes de mirar ningún Allow.
func TestPermisoElOrdenEnElFicheroNoImporta(t *testing.T) {
	alReves := pol(t, `{"Version":"2012-10-17","Statement":[
		{"Effect":"Deny","Action":"s3:DeleteBucket","Resource":"*"},
		{"Effect":"Allow","Action":"s3:*","Resource":"*"}]}`)
	derecho := pol(t, `{"Version":"2012-10-17","Statement":[
		{"Effect":"Allow","Action":"s3:*","Resource":"*"},
		{"Effect":"Deny","Action":"s3:DeleteBucket","Resource":"*"}]}`)

	for nombre, p := range map[string]Politica{"al reves": alReves, "derecho": derecho} {
		if Evaluar([]Politica{p}, "s3:DeleteBucket", "arn:aws:s3:::fotos").Permitido {
			t.Fatalf("con las sentencias %s, el Deny no gano", nombre)
		}
	}
}

// Y en políticas SEPARADAS también: un Deny en una prohíbe lo que otra permite.
func TestPermisoUnDenyEnOtraPoliticaTambienGana(t *testing.T) {
	deja := pol(t, `{"Version":"2012-10-17","Statement":[
		{"Effect":"Allow","Action":"*","Resource":"*"}]}`)
	prohibe := pol(t, `{"Version":"2012-10-17","Statement":[
		{"Effect":"Deny","Action":"s3:DeleteBucket","Resource":"*"}]}`)

	if Evaluar([]Politica{deja, prohibe}, "s3:DeleteBucket", "arn:aws:s3:::x").Permitido {
		t.Fatal("un Deny de otra politica no gano")
	}
}

// 🥇 EL COMODÍN NO SABE DE SERVICIOS: sabe de letras. `s3:*` incluye `s3:DeleteBucket`, y
// `s3:Get*` incluye `s3:GetBucketNotification`, que probablemente no era lo que querías.
func TestPermisoElComodinNoSabeDeServicios(t *testing.T) {
	casos := []struct {
		patron, accion string
		quiero         bool
	}{
		{"s3:*", "s3:GetObject", true},
		{"s3:*", "s3:DeleteBucket", true}, // ⚠️ esto también
		{"s3:Get*", "s3:GetObject", true},
		{"s3:Get*", "s3:GetBucketNotification", true}, // ⚠️ y esto
		{"s3:Get*", "s3:PutObject", false},
		{"*", "dynamodb:PutItem", true},
		{"s3:GetObject", "s3:GetObjectAcl", false},
		{"s3:Get?bject", "s3:GetObject", true},
		// El punto es un punto, no «cualquier carácter»: no son expresiones regulares.
		{"s3:Get.bject", "s3:GetObject", false},
	}
	for _, c := range casos {
		if got := encaja(c.patron, c.accion); got != c.quiero {
			t.Errorf("encaja(%q, %q) = %v, quiero %v", c.patron, c.accion, got, c.quiero)
		}
	}
}

// El comodín en el recurso: la trampa de la barra. `arn:aws:s3:::fotos/*` NO incluye el bucket
// en sí, y por eso hacen falta las dos formas.
func TestPermisoElBucketYSusObjetosSonDosRecursos(t *testing.T) {
	p := pol(t, `{"Version":"2012-10-17","Statement":[
		{"Effect":"Allow","Action":"s3:*","Resource":"arn:aws:s3:::fotos/*"}]}`)

	if !Evaluar([]Politica{p}, "s3:GetObject", "arn:aws:s3:::fotos/gato.png").Permitido {
		t.Fatal("no dejo leer un objeto")
	}
	// ⚠️ Y esto NO pasa: listar el bucket es sobre `arn:aws:s3:::fotos`, sin barra.
	if Evaluar([]Politica{p}, "s3:ListBucket", "arn:aws:s3:::fotos").Permitido {
		t.Fatal("`fotos/*` dejo hacer algo sobre `fotos`, y son dos recursos distintos")
	}
}

// ⚠️ `Action` y `Resource` pueden ser una CADENA o una LISTA. Las dos formas tienen que leerse.
func TestPermisoUnaCadenaOUnaLista(t *testing.T) {
	unaSola := pol(t, `{"Version":"2012-10-17","Statement":[
		{"Effect":"Allow","Action":"s3:GetObject","Resource":"arn:aws:s3:::fotos/*"}]}`)
	varias := pol(t, `{"Version":"2012-10-17","Statement":[
		{"Effect":"Allow","Action":["s3:GetObject","s3:PutObject"],
		 "Resource":["arn:aws:s3:::fotos/*","arn:aws:s3:::copias/*"]}]}`)

	if len(unaSola.Statement[0].Action) != 1 {
		t.Fatalf("la cadena no se leyo como lista de uno: %v", unaSola.Statement[0].Action)
	}
	if len(varias.Statement[0].Resource) != 2 {
		t.Fatalf("la lista no se leyo: %v", varias.Statement[0].Resource)
	}
	if !Evaluar([]Politica{varias}, "s3:PutObject", "arn:aws:s3:::copias/x").Permitido {
		t.Fatal("el segundo recurso de la lista no vale")
	}
}

// Una política con un Effect que no existe se rechaza al leerla, no al evaluarla.
func TestPermisoUnEffectRaroSeRechazaAlLeer(t *testing.T) {
	var p Politica
	err := json.Unmarshal([]byte(`{"Version":"2012-10-17","Statement":[
		{"Effect":"Permitir","Action":"*","Resource":"*"}]}`), &p)
	if err == nil {
		t.Fatal("acepto un Effect que no existe")
	}
	if !strings.Contains(err.Error(), "Allow o Deny") {
		t.Fatalf("el error no dice cuales valen: %v", err)
	}
}

// ── contra el SDK oficial ────────────────────────────────────────────────

func nubeConPermisos(t *testing.T, politica string) (*s3.Client, *dynamodb.Client, *sqs.Client) {
	t.Helper()
	perm := NuevosPermisos(nil)
	perm.Poner("prueba", "lector", []Politica{pol(t, politica)})

	mux := http.NewServeMux()
	NuevoS3(NuevoAlmacen(), nil).Rutas(mux)
	tablas, colas := NuevasTablas(), NuevasColas()
	raiz := Raiz(mux, map[string]http.Handler{
		"DynamoDB_":  NuevoDynamo(tablas, nil),
		"AmazonSQS.": NuevoSQS(colas, nil),
	}, nil, nil)
	srv := httptest.NewServer(NuevoAutorizador(perm, nil).Envolver(raiz))
	t.Cleanup(srv.Close)

	return clienteS3(srv.URL), clienteDynamo(srv.URL), clienteSQS(srv.URL)
}

func codigo(err error) string {
	var api smithy.APIError
	if err != nil && errors.As(err, &api) {
		return api.ErrorCode()
	}
	return ""
}

// 🥇 La prueba de fondo: el SDK oficial pide algo que no puede y recibe un AccessDenied con la
// forma de S3, no un «respuesta inesperada».
func TestPermisoElSDKRecibeUnAccessDeniedDeVerdad(t *testing.T) {
	c, _, _ := nubeConPermisos(t, `{"Version":"2012-10-17","Statement":[
		{"Effect":"Allow","Action":"s3:CreateBucket","Resource":"*"}]}`)
	ctx := context.Background()

	if _, err := c.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String("fotos")}); err != nil {
		t.Fatalf("CreateBucket estaba permitido y fallo: %v", err)
	}
	_, err := c.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String("fotos"), Key: aws.String("gato.png"),
		Body: strings.NewReader("x")})
	if err == nil {
		t.Fatal("subio un objeto sin permiso para hacerlo")
	}
	if c := codigo(err); c != "AccessDenied" {
		t.Fatalf("el error llego como %q y esperaba AccessDenied: %v", c, err)
	}
	// 🥇 Y el mensaje DICE QUÉ FALTA. Esa es la diferencia con SignatureDoesNotMatch.
	if !strings.Contains(err.Error(), "s3:PutObject") {
		t.Fatalf("el mensaje no dice que permiso falta: %v", err)
	}
	if !strings.Contains(err.Error(), "arn:aws:s3:::fotos/gato.png") {
		t.Fatalf("el mensaje no dice sobre que recurso: %v", err)
	}
}

// El mismo método y la misma forma de ruta son operaciones distintas según los tramos.
func TestPermisoCrearBucketNoEsSubirObjeto(t *testing.T) {
	c, _, _ := nubeConPermisos(t, `{"Version":"2012-10-17","Statement":[
		{"Effect":"Allow","Action":"s3:PutObject","Resource":"*"}]}`)
	ctx := context.Background()

	// PUT /fotos -> crear bucket, que NO está permitido.
	if _, err := c.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String("fotos")}); err == nil {
		t.Fatal("`s3:PutObject` dejo crear un bucket, y son dos operaciones distintas")
	} else if !strings.Contains(err.Error(), "s3:CreateBucket") {
		t.Fatalf("no dice que permiso falta: %v", err)
	}
}

// 🥇 Para saber sobre QUÉ recurso se pide algo, a veces hay que leer el cuerpo. Y si lo lees,
// hay que volver a dejarlo puesto: si no, el servicio de detrás se encuentra una petición vacía.
func TestPermisoElCuerpoSeVuelveADejarPuesto(t *testing.T) {
	_, din, _ := nubeConPermisos(t, `{"Version":"2012-10-17","Statement":[
		{"Effect":"Allow","Action":"dynamodb:*","Resource":"*"}]}`)
	ctx := context.Background()

	// Si el cuerpo no se restaurase, esto fallaría con «falta TableName».
	if _, err := din.CreateTable(ctx, &dynamodb.CreateTableInput{
		TableName: aws.String("pedidos"),
		KeySchema: []dtypes.KeySchemaElement{
			{AttributeName: aws.String("cliente"), KeyType: dtypes.KeyTypeHash}},
		AttributeDefinitions: []dtypes.AttributeDefinition{
			{AttributeName: aws.String("cliente"), AttributeType: dtypes.ScalarAttributeTypeS}},
		BillingMode: dtypes.BillingModePayPerRequest,
	}); err != nil {
		t.Fatalf("CreateTable: %v — ¿se leyo el cuerpo y no se volvio a dejar?", err)
	}
	if _, err := din.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String("pedidos"),
		Item: map[string]dtypes.AttributeValue{
			"cliente": &dtypes.AttributeValueMemberS{Value: "ana"}}}); err != nil {
		t.Fatalf("PutItem: %v", err)
	}
}

// Y el recurso que sale del cuerpo se usa de verdad: una política por tabla distingue tablas.
func TestPermisoUnaPoliticaPorTabla(t *testing.T) {
	_, din, _ := nubeConPermisos(t, `{"Version":"2012-10-17","Statement":[
		{"Effect":"Allow","Action":"dynamodb:*",
		 "Resource":"arn:aws:dynamodb:us-east-1:000000000000:table/pedidos"}]}`)
	ctx := context.Background()

	crear := func(nombre string) error {
		_, err := din.CreateTable(ctx, &dynamodb.CreateTableInput{
			TableName: aws.String(nombre),
			KeySchema: []dtypes.KeySchemaElement{
				{AttributeName: aws.String("id"), KeyType: dtypes.KeyTypeHash}},
			AttributeDefinitions: []dtypes.AttributeDefinition{
				{AttributeName: aws.String("id"), AttributeType: dtypes.ScalarAttributeTypeS}},
			BillingMode: dtypes.BillingModePayPerRequest})
		return err
	}
	if err := crear("pedidos"); err != nil {
		t.Fatalf("la tabla permitida fallo: %v", err)
	}
	if err := crear("secretos"); err == nil {
		t.Fatal("creo una tabla que la politica no menciona")
	}
}

// Las colas también: el nombre sale de la URL que va dentro del cuerpo.
func TestPermisoUnaPoliticaPorCola(t *testing.T) {
	_, _, q := nubeConPermisos(t, `{"Version":"2012-10-17","Statement":[
		{"Effect":"Allow","Action":["sqs:CreateQueue","sqs:SendMessage"],
		 "Resource":"arn:aws:sqs:us-east-1:000000000000:trabajos"}]}`)
	ctx := context.Background()

	r, err := q.CreateQueue(ctx, &sqs.CreateQueueInput{QueueName: aws.String("trabajos")})
	if err != nil {
		t.Fatalf("CreateQueue: %v", err)
	}
	if _, err := q.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl: r.QueueUrl, MessageBody: aws.String("hola")}); err != nil {
		t.Fatalf("SendMessage: %v", err)
	}
	// Y recibir NO está permitido.
	if _, err := q.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
		QueueUrl: r.QueueUrl}); err == nil {
		t.Fatal("recibio sin permiso")
	} else if !strings.Contains(err.Error(), "sqs:ReceiveMessage") {
		t.Fatalf("no dice que permiso falta: %v", err)
	}
}

// El error de permisos tiene la forma del SERVICIO al que ibas, no una forma genérica.
func TestPermisoElErrorTieneLaFormaDelServicio(t *testing.T) {
	c, din, _ := nubeConPermisos(t, `{"Version":"2012-10-17","Statement":[
		{"Effect":"Deny","Action":"*","Resource":"*"}]}`)
	ctx := context.Background()

	// S3 contesta texto etiquetado, y el SDK lo convierte en un error tipado.
	if _, err := c.ListBuckets(ctx, &s3.ListBucketsInput{}); err == nil {
		t.Fatal("un Deny total dejo listar")
	} else if codigo(err) != "AccessDenied" {
		t.Fatalf("S3 devolvio %q: %v", codigo(err), err)
	}
	// DynamoDB contesta JSON, y también sale tipado.
	if _, err := din.ListTables(ctx, &dynamodb.ListTablesInput{}); err == nil {
		t.Fatal("un Deny total dejo listar tablas")
	} else if !strings.Contains(err.Error(), "AccessDenied") {
		t.Fatalf("DynamoDB devolvio: %v", err)
	}
}

// Una clave sin políticas no se prohíbe nada: es el modo de los nueve episodios anteriores.
func TestPermisoSinPoliticasNoSeProhibeNada(t *testing.T) {
	perm := NuevosPermisos(nil)
	// ⚠️ Se registra OTRA clave, no la que usa el cliente.
	perm.Poner("otro", "otro", []Politica{pol(t,
		`{"Version":"2012-10-17","Statement":[{"Effect":"Deny","Action":"*","Resource":"*"}]}`)})

	mux := http.NewServeMux()
	NuevoS3(NuevoAlmacen(), nil).Rutas(mux)
	srv := httptest.NewServer(NuevoAutorizador(perm, nil).Envolver(
		Raiz(mux, nil, nil, nil)))
	defer srv.Close()

	if _, err := clienteS3(srv.URL).ListBuckets(context.Background(),
		&s3.ListBucketsInput{}); err != nil {
		t.Fatalf("una clave sin politicas deberia poder todo: %v", err)
	}
}

// Y el filtro por sub-recurso: `?notification` es otra operación con la misma ruta, y eso se
// nota también en los permisos.
func TestPermisoLaConsultaCambiaElPermiso(t *testing.T) {
	c, _, _ := nubeConPermisos(t, `{"Version":"2012-10-17","Statement":[
		{"Effect":"Allow","Action":["s3:CreateBucket","s3:PutObject"],"Resource":"*"}]}`)
	ctx := context.Background()
	if _, err := c.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String("fotos")}); err != nil {
		t.Fatalf("CreateBucket: %v", err)
	}
	// PUT /fotos?notification NO es `s3:CreateBucket` ni `s3:PutObject`.
	_, err := c.PutBucketNotificationConfiguration(ctx,
		&s3.PutBucketNotificationConfigurationInput{
			Bucket:                    aws.String("fotos"),
			NotificationConfiguration: &s3types.NotificationConfiguration{}})
	if err == nil {
		t.Fatal("configuro notificaciones sin ese permiso")
	}
	if !strings.Contains(err.Error(), "s3:PutBucketNotification") {
		t.Fatalf("no dice que permiso falta: %v", err)
	}
}
