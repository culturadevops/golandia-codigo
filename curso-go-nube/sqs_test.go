package main

// Pruebas de SQS contra `github.com/aws/aws-sdk-go-v2/service/sqs`, el cliente oficial.
//
// ⚠️ Ninguna prueba DUERME. El reloj de las colas es inyectable, así que "que pasen treinta
// segundos" es una asignación. Una prueba que duerme para comprobar un tiempo de espera es
// lenta y, cuando la máquina va cargada, además es inestable.

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

// relojFalso permite mover el tiempo a mano.
type relojFalso struct{ ahora time.Time }

func (r *relojFalso) leer() time.Time         { return r.ahora }
func (r *relojFalso) avanzar(d time.Duration) { r.ahora = r.ahora.Add(d) }

func nubeSQS(t *testing.T) (*sqs.Client, *Colas, *relojFalso) {
	t.Helper()
	reloj := &relojFalso{ahora: time.Date(2026, 8, 25, 12, 0, 0, 0, time.UTC)}
	mux := http.NewServeMux()
	NuevoS3(NuevoAlmacen(), nil).Rutas(mux)
	colas := NuevasColas()
	colas.reloj = reloj.leer
	srv := httptest.NewServer(Raiz(mux, map[string]http.Handler{
		"AmazonSQS.": NuevoSQS(colas, nil),
	}, nil, nil))
	t.Cleanup(srv.Close)

	c := sqs.New(sqs.Options{
		Region:       "us-east-1",
		Credentials:  credentials.NewStaticCredentialsProvider("prueba", "prueba", ""),
		BaseEndpoint: aws.String(srv.URL),
	})
	return c, colas, reloj
}

func crearCola(t *testing.T, c *sqs.Client, nombre string, attr map[string]string) string {
	t.Helper()
	res, err := c.CreateQueue(context.Background(), &sqs.CreateQueueInput{
		QueueName: aws.String(nombre), Attributes: attr,
	})
	if err != nil {
		t.Fatalf("CreateQueue %s: %v", nombre, err)
	}
	return aws.ToString(res.QueueUrl)
}

func TestColaCicloCompleto(t *testing.T) {
	ctx := context.Background()
	c, _, _ := nubeSQS(t)
	url := crearCola(t, c, "pedidos", nil)

	lista, err := c.ListQueues(ctx, &sqs.ListQueuesInput{})
	if err != nil {
		t.Fatalf("ListQueues: %v", err)
	}
	if len(lista.QueueUrls) != 1 {
		t.Fatalf("ListQueues = %v", lista.QueueUrls)
	}

	env, err := c.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl: aws.String(url), MessageBody: aws.String("procesa el pedido 7"),
	})
	if err != nil {
		t.Fatalf("SendMessage: %v", err)
	}
	if aws.ToString(env.MessageId) == "" {
		t.Error("no devolvio MessageId")
	}

	rec, err := c.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{QueueUrl: aws.String(url)})
	if err != nil {
		t.Fatalf("ReceiveMessage: %v", err)
	}
	if len(rec.Messages) != 1 {
		t.Fatalf("recibi %d mensajes, quiero 1", len(rec.Messages))
	}
	if aws.ToString(rec.Messages[0].Body) != "procesa el pedido 7" {
		t.Errorf("cuerpo = %q", aws.ToString(rec.Messages[0].Body))
	}

	if _, err := c.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		QueueUrl: aws.String(url), ReceiptHandle: rec.Messages[0].ReceiptHandle,
	}); err != nil {
		t.Fatalf("DeleteMessage: %v", err)
	}
	if _, err := c.DeleteQueue(ctx, &sqs.DeleteQueueInput{QueueUrl: aws.String(url)}); err != nil {
		t.Fatalf("DeleteQueue: %v", err)
	}
}

// 🥇 LA prueba del episodio: recibir NO saca el mensaje de la cola. Lo esconde.
func TestRecibirNoSacaElMensajeLoEsconde(t *testing.T) {
	ctx := context.Background()
	c, _, reloj := nubeSQS(t)
	url := crearCola(t, c, "trabajos", map[string]string{"VisibilityTimeout": "30"})
	if _, err := c.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl: aws.String(url), MessageBody: aws.String("uno"),
	}); err != nil {
		t.Fatalf("SendMessage: %v", err)
	}

	// Primera recepción: llega.
	primera, _ := c.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{QueueUrl: aws.String(url)})
	if len(primera.Messages) != 1 {
		t.Fatalf("la primera recepcion dio %d mensajes", len(primera.Messages))
	}

	// Inmediatamente después: NO llega. Está invisible, pero sigue en la cola.
	vacia, _ := c.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{QueueUrl: aws.String(url)})
	if len(vacia.Messages) != 0 {
		t.Fatalf("lo entrego dos veces seguidas: la visibilidad no funciona")
	}

	// Pasan 31 segundos sin borrarlo: vuelve. Nadie lo perdio.
	reloj.avanzar(31 * time.Second)
	otra, _ := c.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{QueueUrl: aws.String(url)})
	if len(otra.Messages) != 1 {
		t.Fatalf("no volvio despues de la visibilidad: %d mensajes", len(otra.Messages))
	}
	// Y el contador de entregas va por dos: es lo que delata un consumidor que no borra.
	if otra.Messages[0].Attributes["ApproximateReceiveCount"] != "2" {
		t.Errorf("ApproximateReceiveCount = %q, quiero 2",
			otra.Messages[0].Attributes["ApproximateReceiveCount"])
	}
}

// El recibo cambia en CADA entrega, y el viejo deja de valer. Es el fallo clásico del
// consumidor lento: borra con un recibo caducado, cree que borró, y el mensaje sigue ahí.
func TestElReciboCambiaEnCadaEntrega(t *testing.T) {
	ctx := context.Background()
	c, _, reloj := nubeSQS(t)
	url := crearCola(t, c, "trabajos", map[string]string{"VisibilityTimeout": "10"})
	_, _ = c.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl: aws.String(url), MessageBody: aws.String("uno"),
	})

	primera, _ := c.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{QueueUrl: aws.String(url)})
	reciboViejo := primera.Messages[0].ReceiptHandle

	reloj.avanzar(11 * time.Second)
	segunda, _ := c.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{QueueUrl: aws.String(url)})
	if aws.ToString(segunda.Messages[0].ReceiptHandle) == aws.ToString(reciboViejo) {
		t.Fatal("el recibo no cambio entre entregas")
	}

	// Borrar con el recibo viejo: SQS no falla, pero NO borra.
	if _, err := c.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		QueueUrl: aws.String(url), ReceiptHandle: reciboViejo,
	}); err != nil {
		t.Fatalf("borrar con recibo caducado no deberia dar error: %v", err)
	}
	reloj.avanzar(11 * time.Second)
	sigue, _ := c.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{QueueUrl: aws.String(url)})
	if len(sigue.Messages) != 1 {
		t.Fatal("el recibo caducado SI borro, y no debia")
	}
}

// 🥇 La cola de mensajes muertos: a la N-ésima entrega, en vez de darlo otra vez, se aparta.
func TestLaColaDeMuertosSeLlevaElQueSiempreFalla(t *testing.T) {
	ctx := context.Background()
	c, _, reloj := nubeSQS(t)
	crearCola(t, c, "muertos", nil)
	url := crearCola(t, c, "trabajos", map[string]string{
		"VisibilityTimeout": "5",
		"RedrivePolicy": `{"deadLetterTargetArn":"arn:aws:sqs:us-east-1:000000000000:muertos",` +
			`"maxReceiveCount":2}`,
	})
	_, _ = c.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl: aws.String(url), MessageBody: aws.String("el que siempre revienta"),
	})

	// Dos entregas: el consumidor lo coge y no lo borra (porque falla).
	for i := 1; i <= 2; i++ {
		r, _ := c.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{QueueUrl: aws.String(url)})
		if len(r.Messages) != 1 {
			t.Fatalf("entrega %d: recibi %d mensajes", i, len(r.Messages))
		}
		reloj.avanzar(6 * time.Second)
	}
	// La tercera vez ya no se entrega: se aparta.
	tercera, _ := c.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{QueueUrl: aws.String(url)})
	if len(tercera.Messages) != 0 {
		t.Fatalf("lo entrego una tercera vez: el redrive no funciona")
	}

	// Y está en la cola de muertos, entero.
	muertosURL, err := c.GetQueueUrl(ctx, &sqs.GetQueueUrlInput{QueueName: aws.String("muertos")})
	if err != nil {
		t.Fatalf("GetQueueUrl: %v", err)
	}
	enMuertos, _ := c.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{QueueUrl: muertosURL.QueueUrl})
	if len(enMuertos.Messages) != 1 {
		t.Fatalf("la cola de muertos tiene %d mensajes, quiero 1", len(enMuertos.Messages))
	}
	if aws.ToString(enMuertos.Messages[0].Body) != "el que siempre revienta" {
		t.Errorf("llego con otro cuerpo: %q", aws.ToString(enMuertos.Messages[0].Body))
	}
}

// Pedir a una cola vacía NO es un error: es que no hay trabajo.
func TestColaVaciaNoEsUnError(t *testing.T) {
	ctx := context.Background()
	c, _, _ := nubeSQS(t)
	url := crearCola(t, c, "vacia", nil)
	res, err := c.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{QueueUrl: aws.String(url)})
	if err != nil {
		t.Fatalf("recibir de una cola vacia dio error: %v", err)
	}
	if len(res.Messages) != 0 {
		t.Errorf("devolvio %d mensajes de una cola vacia", len(res.Messages))
	}
}

func TestColaQueNoExisteLlegaComoQueueDoesNotExist(t *testing.T) {
	c, _, _ := nubeSQS(t)
	_, err := c.GetQueueUrl(context.Background(), &sqs.GetQueueUrlInput{
		QueueName: aws.String("fantasma"),
	})
	var noEsta *types.QueueDoesNotExist
	if !errors.As(err, &noEsta) {
		t.Fatalf("el SDK no lo vio como QueueDoesNotExist: %v", err)
	}
}

// Los dos números que SQS llama "aproximados", y por qué lo son: dependen del instante.
func TestLosNumerosAproximadosDeLaCola(t *testing.T) {
	ctx := context.Background()
	c, _, reloj := nubeSQS(t)
	url := crearCola(t, c, "cuentas", map[string]string{"VisibilityTimeout": "20"})
	for _, cuerpo := range []string{"a", "b", "c"} {
		_, _ = c.SendMessage(ctx, &sqs.SendMessageInput{
			QueueUrl: aws.String(url), MessageBody: aws.String(cuerpo),
		})
	}
	_, _ = c.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
		QueueUrl: aws.String(url), MaxNumberOfMessages: 2,
	})

	res, err := c.GetQueueAttributes(ctx, &sqs.GetQueueAttributesInput{
		QueueUrl:       aws.String(url),
		AttributeNames: []types.QueueAttributeName{types.QueueAttributeNameAll},
	})
	if err != nil {
		t.Fatalf("GetQueueAttributes: %v", err)
	}
	if res.Attributes["ApproximateNumberOfMessages"] != "1" {
		t.Errorf("visibles = %q, quiero 1", res.Attributes["ApproximateNumberOfMessages"])
	}
	if res.Attributes["ApproximateNumberOfMessagesNotVisible"] != "2" {
		t.Errorf("en vuelo = %q, quiero 2",
			res.Attributes["ApproximateNumberOfMessagesNotVisible"])
	}
	// Y en cuanto pasa la visibilidad, los mismos tres vuelven a estar visibles. De ahi el
	// "aproximado": la respuesta depende del instante en que preguntas.
	reloj.avanzar(21 * time.Second)
	res, _ = c.GetQueueAttributes(ctx, &sqs.GetQueueAttributesInput{
		QueueUrl:       aws.String(url),
		AttributeNames: []types.QueueAttributeName{types.QueueAttributeNameAll},
	})
	if res.Attributes["ApproximateNumberOfMessages"] != "3" {
		t.Errorf("tras la visibilidad, visibles = %q, quiero 3",
			res.Attributes["ApproximateNumberOfMessages"])
	}
}

// Un consumidor que va a tardar más pide prórroga; y con 0 lo devuelve a la cola al instante.
func TestCambiarLaVisibilidad(t *testing.T) {
	ctx := context.Background()
	c, _, _ := nubeSQS(t)
	url := crearCola(t, c, "trabajos", map[string]string{"VisibilityTimeout": "300"})
	_, _ = c.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl: aws.String(url), MessageBody: aws.String("uno"),
	})
	r, _ := c.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{QueueUrl: aws.String(url)})

	// "No puedo con esto, que lo coja otro YA."
	if _, err := c.ChangeMessageVisibility(ctx, &sqs.ChangeMessageVisibilityInput{
		QueueUrl: aws.String(url), ReceiptHandle: r.Messages[0].ReceiptHandle,
		VisibilityTimeout: 0,
	}); err != nil {
		t.Fatalf("ChangeMessageVisibility: %v", err)
	}
	otra, _ := c.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{QueueUrl: aws.String(url)})
	if len(otra.Messages) != 1 {
		t.Fatal("con visibilidad 0 el mensaje deberia volver de inmediato")
	}
}

func TestEnviarYBorrarEnLote(t *testing.T) {
	ctx := context.Background()
	c, _, _ := nubeSQS(t)
	url := crearCola(t, c, "lotes", nil)
	res, err := c.SendMessageBatch(ctx, &sqs.SendMessageBatchInput{
		QueueUrl: aws.String(url),
		Entries: []types.SendMessageBatchRequestEntry{
			{Id: aws.String("1"), MessageBody: aws.String("uno")},
			{Id: aws.String("2"), MessageBody: aws.String("dos")},
			{Id: aws.String("3"), MessageBody: aws.String("tres")},
		},
	})
	if err != nil {
		t.Fatalf("SendMessageBatch: %v", err)
	}
	if len(res.Successful) != 3 {
		t.Fatalf("enviados %d de 3", len(res.Successful))
	}

	rec, _ := c.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
		QueueUrl: aws.String(url), MaxNumberOfMessages: 10,
	})
	if len(rec.Messages) != 3 {
		t.Fatalf("recibi %d de 3", len(rec.Messages))
	}
	var borrados []types.DeleteMessageBatchRequestEntry
	for i, m := range rec.Messages {
		borrados = append(borrados, types.DeleteMessageBatchRequestEntry{
			Id: aws.String(string(rune('a' + i))), ReceiptHandle: m.ReceiptHandle,
		})
	}
	if _, err := c.DeleteMessageBatch(ctx, &sqs.DeleteMessageBatchInput{
		QueueUrl: aws.String(url), Entries: borrados,
	}); err != nil {
		t.Fatalf("DeleteMessageBatch: %v", err)
	}
	attr, _ := c.GetQueueAttributes(ctx, &sqs.GetQueueAttributesInput{
		QueueUrl:       aws.String(url),
		AttributeNames: []types.QueueAttributeName{types.QueueAttributeNameAll},
	})
	if attr.Attributes["ApproximateNumberOfMessages"] != "0" ||
		attr.Attributes["ApproximateNumberOfMessagesNotVisible"] != "0" {
		t.Errorf("la cola no quedo vacia: %v", attr.Attributes)
	}
}

// Los tres servicios, en el MISMO puerto.
func TestLosTresServiciosConvivenEnElMismoPuerto(t *testing.T) {
	ctx := context.Background()
	mux := http.NewServeMux()
	NuevoS3(NuevoAlmacen(), nil).Rutas(mux)
	srv := httptest.NewServer(Raiz(mux, map[string]http.Handler{
		"DynamoDB_":  NuevoDynamo(NuevasTablas(), nil),
		"AmazonSQS.": NuevoSQS(NuevasColas(), nil),
	}, nil, nil))
	defer srv.Close()

	if _, err := clienteS3(srv.URL).ListBuckets(ctx, nil); err != nil {
		t.Fatalf("S3: %v", err)
	}
	din := clienteDynamo(srv.URL)
	if _, err := din.ListTables(ctx, nil); err != nil {
		t.Fatalf("DynamoDB: %v", err)
	}
	col := sqs.New(sqs.Options{
		Region:       "us-east-1",
		Credentials:  credentials.NewStaticCredentialsProvider("prueba", "prueba", ""),
		BaseEndpoint: aws.String(srv.URL),
	})
	if _, err := col.ListQueues(ctx, &sqs.ListQueuesInput{}); err != nil {
		t.Fatalf("SQS: %v", err)
	}
}
