package main

// Igual que en S3: estas pruebas NO usan un cliente hecho a medida. Usan
// `github.com/aws/aws-sdk-go-v2/service/dynamodb` y `feature/dynamodb/attributevalue`, que es lo
// que usarias contra Amazon. Si el SDK oficial lo entiende, es compatible.
//
//	go test ./... -count=1 -v

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// Pedido es una struct de Go normal. El paquete `attributevalue` la convierte al mapa con tipos
// que entiende DynamoDB, y al reves. Es el puente que evita escribir {"S": ...} a mano.
type Pedido struct {
	Cliente string  `dynamodbav:"cliente"` // clave de particion
	Fecha   string  `dynamodbav:"fecha"`   // clave de orden, en ISO-8601 A PROPOSITO
	Total   float64 `dynamodbav:"total"`
	Estado  string  `dynamodbav:"estado"`
}

func nubeDynamo(t *testing.T) *dynamodb.Client {
	t.Helper()
	mux := http.NewServeMux()
	NuevoS3(NuevoAlmacen(), nil).Rutas(mux)
	srv := httptest.NewServer(Raiz(mux, map[string]http.Handler{
		"DynamoDB_": NuevoDynamo(NuevasTablas(), nil),
	}, nil, nil))
	t.Cleanup(srv.Close)

	return clienteDynamo(srv.URL)
}

// clienteDynamo esta separado porque lo reutiliza la prueba de que los tres servicios conviven
// en el mismo puerto (`sqs_test.go`).
func clienteDynamo(url string) *dynamodb.Client {
	return dynamodb.New(dynamodb.Options{
		Region:       "us-east-1",
		Credentials:  credentials.NewStaticCredentialsProvider("prueba", "prueba", ""),
		BaseEndpoint: aws.String(url),
	})
}

// tablaPedidos crea la tabla con clave compuesta: particion `cliente`, orden `fecha`.
func tablaPedidos(t *testing.T, c *dynamodb.Client) {
	t.Helper()
	ctx := context.Background()
	_, err := c.CreateTable(ctx, &dynamodb.CreateTableInput{
		TableName: aws.String("pedidos"),
		AttributeDefinitions: []types.AttributeDefinition{
			{AttributeName: aws.String("cliente"), AttributeType: types.ScalarAttributeTypeS},
			{AttributeName: aws.String("fecha"), AttributeType: types.ScalarAttributeTypeS},
		},
		KeySchema: []types.KeySchemaElement{
			{AttributeName: aws.String("cliente"), KeyType: types.KeyTypeHash},
			{AttributeName: aws.String("fecha"), KeyType: types.KeyTypeRange},
		},
		BillingMode: types.BillingModePayPerRequest,
	})
	if err != nil {
		t.Fatalf("CreateTable: %v", err)
	}
}

func guardar(t *testing.T, c *dynamodb.Client, p Pedido) {
	t.Helper()
	elem, err := attributevalue.MarshalMap(p)
	if err != nil {
		t.Fatalf("MarshalMap: %v", err)
	}
	if _, err := c.PutItem(context.Background(), &dynamodb.PutItemInput{
		TableName: aws.String("pedidos"), Item: elem,
	}); err != nil {
		t.Fatalf("PutItem: %v", err)
	}
}

func TestTablaCicloCompleto(t *testing.T) {
	ctx := context.Background()
	c := nubeDynamo(t)
	tablaPedidos(t, c)

	lista, err := c.ListTables(ctx, &dynamodb.ListTablesInput{})
	if err != nil {
		t.Fatalf("ListTables: %v", err)
	}
	if len(lista.TableNames) != 1 || lista.TableNames[0] != "pedidos" {
		t.Fatalf("ListTables = %v", lista.TableNames)
	}

	ficha, err := c.DescribeTable(ctx, &dynamodb.DescribeTableInput{TableName: aws.String("pedidos")})
	if err != nil {
		t.Fatalf("DescribeTable: %v", err)
	}
	if ficha.Table.TableStatus != types.TableStatusActive {
		t.Errorf("TableStatus = %v, quiero ACTIVE", ficha.Table.TableStatus)
	}
	if len(ficha.Table.KeySchema) != 2 {
		t.Errorf("KeySchema tiene %d elementos, quiero 2", len(ficha.Table.KeySchema))
	}

	if _, err := c.DeleteTable(ctx, &dynamodb.DeleteTableInput{
		TableName: aws.String("pedidos"),
	}); err != nil {
		t.Fatalf("DeleteTable: %v", err)
	}
	lista, _ = c.ListTables(ctx, &dynamodb.ListTablesInput{})
	if len(lista.TableNames) != 0 {
		t.Errorf("quedaron tablas: %v", lista.TableNames)
	}
}

func TestGuardarYLeerConStructDeGo(t *testing.T) {
	ctx := context.Background()
	c := nubeDynamo(t)
	tablaPedidos(t, c)
	guardar(t, c, Pedido{Cliente: "ana", Fecha: "2026-01-15", Total: 42.5, Estado: "pagado"})

	res, err := c.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String("pedidos"),
		Key: map[string]types.AttributeValue{
			"cliente": &types.AttributeValueMemberS{Value: "ana"},
			"fecha":   &types.AttributeValueMemberS{Value: "2026-01-15"},
		},
	})
	if err != nil {
		t.Fatalf("GetItem: %v", err)
	}
	var vuelve Pedido
	if err := attributevalue.UnmarshalMap(res.Item, &vuelve); err != nil {
		t.Fatalf("UnmarshalMap: %v", err)
	}
	if vuelve.Total != 42.5 || vuelve.Estado != "pagado" {
		t.Errorf("volvio %+v", vuelve)
	}
}

// Que no exista NO es un error en DynamoDB: es un 200 sin Item. Si lo tratas como error, tu
// codigo confunde "no hay" con "se rompio".
func TestLoQueNoExisteNoEsUnError(t *testing.T) {
	ctx := context.Background()
	c := nubeDynamo(t)
	tablaPedidos(t, c)

	res, err := c.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String("pedidos"),
		Key: map[string]types.AttributeValue{
			"cliente": &types.AttributeValueMemberS{Value: "nadie"},
			"fecha":   &types.AttributeValueMemberS{Value: "2026-01-01"},
		},
	})
	if err != nil {
		t.Fatalf("GetItem devolvio error y no deberia: %v", err)
	}
	if res.Item != nil {
		t.Errorf("devolvio un elemento que no existe: %v", res.Item)
	}
}

func TestTablaQueNoExisteLlegaComoResourceNotFound(t *testing.T) {
	c := nubeDynamo(t)
	_, err := c.Scan(context.Background(), &dynamodb.ScanInput{TableName: aws.String("fantasma")})
	var noEsta *types.ResourceNotFoundException
	if !errors.As(err, &noEsta) {
		t.Fatalf("el SDK no lo vio como ResourceNotFoundException: %v", err)
	}
}

func TestCrearDosVecesLaMismaTablaLlegaComoResourceInUse(t *testing.T) {
	c := nubeDynamo(t)
	tablaPedidos(t, c)
	_, err := c.CreateTable(context.Background(), &dynamodb.CreateTableInput{
		TableName: aws.String("pedidos"),
		AttributeDefinitions: []types.AttributeDefinition{
			{AttributeName: aws.String("cliente"), AttributeType: types.ScalarAttributeTypeS},
		},
		KeySchema: []types.KeySchemaElement{
			{AttributeName: aws.String("cliente"), KeyType: types.KeyTypeHash},
		},
	})
	var enUso *types.ResourceInUseException
	if !errors.As(err, &enUso) {
		t.Fatalf("el SDK no lo vio como ResourceInUseException: %v", err)
	}
}

// 🥇 La prueba que enseña como piensa DynamoDB: Query SOLO sabe de la clave.
func TestQuerySoloMiraLaClave(t *testing.T) {
	ctx := context.Background()
	c := nubeDynamo(t)
	tablaPedidos(t, c)
	guardar(t, c, Pedido{Cliente: "ana", Fecha: "2026-01-15", Total: 10, Estado: "pagado"})
	guardar(t, c, Pedido{Cliente: "ana", Fecha: "2026-02-03", Total: 20, Estado: "enviado"})
	guardar(t, c, Pedido{Cliente: "ana", Fecha: "2026-03-21", Total: 30, Estado: "pagado"})
	guardar(t, c, Pedido{Cliente: "beto", Fecha: "2026-02-10", Total: 99, Estado: "pagado"})

	// Todos los de ana: igualdad EXACTA en la clave de particion.
	res, err := c.Query(ctx, &dynamodb.QueryInput{
		TableName:                 aws.String("pedidos"),
		KeyConditionExpression:    aws.String("cliente = :c"),
		ExpressionAttributeValues: map[string]types.AttributeValue{":c": &types.AttributeValueMemberS{Value: "ana"}},
	})
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	if res.Count != 3 {
		t.Errorf("Count = %d, quiero 3", res.Count)
	}

	// Un rango sobre la clave de ORDEN. Y funciona porque las fechas van en ISO-8601: el orden
	// alfabetico coincide con el cronologico. Con "15/01/2026" esto daria basura.
	res, err = c.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String("pedidos"),
		KeyConditionExpression: aws.String("cliente = :c AND fecha BETWEEN :a AND :b"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":c": &types.AttributeValueMemberS{Value: "ana"},
			":a": &types.AttributeValueMemberS{Value: "2026-02-01"},
			":b": &types.AttributeValueMemberS{Value: "2026-02-28"},
		},
	})
	if err != nil {
		t.Fatalf("Query BETWEEN: %v", err)
	}
	if res.Count != 1 {
		t.Fatalf("BETWEEN dio %d, quiero 1", res.Count)
	}
	var uno Pedido
	_ = attributevalue.UnmarshalMap(res.Items[0], &uno)
	if uno.Fecha != "2026-02-03" {
		t.Errorf("salio %s", uno.Fecha)
	}

	// begins_with sobre la clave de orden: "todo lo de 2026-03".
	res, err = c.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String("pedidos"),
		KeyConditionExpression: aws.String("cliente = :c AND begins_with(fecha, :p)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":c": &types.AttributeValueMemberS{Value: "ana"},
			":p": &types.AttributeValueMemberS{Value: "2026-03"},
		},
	})
	if err != nil {
		t.Fatalf("Query begins_with: %v", err)
	}
	if res.Count != 1 {
		t.Errorf("begins_with dio %d, quiero 1", res.Count)
	}
}

// ⛔ Sobre la clave de particion SOLO hay igualdad. No es una limitacion de nubecita: es la
// regla de DynamoDB, y por eso el error tiene que llegar como ValidationException.
func TestNoSePuedeBuscarPorPrefijoEnLaClaveDeParticion(t *testing.T) {
	c := nubeDynamo(t)
	tablaPedidos(t, c)
	_, err := c.Query(context.Background(), &dynamodb.QueryInput{
		TableName:              aws.String("pedidos"),
		KeyConditionExpression: aws.String("begins_with(cliente, :p)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":p": &types.AttributeValueMemberS{Value: "a"},
		},
	})
	if err == nil {
		t.Fatal("dejo buscar por prefijo en la clave de particion, y DynamoDB no deja")
	}
	// El mensaje tiene que DECIR por que. Un "ValidationException" a secas no ayuda a nadie.
	if !strings.Contains(err.Error(), "clave de particion") {
		t.Fatalf("el error no explica por que: %v", err)
	}
}

func TestQueryAlReves(t *testing.T) {
	ctx := context.Background()
	c := nubeDynamo(t)
	tablaPedidos(t, c)
	guardar(t, c, Pedido{Cliente: "ana", Fecha: "2026-01-15", Total: 10})
	guardar(t, c, Pedido{Cliente: "ana", Fecha: "2026-03-21", Total: 30})

	res, err := c.Query(ctx, &dynamodb.QueryInput{
		TableName:                 aws.String("pedidos"),
		KeyConditionExpression:    aws.String("cliente = :c"),
		ExpressionAttributeValues: map[string]types.AttributeValue{":c": &types.AttributeValueMemberS{Value: "ana"}},
		ScanIndexForward:          aws.Bool(false),
		Limit:                     aws.Int32(1),
	})
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	var uno Pedido
	_ = attributevalue.UnmarshalMap(res.Items[0], &uno)
	if uno.Fecha != "2026-03-21" {
		t.Errorf("el mas reciente deberia ser 2026-03-21 y salio %s", uno.Fecha)
	}
}

// La escritura condicional: la unica herramienta de concurrencia que da DynamoDB.
func TestEscrituraCondicional(t *testing.T) {
	ctx := context.Background()
	c := nubeDynamo(t)
	tablaPedidos(t, c)
	elem, _ := attributevalue.MarshalMap(Pedido{Cliente: "ana", Fecha: "2026-01-15", Total: 10})

	// La primera pasa: nadie lo habia escrito.
	if _, err := c.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String("pedidos"), Item: elem,
		ConditionExpression: aws.String("attribute_not_exists(cliente)"),
	}); err != nil {
		t.Fatalf("la primera escritura condicional deberia pasar: %v", err)
	}
	// La segunda no, y tiene que llegar con SU tipo de error.
	_, err := c.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String("pedidos"), Item: elem,
		ConditionExpression: aws.String("attribute_not_exists(cliente)"),
	})
	var fallo *types.ConditionalCheckFailedException
	if !errors.As(err, &fallo) {
		t.Fatalf("el SDK no lo vio como ConditionalCheckFailedException: %v", err)
	}
}

// Scan lee la tabla ENTERA. Aqui son cuatro elementos; en Amazon es la diferencia entre unos
// centavos y una factura que duele.
func TestScanLeeTodo(t *testing.T) {
	ctx := context.Background()
	c := nubeDynamo(t)
	tablaPedidos(t, c)
	guardar(t, c, Pedido{Cliente: "ana", Fecha: "2026-01-15"})
	guardar(t, c, Pedido{Cliente: "beto", Fecha: "2026-02-10"})

	res, err := c.Scan(ctx, &dynamodb.ScanInput{TableName: aws.String("pedidos")})
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if res.Count != 2 || res.ScannedCount != 2 {
		t.Errorf("Count=%d ScannedCount=%d, quiero 2 y 2", res.Count, res.ScannedCount)
	}
}

func TestBorrarElemento(t *testing.T) {
	ctx := context.Background()
	c := nubeDynamo(t)
	tablaPedidos(t, c)
	guardar(t, c, Pedido{Cliente: "ana", Fecha: "2026-01-15"})

	clave := map[string]types.AttributeValue{
		"cliente": &types.AttributeValueMemberS{Value: "ana"},
		"fecha":   &types.AttributeValueMemberS{Value: "2026-01-15"},
	}
	if _, err := c.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String("pedidos"), Key: clave,
	}); err != nil {
		t.Fatalf("DeleteItem: %v", err)
	}
	// Borrar dos veces no es un error, igual que en S3.
	if _, err := c.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String("pedidos"), Key: clave,
	}); err != nil {
		t.Fatalf("el segundo DeleteItem no deberia fallar: %v", err)
	}
	res, _ := c.Scan(ctx, &dynamodb.ScanInput{TableName: aws.String("pedidos")})
	if res.Count != 0 {
		t.Errorf("quedaron %d elementos", res.Count)
	}
}

// 🥇 Los numeros viajan como CADENA por el cable, y no es un descuido: asi no pasan por el float
// de ningun lenguaje y no pierden precision.
func TestLosNumerosViajanComoCadena(t *testing.T) {
	ctx := context.Background()
	c := nubeDynamo(t)
	tablaPedidos(t, c)
	if _, err := c.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String("pedidos"),
		Item: map[string]types.AttributeValue{
			"cliente": &types.AttributeValueMemberS{Value: "ana"},
			"fecha":   &types.AttributeValueMemberS{Value: "2026-01-15"},
			// 20 digitos: no cabe entero en un float64.
			"exacto": &types.AttributeValueMemberN{Value: "12345678901234567890"},
		},
	}); err != nil {
		t.Fatalf("PutItem: %v", err)
	}
	res, err := c.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String("pedidos"),
		Key: map[string]types.AttributeValue{
			"cliente": &types.AttributeValueMemberS{Value: "ana"},
			"fecha":   &types.AttributeValueMemberS{Value: "2026-01-15"},
		},
	})
	if err != nil {
		t.Fatalf("GetItem: %v", err)
	}
	n, ok := res.Item["exacto"].(*types.AttributeValueMemberN)
	if !ok {
		t.Fatalf("el atributo no volvio como N: %T", res.Item["exacto"])
	}
	if n.Value != "12345678901234567890" {
		t.Errorf("volvio %q: perdio precision por el camino", n.Value)
	}
}

// Clave de orden NUMERICA: el 10 va DESPUES del 9. Si ordenaras por la cadena, iria antes.
func TestLaClaveDeOrdenNumericaSeOrdenaComoNumero(t *testing.T) {
	ctx := context.Background()
	c := nubeDynamo(t)
	if _, err := c.CreateTable(ctx, &dynamodb.CreateTableInput{
		TableName: aws.String("eventos"),
		AttributeDefinitions: []types.AttributeDefinition{
			{AttributeName: aws.String("flujo"), AttributeType: types.ScalarAttributeTypeS},
			{AttributeName: aws.String("secuencia"), AttributeType: types.ScalarAttributeTypeN},
		},
		KeySchema: []types.KeySchemaElement{
			{AttributeName: aws.String("flujo"), KeyType: types.KeyTypeHash},
			{AttributeName: aws.String("secuencia"), KeyType: types.KeyTypeRange},
		},
	}); err != nil {
		t.Fatalf("CreateTable: %v", err)
	}
	for _, n := range []string{"10", "9", "2"} {
		if _, err := c.PutItem(ctx, &dynamodb.PutItemInput{
			TableName: aws.String("eventos"),
			Item: map[string]types.AttributeValue{
				"flujo":     &types.AttributeValueMemberS{Value: "a"},
				"secuencia": &types.AttributeValueMemberN{Value: n},
			},
		}); err != nil {
			t.Fatalf("PutItem %s: %v", n, err)
		}
	}
	res, err := c.Query(ctx, &dynamodb.QueryInput{
		TableName:                 aws.String("eventos"),
		KeyConditionExpression:    aws.String("flujo = :f"),
		ExpressionAttributeValues: map[string]types.AttributeValue{":f": &types.AttributeValueMemberS{Value: "a"}},
	})
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	var orden []string
	for _, e := range res.Items {
		orden = append(orden, e["secuencia"].(*types.AttributeValueMemberN).Value)
	}
	quiero := []string{"2", "9", "10"}
	for i := range quiero {
		if i >= len(orden) || orden[i] != quiero[i] {
			t.Fatalf("orden = %v, quiero %v (si ordenas por cadena sale 10, 2, 9)", orden, quiero)
		}
	}
}

// S3 y DynamoDB en el MISMO puerto: se reparten mirando la cabecera X-Amz-Target.
func TestLosDosServiciosConvivenEnElMismoPuerto(t *testing.T) {
	ctx := context.Background()
	mux := http.NewServeMux()
	NuevoS3(NuevoAlmacen(), nil).Rutas(mux)
	srv := httptest.NewServer(Raiz(mux, map[string]http.Handler{
		"DynamoDB_": NuevoDynamo(NuevasTablas(), nil),
	}, nil, nil))
	defer srv.Close()

	din := dynamodb.New(dynamodb.Options{
		Region:       "us-east-1",
		Credentials:  credentials.NewStaticCredentialsProvider("prueba", "prueba", ""),
		BaseEndpoint: aws.String(srv.URL),
	})
	if _, err := din.ListTables(ctx, &dynamodb.ListTablesInput{}); err != nil {
		t.Fatalf("DynamoDB en el puerto compartido: %v", err)
	}
	// Y S3, en el mismo sitio, sigue funcionando.
	s3c := clienteS3(srv.URL)
	if _, err := s3c.ListBuckets(ctx, nil); err != nil {
		t.Fatalf("S3 en el puerto compartido: %v", err)
	}
}
