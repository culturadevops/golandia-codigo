package main

// Las pruebas de la persistencia. Y como siempre en este proyecto, quien habla es el SDK
// oficial de AWS: se guarda un estado creado con el SDK, se levanta otra nubecita desde el
// fichero, y se pregunta con el SDK otra vez.
//
//	go test ./... -count=1 -race -run Persist

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	dtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

// casa monta unos almacenes vacíos y devuelve la función que arma su instantánea.
func casa() (*Almacen, *Tablas, *Colas, *Funciones, func() Instantanea) {
	a, ts, cs, f := NuevoAlmacen(), NuevasTablas(), NuevasColas(), NuevasFunciones()
	return a, ts, cs, f, func() Instantanea { return Reunir(a, ts, cs, f) }
}

// La prueba de fondo: se llena una nubecita, se guarda, se levanta otra desde el fichero y el
// SDK oficial encuentra lo mismo.
func TestPersistIdaYVuelta(t *testing.T) {
	ruta := filepath.Join(t.TempDir(), "estado.json")

	// ── primera vida ────────────────────────────────────────────────────
	a1, ts1, cs1, f1, reunir1 := casa()
	c1 := clienteS3(servidorCon(t, a1, ts1, cs1, f1))
	ctx := context.Background()
	if _, err := c1.CreateBucket(ctx, &s3.CreateBucketInput{Bucket: aws.String("fotos")}); err != nil {
		t.Fatalf("CreateBucket: %v", err)
	}
	if _, err := c1.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String("fotos"), Key: aws.String("perros/luna.txt"),
		Body: strings.NewReader("una perra que se llama Luna")}); err != nil {
		t.Fatalf("PutObject: %v", err)
	}

	g := NuevaGuarda(ruta, time.Millisecond, reunir1, nil)
	g.Escribir()
	if g.veces != 1 {
		t.Fatalf("escrituras = %d, quiero 1", g.veces)
	}

	// ── segunda vida ────────────────────────────────────────────────────
	inst, err := Cargar(ruta)
	if err != nil {
		t.Fatalf("Cargar: %v", err)
	}
	if inst == nil {
		t.Fatal("el fichero existe y Cargar devolvio nil")
	}
	a2, ts2, cs2, f2, _ := casa()
	if err := Aplicar(inst, a2, ts2, cs2, f2); err != nil {
		t.Fatalf("Aplicar: %v", err)
	}
	c2 := clienteS3(servidorCon(t, a2, ts2, cs2, f2))

	res, err := c2.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String("fotos"), Key: aws.String("perros/luna.txt")})
	if err != nil {
		t.Fatalf("GetObject en la vida nueva: %v", err)
	}
	defer res.Body.Close()
	cuerpo, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("leer el cuerpo: %v", err)
	}
	if string(cuerpo) != "una perra que se llama Luna" {
		t.Fatalf("el contenido volvio como %q", cuerpo)
	}
	// El ETag NO estaba en el fichero: se recalcula al cargar. Si sale bien, es que los bytes
	// son los mismos.
	if aws.ToString(res.ETag) == "" {
		t.Fatal("el ETag no se recalculo al cargar")
	}
	if aws.ToInt64(res.ContentLength) != 27 {
		t.Fatalf("tamaño = %d, quiero 27", aws.ToInt64(res.ContentLength))
	}
}

// 🥇 Un []byte en JSON sale en BASE64. El fichero de estado no se lee a simple vista, y esto lo
// demuestra en vez de contarlo.
func TestPersistLosBytesViajanEnBase64(t *testing.T) {
	ruta := filepath.Join(t.TempDir(), "estado.json")
	a, _, _, _, reunir := casa()
	a.CrearBucket("fotos")
	a.Poner("fotos", "nota.txt", []byte("hola mundo"), "text/plain")

	NuevaGuarda(ruta, time.Millisecond, reunir, nil).Escribir()
	crudo, err := os.ReadFile(ruta)
	if err != nil {
		t.Fatalf("leer: %v", err)
	}
	texto := string(crudo)
	if strings.Contains(texto, "hola mundo") {
		t.Fatal("el contenido salio en claro, y un []byte deberia salir en base64")
	}
	if !strings.Contains(texto, base64.StdEncoding.EncodeToString([]byte("hola mundo"))) {
		t.Fatalf("no encuentro el base64 del contenido en el fichero:\n%s", texto)
	}
}

// 🥇 LA DECISIÓN DEL EPISODIO: un mensaje que estaba invisible vuelve VISIBLE al restaurar. Su
// recibo era una promesa hecha a alguien que ya no existe.
func TestPersistUnMensajeEnVueloVuelveVisible(t *testing.T) {
	ruta := filepath.Join(t.TempDir(), "estado.json")

	a1, ts1, cs1, f1, reunir1 := casa()
	base := servidorCon(t, a1, ts1, cs1, f1)
	q1 := clienteSQS(base)
	ctx := context.Background()
	crear, err := q1.CreateQueue(ctx, &sqs.CreateQueueInput{QueueName: aws.String("trabajos")})
	if err != nil {
		t.Fatalf("CreateQueue: %v", err)
	}
	url := crear.QueueUrl
	if _, err := q1.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl: url, MessageBody: aws.String("cobrar 10 euros")}); err != nil {
		t.Fatalf("SendMessage: %v", err)
	}
	// Alguien lo recibe: ahora está invisible 30 segundos y tiene un recibo.
	rec, err := q1.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{QueueUrl: url})
	if err != nil || len(rec.Messages) != 1 {
		t.Fatalf("ReceiveMessage: %v (%d mensajes)", err, len(rec.Messages))
	}
	// Y ahora mismo la cola no le da ese mensaje a nadie más.
	vacio, _ := q1.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{QueueUrl: url})
	if len(vacio.Messages) != 0 {
		t.Fatal("el mensaje seguia visible despues de recibirlo")
	}

	NuevaGuarda(ruta, time.Millisecond, reunir1, nil).Escribir()

	// El fichero NO puede llevar ni el recibo ni el momento en que vuelve a verse.
	crudo, _ := os.ReadFile(ruta)
	for _, prohibido := range []string{"visibleEn", "recibo", "VisibleEn", "Recibo"} {
		if strings.Contains(string(crudo), prohibido) {
			t.Fatalf("el fichero guarda %q, y eso es una conversacion a medias", prohibido)
		}
	}

	a2, ts2, cs2, f2, _ := casa()
	inst, err := Cargar(ruta)
	if err != nil {
		t.Fatalf("Cargar: %v", err)
	}
	if err := Aplicar(inst, a2, ts2, cs2, f2); err != nil {
		t.Fatalf("Aplicar: %v", err)
	}
	q2 := clienteSQS(servidorCon(t, a2, ts2, cs2, f2))
	// Hay que volver a pedir la URL: la arma el servidor nuevo, con su propio Host.
	u2, err := q2.GetQueueUrl(ctx, &sqs.GetQueueUrlInput{QueueName: aws.String("trabajos")})
	if err != nil {
		t.Fatalf("GetQueueUrl: %v", err)
	}
	otra, err := q2.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{QueueUrl: u2.QueueUrl})
	if err != nil {
		t.Fatalf("ReceiveMessage en la vida nueva: %v", err)
	}
	if len(otra.Messages) != 1 {
		t.Fatal("el mensaje no volvio visible: se quedo escondido para siempre")
	}
	if aws.ToString(otra.Messages[0].Body) != "cobrar 10 euros" {
		t.Fatalf("cuerpo = %q", aws.ToString(otra.Messages[0].Body))
	}
}

// La cuenta de entregas SÍ sobrevive: si no, un mensaje que falla una y otra vez nunca llegaría
// a la cola de muertos, porque cada reinicio le pondría el contador a cero.
func TestPersistLaCuentaDeEntregasSobrevive(t *testing.T) {
	ruta := filepath.Join(t.TempDir(), "estado.json")
	_, _, cs1, _, reunir1 := casa()
	vis := time.Duration(0)
	cs1.Crear("trabajos", &vis, 3, "muertos")
	cs1.Crear("muertos", nil, 0, "")
	if _, err := cs1.Enviar("trabajos", "el que siempre falla"); err != nil {
		t.Fatalf("Enviar: %v", err)
	}
	// Tres entregas: con MaxRecibos = 3, la cuarta es la que lo aparta.
	for i := 0; i < 3; i++ {
		if _, err := cs1.Recibir("trabajos", 1, nil); err != nil {
			t.Fatalf("Recibir: %v", err)
		}
	}

	NuevaGuarda(ruta, time.Millisecond, reunir1, nil).Escribir()
	inst, _ := Cargar(ruta)
	a2, ts2, cs2, f2, _ := casa()
	if err := Aplicar(inst, a2, ts2, cs2, f2); err != nil {
		t.Fatalf("Aplicar: %v", err)
	}

	// Cuarta entrega: con el contador restaurado, esta lo aparta.
	if _, err := cs2.Recibir("trabajos", 1, nil); err != nil {
		t.Fatalf("Recibir tras cargar: %v", err)
	}
	vivos, _, _ := cs2.Cuentas("trabajos")
	muertos, _, _ := cs2.Cuentas("muertos")
	if muertos != 1 {
		t.Fatalf("en muertos hay %d y quiero 1: el contador de entregas se reinicio "+
			"al cargar (vivos=%d)", muertos, vivos)
	}
}

// Las tablas se reconstruyen con el mismo código que las construye. Y los números de DynamoDB,
// que viajan como cadena, vuelven intactos.
func TestPersistLasTablasVuelvenEnteras(t *testing.T) {
	ruta := filepath.Join(t.TempDir(), "estado.json")
	a1, ts1, cs1, f1, reunir1 := casa()
	din1 := clienteDynamo(servidorCon(t, a1, ts1, cs1, f1))
	ctx := context.Background()
	if _, err := din1.CreateTable(ctx, &dynamodb.CreateTableInput{
		TableName: aws.String("pedidos"),
		KeySchema: []dtypes.KeySchemaElement{
			{AttributeName: aws.String("cliente"), KeyType: dtypes.KeyTypeHash},
			{AttributeName: aws.String("fecha"), KeyType: dtypes.KeyTypeRange},
		},
		AttributeDefinitions: []dtypes.AttributeDefinition{
			{AttributeName: aws.String("cliente"), AttributeType: dtypes.ScalarAttributeTypeS},
			{AttributeName: aws.String("fecha"), AttributeType: dtypes.ScalarAttributeTypeS},
		},
		BillingMode: dtypes.BillingModePayPerRequest,
	}); err != nil {
		t.Fatalf("CreateTable: %v", err)
	}
	if _, err := din1.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String("pedidos"),
		Item: map[string]dtypes.AttributeValue{
			"cliente": &dtypes.AttributeValueMemberS{Value: "ana"},
			"fecha":   &dtypes.AttributeValueMemberS{Value: "2026-08-25"},
			// Un identificador de veinte dígitos: si pasara por un float, volvería cambiado.
			"pedido": &dtypes.AttributeValueMemberN{Value: "12345678901234567890"},
		}}); err != nil {
		t.Fatalf("PutItem: %v", err)
	}

	NuevaGuarda(ruta, time.Millisecond, reunir1, nil).Escribir()
	inst, _ := Cargar(ruta)
	a2, ts2, cs2, f2, _ := casa()
	if err := Aplicar(inst, a2, ts2, cs2, f2); err != nil {
		t.Fatalf("Aplicar: %v", err)
	}
	din2 := clienteDynamo(servidorCon(t, a2, ts2, cs2, f2))

	got, err := din2.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String("pedidos"),
		Key: map[string]dtypes.AttributeValue{
			"cliente": &dtypes.AttributeValueMemberS{Value: "ana"},
			"fecha":   &dtypes.AttributeValueMemberS{Value: "2026-08-25"},
		}})
	if err != nil {
		t.Fatalf("GetItem en la vida nueva: %v", err)
	}
	n, ok := got.Item["pedido"].(*dtypes.AttributeValueMemberN)
	if !ok {
		t.Fatalf("el atributo volvio como %T", got.Item["pedido"])
	}
	if n.Value != "12345678901234567890" {
		t.Fatalf("el numero volvio como %q: perdio precision por el camino", n.Value)
	}
}

// 🥇 No se escribe lo mismo dos veces. Y esto solo funciona porque la exportación va ORDENADA:
// el recorrido de un mapa en Go es aleatorio, y sin ordenar el JSON saldría distinto cada vez.
func TestPersistNoReescribeLoMismo(t *testing.T) {
	ruta := filepath.Join(t.TempDir(), "estado.json")
	a, _, _, _, reunir := casa()
	a.CrearBucket("fotos")
	for i := 0; i < 40; i++ {
		a.Poner("fotos", "f"+string(rune('a'+i%20))+".txt", []byte("x"), "text/plain")
	}
	g := NuevaGuarda(ruta, time.Millisecond, reunir, nil)

	for i := 0; i < 25; i++ {
		g.Escribir()
	}
	if g.veces != 1 {
		t.Fatalf("escribio %d veces el mismo estado, y con uno bastaba", g.veces)
	}
	if g.saltos != 24 {
		t.Fatalf("saltos = %d, quiero 24", g.saltos)
	}

	// Y en cuanto cambia algo de verdad, sí escribe.
	a.Poner("fotos", "nueva.txt", []byte("y"), "text/plain")
	g.Escribir()
	if g.veces != 2 {
		t.Fatalf("cambio el estado y no lo guardo (escrituras = %d)", g.veces)
	}
}

// Los avisos se agrupan: una ráfaga de peticiones es UNA escritura.
func TestPersistLosAvisosSeAgrupan(t *testing.T) {
	ruta := filepath.Join(t.TempDir(), "estado.json")
	a, _, _, _, reunir := casa()
	g := NuevaGuarda(ruta, 60*time.Millisecond, reunir, nil)
	ctx, parar := context.WithCancel(context.Background())
	g.Arrancar(ctx)

	a.CrearBucket("uno")
	for i := 0; i < 500; i++ {
		g.Cambio()
	}
	time.Sleep(200 * time.Millisecond)
	parar()
	g.Esperar()

	if g.veces == 0 {
		t.Fatal("no escribio nada")
	}
	if g.veces > 3 {
		t.Fatalf("quinientos avisos dieron %d escrituras: no se estan agrupando", g.veces)
	}
	if _, err := os.Stat(ruta); err != nil {
		t.Fatalf("el fichero no esta: %v", err)
	}
}

// Un fichero de otra versión no se carga a medias: se rechaza entero.
func TestPersistOtraVersionNoSeCarga(t *testing.T) {
	ruta := filepath.Join(t.TempDir(), "estado.json")
	if err := os.WriteFile(ruta, []byte(`{"version":99,"creada":"2026-08-25T00:00:00Z"}`),
		0o644); err != nil {
		t.Fatalf("escribir: %v", err)
	}
	if _, err := Cargar(ruta); err == nil {
		t.Fatal("cargo un fichero de una version que no conoce")
	}
}

// Y un campo que no conocemos tampoco cuela: lo escribió alguien más nuevo que nosotros.
func TestPersistUnCampoDesconocidoSeRechaza(t *testing.T) {
	ruta := filepath.Join(t.TempDir(), "estado.json")
	if err := os.WriteFile(ruta,
		[]byte(`{"version":1,"creada":"2026-08-25T00:00:00Z","secretos":[]}`), 0o644); err != nil {
		t.Fatalf("escribir: %v", err)
	}
	_, err := Cargar(ruta)
	if err == nil {
		t.Fatal("cargo un fichero con un campo que no conoce")
	}
	if !strings.Contains(err.Error(), "secretos") {
		t.Fatalf("el error no dice cual es el campo raro: %v", err)
	}
}

// Que el fichero no exista no es un error: es la primera vez.
func TestPersistSinFicheroNoEsError(t *testing.T) {
	inst, err := Cargar(filepath.Join(t.TempDir(), "no-existe.json"))
	if err != nil {
		t.Fatalf("un fichero que no existe dio error: %v", err)
	}
	if inst != nil {
		t.Fatal("devolvio una instantanea de la nada")
	}
}

// 🥇 EL DETALLE DE GO QUE MUERDE: un time.Time que vuelve de JSON representa el MISMO instante,
// pero `==` dice que no. El valor original lleva pegado un reloj monótono que el texto no puede
// llevar. Por eso las comparaciones de structs con fechas fallan después de un ida y vuelta.
func TestPersistUnTimeQueVuelveDeJSONNoEsIgualCon(t *testing.T) {
	original := time.Now() // sin .UTC(): así lleva pegado el reloj monótono
	crudo, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var vuelta time.Time
	if err := json.Unmarshal(crudo, &vuelta); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	// Es el mismo instante. De eso no hay duda.
	if !vuelta.Equal(original) {
		t.Fatalf("el instante cambio: %v vs %v", vuelta, original)
	}
	// Y aun así, `==` dice que no. Por eso `reflect.DeepEqual` sobre una struct con fechas
	// falla después de un ida y vuelta, y por eso las fechas se comparan con `Equal`.
	if vuelta == original {
		t.Fatal("esperaba que == fuera falso: el original lleva reloj monotono y la vuelta no")
	}
	t.Logf("Equal dice que si, == dice que no. Y los dos tienen razon.")
}

func TestPersistElRelojMonotonoSePierde(t *testing.T) {
	// time.Now() SÍ lleva reloj monótono. Formatear a texto lo tira.
	conMonotono := time.Now()
	sinMonotono, err := time.Parse(time.RFC3339Nano, conMonotono.Format(time.RFC3339Nano))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if conMonotono == sinMonotono {
		t.Fatal("esperaba que == fuera falso: el original lleva reloj monotono y la vuelta no")
	}
	if !conMonotono.Round(0).Equal(sinMonotono) && conMonotono.Sub(sinMonotono) > time.Microsecond {
		t.Fatalf("y aun asi tendrian que ser el mismo instante: %v vs %v",
			conMonotono, sinMonotono)
	}
}

// Escribir es atómico: no queda un fichero a medias, y el temporal no se queda por ahí.
func TestPersistEscribirNoDejaBasura(t *testing.T) {
	dir := t.TempDir()
	ruta := filepath.Join(dir, "estado.json")
	a, _, _, _, reunir := casa()
	a.CrearBucket("uno")
	g := NuevaGuarda(ruta, time.Millisecond, reunir, nil)
	g.Escribir()
	a.CrearBucket("dos")
	g.Escribir()

	entradas, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	if len(entradas) != 1 || entradas[0].Name() != "estado.json" {
		nombres := []string{}
		for _, e := range entradas {
			nombres = append(nombres, e.Name())
		}
		t.Fatalf("en el directorio hay %v, y solo deberia estar el estado", nombres)
	}
}

// Las funciones de Lambda vuelven, pero sus canales son nuevos: se puede invocar tras cargar.
func TestPersistLasFuncionesVuelvenUsables(t *testing.T) {
	ruta := filepath.Join(t.TempDir(), "estado.json")
	_, _, _, f1, reunir1 := casa()
	if _, err := f1.Crear("saludar", "provided.al2023", "bootstrap", 128, 5); err != nil {
		t.Fatalf("Crear: %v", err)
	}
	NuevaGuarda(ruta, time.Millisecond, reunir1, nil).Escribir()

	inst, _ := Cargar(ruta)
	a2, ts2, cs2, f2, _ := casa()
	if err := Aplicar(inst, a2, ts2, cs2, f2); err != nil {
		t.Fatalf("Aplicar: %v", err)
	}
	fn, err := f2.Buscar("saludar")
	if err != nil {
		t.Fatalf("la funcion no volvio: %v", err)
	}
	if fn.Segundos != 5 {
		t.Fatalf("el plazo volvio como %d", fn.Segundos)
	}
	if fn.espera == nil {
		t.Fatal("la funcion volvio sin canal: nadie podria invocarla")
	}

	// Y de verdad se puede invocar: un runtime de mentira recoge y contesta.
	go func() {
		inv, err := f2.Pedir(context.Background(), fn)
		if err == nil {
			f2.Contestar(inv.ID, Resultado{Cuerpo: []byte(`"vivo"`)})
		}
	}()
	res, err := f2.Invocar(context.Background(), fn, []byte(`{}`))
	if err != nil {
		t.Fatalf("Invocar tras cargar: %v", err)
	}
	if string(res.Cuerpo) != `"vivo"` {
		t.Fatalf("respuesta = %s", res.Cuerpo)
	}
}
