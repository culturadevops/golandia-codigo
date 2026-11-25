package main

// Las pruebas de los disparadores. Quien crea las colas, sube los objetos y monta los mapeos es
// el SDK oficial de AWS; lo que hace de función es un runtime casero, como en el episodio 5.
//
//	go test ./... -count=1 -race -run Disparador

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	lambdasdk "github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	sqstypes "github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

// nubeConDisparadores monta la nubecita entera con los disparadores enchufados.
func nubeConDisparadores(t *testing.T, sondeo time.Duration) (
	*lambdasdk.Client, *s3.Client, *sqs.Client, *Funciones, *Disparadores) {
	t.Helper()
	almacen, tablas, colas, fns := NuevoAlmacen(), NuevasTablas(), NuevasColas(), NuevasFunciones()
	disp := NuevosDisparadores(fns, nil, colas, sondeo, nil)
	t.Cleanup(disp.Apagar)

	mux := http.NewServeMux()
	s3srv := NuevoS3(almacen, nil)
	s3srv.ConNotificador(NuevoNotificador(disp, colas, nil))
	s3srv.Rutas(mux)

	lam := NuevoLambda(fns, nil).ConDisparadores(disp)
	rutas := lam.Rutas()
	porRuta := map[string]http.Handler{}
	for _, p := range lam.Prefijos() {
		porRuta[p] = rutas
	}
	srv := httptest.NewServer(Raiz(mux, map[string]http.Handler{
		"DynamoDB_":  NuevoDynamo(tablas, nil),
		"AmazonSQS.": NuevoSQS(colas, nil),
	}, porRuta, nil))
	t.Cleanup(srv.Close)

	return lambdasdk.New(lambdasdk.Options{
			Region: "us-east-1", Credentials: credencialesFijas("prueba", "prueba"),
			BaseEndpoint: aws.String(srv.URL)}),
		clienteS3(srv.URL), clienteSQS(srv.URL), fns, disp
}

// runtimeQueAtiende deja a una función contestando en bucle con lo que le digas, y devuelve los
// eventos que le fueron llegando.
type espia struct {
	mu       sync.Mutex
	eventos  [][]byte
	responde func(evento []byte) Resultado
	parar    func()
}

func atender(t *testing.T, fns *Funciones, nombre string,
	responde func([]byte) Resultado) *espia {
	t.Helper()
	fn, err := fns.Buscar(nombre)
	if err != nil {
		t.Fatalf("Buscar %s: %v", nombre, err)
	}
	e := &espia{responde: responde}
	ctx, cancelar := context.WithCancel(context.Background())
	fin := make(chan struct{})
	go func() {
		defer close(fin)
		for {
			inv, err := fns.Pedir(ctx, fn)
			if err != nil {
				return
			}
			e.mu.Lock()
			e.eventos = append(e.eventos, inv.Evento)
			e.mu.Unlock()
			fns.Contestar(inv.ID, e.responde(inv.Evento))
		}
	}()
	e.parar = func() { cancelar(); <-fin }
	t.Cleanup(e.parar)
	return e
}

func (e *espia) cuantos() int {
	e.mu.Lock()
	defer e.mu.Unlock()
	return len(e.eventos)
}

func (e *espia) ultimo() []byte {
	e.mu.Lock()
	defer e.mu.Unlock()
	if len(e.eventos) == 0 {
		return nil
	}
	return e.eventos[len(e.eventos)-1]
}

// hasta espera a que se cumpla algo, sin dormir a ciegas.
func hasta(t *testing.T, cuanto time.Duration, que string, cond func() bool) {
	t.Helper()
	limite := time.Now().Add(cuanto)
	for time.Now().Before(limite) {
		if cond() {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("se agotó el tiempo esperando a que %s", que)
}

func crearFn(t *testing.T, c *lambdasdk.Client, nombre string, segundos int32) {
	t.Helper()
	crearFuncion(t, c, nombre, segundos)
}

func crearColaSimple(t *testing.T, q *sqs.Client, nombre string) string {
	t.Helper()
	r, err := q.CreateQueue(context.Background(), &sqs.CreateQueueInput{
		QueueName: aws.String(nombre)})
	if err != nil {
		t.Fatalf("CreateQueue: %v", err)
	}
	return aws.ToString(r.QueueUrl)
}

// 🥇 La prueba de fondo: mando un mensaje a una cola y la función se ejecuta sola.
func TestDisparadorSQSInvocaSolo(t *testing.T) {
	lam, _, q, fns, _ := nubeConDisparadores(t, 20*time.Millisecond)
	crearFn(t, lam, "procesar", 5)
	url := crearColaSimple(t, q, "trabajos")
	e := atender(t, fns, "procesar", func([]byte) Resultado {
		return Resultado{Cuerpo: []byte(`"hecho"`)}
	})

	if _, err := lam.CreateEventSourceMapping(context.Background(),
		&lambdasdk.CreateEventSourceMappingInput{
			EventSourceArn: aws.String("arn:aws:sqs:us-east-1:000000000000:trabajos"),
			FunctionName:   aws.String("procesar"),
			BatchSize:      aws.Int32(10),
		}); err != nil {
		t.Fatalf("CreateEventSourceMapping: %v", err)
	}

	if _, err := q.SendMessage(context.Background(), &sqs.SendMessageInput{
		QueueUrl: aws.String(url), MessageBody: aws.String("cobrar 10 euros")}); err != nil {
		t.Fatalf("SendMessage: %v", err)
	}

	hasta(t, 3*time.Second, "la funcion reciba el mensaje", func() bool { return e.cuantos() > 0 })

	// Y el evento tiene el formato de verdad de SQS.
	var ev struct {
		Records []struct {
			MessageID     string            `json:"messageId"`
			ReceiptHandle string            `json:"receiptHandle"`
			Body          string            `json:"body"`
			EventSource   string            `json:"eventSource"`
			EventSourceAR string            `json:"eventSourceARN"`
			Attributes    map[string]string `json:"attributes"`
		} `json:"Records"`
	}
	if err := json.Unmarshal(e.ultimo(), &ev); err != nil {
		t.Fatalf("el evento no es JSON: %v\n%s", err, e.ultimo())
	}
	if len(ev.Records) != 1 {
		t.Fatalf("llegaron %d registros", len(ev.Records))
	}
	r := ev.Records[0]
	if r.Body != "cobrar 10 euros" {
		t.Fatalf("body = %q", r.Body)
	}
	if r.EventSource != "aws:sqs" {
		t.Fatalf("eventSource = %q", r.EventSource)
	}
	if !strings.HasSuffix(r.EventSourceAR, ":trabajos") {
		t.Fatalf("eventSourceARN = %q", r.EventSourceAR)
	}
	if r.Attributes["ApproximateReceiveCount"] != "1" {
		t.Fatalf("la cuenta de entregas llego como %q", r.Attributes["ApproximateReceiveCount"])
	}
}

// 🥇 Y LA REGLA DEL EPISODIO: la función NO borra el mensaje. Lo borra el disparador, y solo si
// fue bien. Aquí va bien, así que la cola se queda vacía.
func TestDisparadorBorraElMensajeSiLaFuncionVaBien(t *testing.T) {
	lam, _, q, fns, _ := nubeConDisparadores(t, 20*time.Millisecond)
	crearFn(t, lam, "procesar", 5)
	url := crearColaSimple(t, q, "trabajos")
	atender(t, fns, "procesar", func([]byte) Resultado {
		return Resultado{Cuerpo: []byte(`"hecho"`)}
	})
	mapear(t, lam, "trabajos", "procesar", 10)

	enviar(t, q, url, "uno")
	esperarVacia(t, q, url, 3*time.Second)
}

// 🥇 Y si la función falla, NO se borra. El mensaje vuelve — sin que nadie escriba reintentos.
func TestDisparadorSiLaFuncionFallaElMensajeVuelve(t *testing.T) {
	lam, _, q, fns, _ := nubeConDisparadores(t, 20*time.Millisecond)
	crearFn(t, lam, "procesar", 5)
	url := crearColaSimple(t, q, "trabajos")
	e := atender(t, fns, "procesar", func([]byte) Resultado {
		return Resultado{Cuerpo: []byte(`{"errorMessage":"no puedo"}`), Fallo: true}
	})
	mapear(t, lam, "trabajos", "procesar", 10)

	enviar(t, q, url, "el que falla")

	// Se entrega varias veces: la visibilidad es de 30 s, pero como el disparador recibe con
	// el valor por defecto de la cola, hay que mirar que al menos llegó una vez y que NO se
	// borró.
	hasta(t, 3*time.Second, "la funcion lo reciba", func() bool { return e.cuantos() > 0 })
	time.Sleep(150 * time.Millisecond)

	attrs, err := q.GetQueueAttributes(context.Background(), &sqs.GetQueueAttributesInput{
		QueueUrl: aws.String(url), AttributeNames: []sqstypesNombre{
			"ApproximateNumberOfMessages", "ApproximateNumberOfMessagesNotVisible"}})
	if err != nil {
		t.Fatalf("GetQueueAttributes: %v", err)
	}
	visibles := attrs.Attributes["ApproximateNumberOfMessages"]
	enVuelo := attrs.Attributes["ApproximateNumberOfMessagesNotVisible"]
	if visibles == "0" && enVuelo == "0" {
		t.Fatal("el mensaje se borro aunque la funcion fallo")
	}
}

// 🥇 LOS FALLOS PARCIALES. Un lote de cinco donde falla uno: sin decir cuál, se reintentan los
// cinco. Diciéndolo, se borran cuatro.
func TestDisparadorFallosParciales(t *testing.T) {
	lam, _, q, fns, _ := nubeConDisparadores(t, 20*time.Millisecond)
	crearFn(t, lam, "procesar", 5)
	url := crearColaSimple(t, q, "trabajos")

	// La función dice que le falló el mensaje cuyo cuerpo es "malo".
	atender(t, fns, "procesar", func(evento []byte) Resultado {
		var ev struct {
			Records []struct {
				MessageID string `json:"messageId"`
				Body      string `json:"body"`
			} `json:"Records"`
		}
		_ = json.Unmarshal(evento, &ev)
		fallos := []map[string]string{}
		for _, r := range ev.Records {
			if r.Body == "malo" {
				fallos = append(fallos, map[string]string{"itemIdentifier": r.MessageID})
			}
		}
		cuerpo, _ := json.Marshal(map[string]any{"batchItemFailures": fallos})
		return Resultado{Cuerpo: cuerpo}
	})
	mapear(t, lam, "trabajos", "procesar", 10)

	for _, cuerpo := range []string{"bueno1", "bueno2", "malo", "bueno3", "bueno4"} {
		enviar(t, q, url, cuerpo)
	}

	// Los cuatro buenos desaparecen; el malo se queda dando vueltas.
	hasta(t, 4*time.Second, "queden solo los fallidos", func() bool {
		v, n := cuentas(t, q, url)
		return v+n == 1
	})
}

// Y la trampa: si la función contesta CUALQUIER otra cosa, se borra el lote entero. Una errata
// en `itemIdentifier` no da error: da pérdida de mensajes.
func TestDisparadorUnaErrataEnElCampoBorraTodo(t *testing.T) {
	lam, _, q, fns, _ := nubeConDisparadores(t, 20*time.Millisecond)
	crearFn(t, lam, "procesar", 5)
	url := crearColaSimple(t, q, "trabajos")
	atender(t, fns, "procesar", func([]byte) Resultado {
		// Fíjate: `itemIdentifiers`, con ese plural de más. Nadie se queja.
		return Resultado{Cuerpo: []byte(`{"batchItemFailures":[{"itemIdentifiers":"x"}]}`)}
	})
	mapear(t, lam, "trabajos", "procesar", 10)

	enviar(t, q, url, "se va a perder")
	esperarVacia(t, q, url, 3*time.Second)
}

// El lote: si mandas cinco y el tamaño es cinco, llega UNA invocación con los cinco dentro.
func TestDisparadorElLoteLlegaJunto(t *testing.T) {
	lam, _, q, fns, _ := nubeConDisparadores(t, 200*time.Millisecond)
	crearFn(t, lam, "procesar", 5)
	url := crearColaSimple(t, q, "trabajos")
	e := atender(t, fns, "procesar", func([]byte) Resultado {
		return Resultado{Cuerpo: []byte(`"ok"`)}
	})

	for i := 0; i < 5; i++ {
		enviar(t, q, url, "mensaje")
	}
	mapear(t, lam, "trabajos", "procesar", 10)

	hasta(t, 3*time.Second, "llegue el lote", func() bool { return e.cuantos() > 0 })
	var ev struct {
		Records []json.RawMessage `json:"Records"`
	}
	_ = json.Unmarshal(e.ultimo(), &ev)
	if len(ev.Records) != 5 {
		t.Fatalf("el lote trajo %d registros y esperaba 5", len(ev.Records))
	}
	if e.cuantos() != 1 {
		t.Fatalf("hubo %d invocaciones y esperaba 1", e.cuantos())
	}
}

// Borrar el mapeo para el bucle, y ESPERA a que pare. Volver antes sería mentir.
func TestDisparadorBorrarElMapeoParaElBucle(t *testing.T) {
	lam, _, q, fns, disp := nubeConDisparadores(t, 20*time.Millisecond)
	crearFn(t, lam, "procesar", 5)
	url := crearColaSimple(t, q, "trabajos")
	e := atender(t, fns, "procesar", func([]byte) Resultado {
		return Resultado{Cuerpo: []byte(`"ok"`)}
	})
	uuid := mapear(t, lam, "trabajos", "procesar", 10)

	enviar(t, q, url, "uno")
	hasta(t, 3*time.Second, "se procese el primero", func() bool { return e.cuantos() > 0 })

	if _, err := lam.DeleteEventSourceMapping(context.Background(),
		&lambdasdk.DeleteEventSourceMappingInput{UUID: aws.String(uuid)}); err != nil {
		t.Fatalf("DeleteEventSourceMapping: %v", err)
	}
	if _, hay := disp.Buscar(uuid); hay {
		t.Fatal("el mapeo sigue en el registro")
	}

	// Y a partir de aquí, lo que se mande se queda en la cola.
	antes := e.cuantos()
	enviar(t, q, url, "dos")
	time.Sleep(200 * time.Millisecond)
	if e.cuantos() != antes {
		t.Fatal("el bucle siguio sacando mensajes despues de borrar el mapeo")
	}
}

// ── S3 -> Lambda ─────────────────────────────────────────────────────────

// 🥇 Subo un fichero y se ejecuta código. Sin bucles, sin llamar a nada.
func TestDisparadorS3InvocaAlSubir(t *testing.T) {
	lam, c, _, fns, disp := nubeConDisparadores(t, 20*time.Millisecond)
	crearFn(t, lam, "miniatura", 5)
	e := atender(t, fns, "miniatura", func([]byte) Resultado {
		return Resultado{Cuerpo: []byte(`"ok"`)}
	})

	ctx := context.Background()
	if _, err := c.CreateBucket(ctx, &s3.CreateBucketInput{Bucket: aws.String("fotos")}); err != nil {
		t.Fatalf("CreateBucket: %v", err)
	}
	notificar(t, c, "fotos", "miniatura", []string{"s3:ObjectCreated:*"}, "", "")

	if _, err := c.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String("fotos"), Key: aws.String("perros/luna.txt"),
		Body: strings.NewReader("una perra")}); err != nil {
		t.Fatalf("PutObject: %v", err)
	}
	disp.Esperar()

	if e.cuantos() != 1 {
		t.Fatalf("la funcion se ejecuto %d veces", e.cuantos())
	}
	var ev struct {
		Records []struct {
			EventName   string `json:"eventName"`
			EventSource string `json:"eventSource"`
			S3          struct {
				Bucket struct {
					Name string `json:"name"`
				} `json:"bucket"`
				Object struct {
					Key  string `json:"key"`
					Size int    `json:"size"`
				} `json:"object"`
			} `json:"s3"`
		} `json:"Records"`
	}
	if err := json.Unmarshal(e.ultimo(), &ev); err != nil {
		t.Fatalf("el evento no es JSON: %v", err)
	}
	r := ev.Records[0]
	if r.EventSource != "aws:s3" {
		t.Fatalf("eventSource = %q", r.EventSource)
	}
	// ⚠️ El nombre del evento viaja SIN el `s3:` de delante.
	if r.EventName != "ObjectCreated:Put" {
		t.Fatalf("eventName = %q, y viaja sin el s3: de delante", r.EventName)
	}
	if r.S3.Bucket.Name != "fotos" || r.S3.Object.Key != "perros/luna.txt" {
		t.Fatalf("bucket/clave = %s/%s", r.S3.Bucket.Name, r.S3.Object.Key)
	}
	if r.S3.Object.Size != 9 {
		t.Fatalf("tamaño = %d", r.S3.Object.Size)
	}
}

// 🥇 El filtro se evalúa en S3, NO dentro de tu función.
func TestDisparadorS3FiltraPorPrefijoYSufijo(t *testing.T) {
	lam, c, _, fns, disp := nubeConDisparadores(t, 20*time.Millisecond)
	crearFn(t, lam, "miniatura", 5)
	e := atender(t, fns, "miniatura", func([]byte) Resultado {
		return Resultado{Cuerpo: []byte(`"ok"`)}
	})

	ctx := context.Background()
	_, _ = c.CreateBucket(ctx, &s3.CreateBucketInput{Bucket: aws.String("fotos")})
	notificar(t, c, "fotos", "miniatura", []string{"s3:ObjectCreated:*"}, "originales/", ".png")

	subir := func(clave string) {
		t.Helper()
		if _, err := c.PutObject(ctx, &s3.PutObjectInput{
			Bucket: aws.String("fotos"), Key: aws.String(clave),
			Body: strings.NewReader("x")}); err != nil {
			t.Fatalf("PutObject %s: %v", clave, err)
		}
	}
	subir("originales/gato.png") // sí
	subir("originales/gato.txt") // no: sufijo
	subir("copias/gato.png")     // no: prefijo
	subir("gato.png")            // no: prefijo
	disp.Esperar()

	if e.cuantos() != 1 {
		t.Fatalf("se ejecuto %d veces y solo una clave encajaba con el filtro", e.cuantos())
	}
}

// Borrar también avisa, con otro nombre de evento.
func TestDisparadorS3AvisaAlBorrar(t *testing.T) {
	lam, c, _, fns, disp := nubeConDisparadores(t, 20*time.Millisecond)
	crearFn(t, lam, "auditar", 5)
	e := atender(t, fns, "auditar", func([]byte) Resultado {
		return Resultado{Cuerpo: []byte(`"ok"`)}
	})

	ctx := context.Background()
	_, _ = c.CreateBucket(ctx, &s3.CreateBucketInput{Bucket: aws.String("fotos")})
	notificar(t, c, "fotos", "auditar", []string{"s3:ObjectRemoved:*"}, "", "")

	// Subir NO avisa: la configuración solo pide los borrados.
	_, _ = c.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String("fotos"), Key: aws.String("gato.txt"),
		Body: strings.NewReader("x")})
	disp.Esperar()
	if e.cuantos() != 0 {
		t.Fatal("aviso de una creacion cuando solo se pidieron los borrados")
	}

	if _, err := c.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String("fotos"), Key: aws.String("gato.txt")}); err != nil {
		t.Fatalf("DeleteObject: %v", err)
	}
	disp.Esperar()
	if e.cuantos() != 1 {
		t.Fatalf("aviso %d veces del borrado", e.cuantos())
	}
	if !strings.Contains(string(e.ultimo()), "ObjectRemoved:Delete") {
		t.Fatalf("el evento no es un borrado: %s", e.ultimo())
	}
}

// S3 también puede avisar a una COLA en vez de a una función. Y así se junta todo: S3 encola y
// el disparador de la cola invoca.
func TestDisparadorS3PuedeEncolar(t *testing.T) {
	_, c, q, _, disp := nubeConDisparadores(t, 20*time.Millisecond)
	ctx := context.Background()
	url := crearColaSimple(t, q, "avisos")
	_, _ = c.CreateBucket(ctx, &s3.CreateBucketInput{Bucket: aws.String("fotos")})
	notificarCola(t, c, "fotos", "avisos", []string{"s3:ObjectCreated:*"})

	_, _ = c.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String("fotos"), Key: aws.String("gato.txt"),
		Body: strings.NewReader("x")})
	disp.Esperar()

	res, err := q.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{QueueUrl: aws.String(url)})
	if err != nil {
		t.Fatalf("ReceiveMessage: %v", err)
	}
	if len(res.Messages) != 1 {
		t.Fatalf("en la cola hay %d mensajes", len(res.Messages))
	}
	if !strings.Contains(aws.ToString(res.Messages[0].Body), "ObjectCreated:Put") {
		t.Fatalf("el cuerpo no es el evento de S3: %s", aws.ToString(res.Messages[0].Body))
	}
}

// Un bucket sin notificaciones no avisa a nadie: es como estuvo siete episodios.
func TestDisparadorSinNotificacionNoPasaNada(t *testing.T) {
	lam, c, _, fns, disp := nubeConDisparadores(t, 20*time.Millisecond)
	crearFn(t, lam, "miniatura", 5)
	e := atender(t, fns, "miniatura", func([]byte) Resultado {
		return Resultado{Cuerpo: []byte(`"ok"`)}
	})
	ctx := context.Background()
	_, _ = c.CreateBucket(ctx, &s3.CreateBucketInput{Bucket: aws.String("fotos")})
	_, _ = c.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String("fotos"), Key: aws.String("gato.txt"),
		Body: strings.NewReader("x")})
	disp.Esperar()
	if e.cuantos() != 0 {
		t.Fatal("invoco sin que nadie lo hubiera pedido")
	}
}

// Y la configuración se puede volver a leer, que es lo que hace el SDK para comprobar.
func TestDisparadorLaConfiguracionSeLee(t *testing.T) {
	lam, c, _, _, _ := nubeConDisparadores(t, 20*time.Millisecond)
	crearFn(t, lam, "miniatura", 5)
	ctx := context.Background()
	_, _ = c.CreateBucket(ctx, &s3.CreateBucketInput{Bucket: aws.String("fotos")})
	notificar(t, c, "fotos", "miniatura", []string{"s3:ObjectCreated:*"}, "originales/", ".png")

	got, err := c.GetBucketNotificationConfiguration(ctx,
		&s3.GetBucketNotificationConfigurationInput{Bucket: aws.String("fotos")})
	if err != nil {
		t.Fatalf("GetBucketNotificationConfiguration: %v", err)
	}
	if len(got.LambdaFunctionConfigurations) != 1 {
		t.Fatalf("volvieron %d configuraciones", len(got.LambdaFunctionConfigurations))
	}
	cfg := got.LambdaFunctionConfigurations[0]
	if !strings.HasSuffix(aws.ToString(cfg.LambdaFunctionArn), ":miniatura") {
		t.Fatalf("ARN = %q", aws.ToString(cfg.LambdaFunctionArn))
	}
	if len(cfg.Events) != 1 || string(cfg.Events[0]) != "s3:ObjectCreated:*" {
		t.Fatalf("eventos = %v", cfg.Events)
	}
}

// ── ayudas ───────────────────────────────────────────────────────────────

type sqstypesNombre = sqstypes.QueueAttributeName

func mapear(t *testing.T, lam *lambdasdk.Client, cola, funcion string, lote int32) string {
	t.Helper()
	r, err := lam.CreateEventSourceMapping(context.Background(),
		&lambdasdk.CreateEventSourceMappingInput{
			EventSourceArn: aws.String("arn:aws:sqs:us-east-1:000000000000:" + cola),
			FunctionName:   aws.String(funcion),
			BatchSize:      aws.Int32(lote),
		})
	if err != nil {
		t.Fatalf("CreateEventSourceMapping: %v", err)
	}
	return aws.ToString(r.UUID)
}

func enviar(t *testing.T, q *sqs.Client, url, cuerpo string) {
	t.Helper()
	if _, err := q.SendMessage(context.Background(), &sqs.SendMessageInput{
		QueueUrl: aws.String(url), MessageBody: aws.String(cuerpo)}); err != nil {
		t.Fatalf("SendMessage: %v", err)
	}
}

func cuentas(t *testing.T, q *sqs.Client, url string) (visibles, enVuelo int) {
	t.Helper()
	attrs, err := q.GetQueueAttributes(context.Background(), &sqs.GetQueueAttributesInput{
		QueueUrl: aws.String(url), AttributeNames: []sqstypesNombre{
			"ApproximateNumberOfMessages", "ApproximateNumberOfMessagesNotVisible"}})
	if err != nil {
		t.Fatalf("GetQueueAttributes: %v", err)
	}
	return atoiSeguro(attrs.Attributes["ApproximateNumberOfMessages"]),
		atoiSeguro(attrs.Attributes["ApproximateNumberOfMessagesNotVisible"])
}

func esperarVacia(t *testing.T, q *sqs.Client, url string, cuanto time.Duration) {
	t.Helper()
	hasta(t, cuanto, "la cola se vacie", func() bool {
		v, n := cuentas(t, q, url)
		return v == 0 && n == 0
	})
}

func atoiSeguro(s string) int {
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return n
		}
		n = n*10 + int(c-'0')
	}
	return n
}

func notificar(t *testing.T, c *s3.Client, bucket, funcion string, eventos []string,
	prefijo, sufijo string) {
	t.Helper()
	cfg := s3types.LambdaFunctionConfiguration{
		LambdaFunctionArn: aws.String(arnDe(funcion)),
	}
	for _, e := range eventos {
		cfg.Events = append(cfg.Events, s3types.Event(e))
	}
	var reglas []s3types.FilterRule
	if prefijo != "" {
		reglas = append(reglas, s3types.FilterRule{
			Name: s3types.FilterRuleNamePrefix, Value: aws.String(prefijo)})
	}
	if sufijo != "" {
		reglas = append(reglas, s3types.FilterRule{
			Name: s3types.FilterRuleNameSuffix, Value: aws.String(sufijo)})
	}
	if len(reglas) > 0 {
		cfg.Filter = &s3types.NotificationConfigurationFilter{
			Key: &s3types.S3KeyFilter{FilterRules: reglas}}
	}
	if _, err := c.PutBucketNotificationConfiguration(context.Background(),
		&s3.PutBucketNotificationConfigurationInput{
			Bucket: aws.String(bucket),
			NotificationConfiguration: &s3types.NotificationConfiguration{
				LambdaFunctionConfigurations: []s3types.LambdaFunctionConfiguration{cfg},
			}}); err != nil {
		t.Fatalf("PutBucketNotificationConfiguration: %v", err)
	}
}

func notificarCola(t *testing.T, c *s3.Client, bucket, cola string, eventos []string) {
	t.Helper()
	cfg := s3types.QueueConfiguration{
		QueueArn: aws.String("arn:aws:sqs:us-east-1:000000000000:" + cola)}
	for _, e := range eventos {
		cfg.Events = append(cfg.Events, s3types.Event(e))
	}
	if _, err := c.PutBucketNotificationConfiguration(context.Background(),
		&s3.PutBucketNotificationConfigurationInput{
			Bucket: aws.String(bucket),
			NotificationConfiguration: &s3types.NotificationConfiguration{
				QueueConfigurations: []s3types.QueueConfiguration{cfg},
			}}); err != nil {
		t.Fatalf("PutBucketNotificationConfiguration: %v", err)
	}
}
