package main

// Las pruebas de SNS. Quien publica y quien se suscribe es el SDK oficial de AWS; quien recibe
// es el SDK oficial de SQS. Si los dos se entienden con lo que hay en medio, es compatible.
//
//	go test ./... -count=1 -race -run SNS

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	snstypes "github.com/aws/aws-sdk-go-v2/service/sns/types"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

func nubeConSNS(t *testing.T) (*sns.Client, *sqs.Client) {
	t.Helper()
	colas := NuevasColas()
	temas := NuevosTemas()
	mux := http.NewServeMux()
	NuevoS3(NuevoAlmacen(), nil).Rutas(mux)
	srv := httptest.NewServer(Raiz(mux, map[string]http.Handler{
		"AmazonSQS.": NuevoSQS(colas, nil),
	}, nil, NuevoSNS(temas, colas, nil)))
	t.Cleanup(srv.Close)

	return sns.New(sns.Options{
		Region:       "us-east-1",
		Credentials:  credencialesFijas("prueba", "prueba"),
		BaseEndpoint: aws.String(srv.URL),
	}), clienteSQS(srv.URL)
}

// tema y cola: las dos cosas que se repiten en todas las pruebas.
func prepara(t *testing.T, n *sns.Client, q *sqs.Client, tema string,
	colas ...string) (string, map[string]string) {
	t.Helper()
	ctx := context.Background()
	r, err := n.CreateTopic(ctx, &sns.CreateTopicInput{Name: aws.String(tema)})
	if err != nil {
		t.Fatalf("CreateTopic: %v", err)
	}
	urls := map[string]string{}
	for _, c := range colas {
		cr, err := q.CreateQueue(ctx, &sqs.CreateQueueInput{QueueName: aws.String(c)})
		if err != nil {
			t.Fatalf("CreateQueue %s: %v", c, err)
		}
		urls[c] = aws.ToString(cr.QueueUrl)
	}
	return aws.ToString(r.TopicArn), urls
}

func suscribir(t *testing.T, n *sns.Client, tema, cola string,
	attrs map[string]string) string {
	t.Helper()
	r, err := n.Subscribe(context.Background(), &sns.SubscribeInput{
		TopicArn:   aws.String(tema),
		Protocol:   aws.String("sqs"),
		Endpoint:   aws.String("arn:aws:sqs:us-east-1:000000000000:" + cola),
		Attributes: attrs,
	})
	if err != nil {
		t.Fatalf("Subscribe: %v", err)
	}
	return aws.ToString(r.SubscriptionArn)
}

func recibir(t *testing.T, q *sqs.Client, url string) []string {
	t.Helper()
	r, err := q.ReceiveMessage(context.Background(), &sqs.ReceiveMessageInput{
		QueueUrl: aws.String(url), MaxNumberOfMessages: 10})
	if err != nil {
		t.Fatalf("ReceiveMessage: %v", err)
	}
	fuera := []string{}
	for _, m := range r.Messages {
		fuera = append(fuera, aws.ToString(m.Body))
	}
	return fuera
}

// 🥇 LA DIFERENCIA ENTRE UNA COLA Y UN TEMA, en una prueba: un mensaje entra y salen tres.
func TestSNSUnMensajeEntraYSalenTres(t *testing.T) {
	n, q := nubeConSNS(t)
	tema, urls := prepara(t, n, q, "pedidos", "facturacion", "almacen", "avisos")
	for _, c := range []string{"facturacion", "almacen", "avisos"} {
		suscribir(t, n, tema, c, map[string]string{"RawMessageDelivery": "true"})
	}

	if _, err := n.Publish(context.Background(), &sns.PublishInput{
		TopicArn: aws.String(tema), Message: aws.String("pedido 42")}); err != nil {
		t.Fatalf("Publish: %v", err)
	}

	for _, c := range []string{"facturacion", "almacen", "avisos"} {
		msgs := recibir(t, q, urls[c])
		if len(msgs) != 1 {
			t.Fatalf("en %s hay %d mensajes y esperaba 1", c, len(msgs))
		}
		if msgs[0] != "pedido 42" {
			t.Fatalf("en %s llego %q", c, msgs[0])
		}
	}
}

// 🥇 LA TRAMPA DEL EPISODIO: sin `RawMessageDelivery`, el mensaje llega ENVUELTO. Tu consumidor
// lee el cuerpo y se encuentra un sobre, no la carta.
func TestSNSSinRawElMensajeLlegaEnvuelto(t *testing.T) {
	n, q := nubeConSNS(t)
	tema, urls := prepara(t, n, q, "pedidos", "facturacion")
	suscribir(t, n, tema, "facturacion", nil) // sin atributos: lo normal

	if _, err := n.Publish(context.Background(), &sns.PublishInput{
		TopicArn: aws.String(tema),
		Subject:  aws.String("un pedido"),
		Message:  aws.String(`{"pedido":42}`)}); err != nil {
		t.Fatalf("Publish: %v", err)
	}

	msgs := recibir(t, q, urls["facturacion"])
	if len(msgs) != 1 {
		t.Fatalf("llegaron %d mensajes", len(msgs))
	}
	// Lo primero que hace todo el mundo: leer el cuerpo como si fuera lo que publicó.
	var loQueCreias struct {
		Pedido int `json:"pedido"`
	}
	_ = json.Unmarshal([]byte(msgs[0]), &loQueCreias)
	if loQueCreias.Pedido == 42 {
		t.Fatal("el mensaje llego sin envolver, y sin RawMessageDelivery deberia venir en un sobre")
	}

	// Y lo que hay de verdad es el sobre.
	var sobre struct {
		Type     string `json:"Type"`
		TopicArn string `json:"TopicArn"`
		Subject  string `json:"Subject"`
		Message  string `json:"Message"`
	}
	if err := json.Unmarshal([]byte(msgs[0]), &sobre); err != nil {
		t.Fatalf("el cuerpo no es el sobre de SNS: %v\n%s", err, msgs[0])
	}
	if sobre.Type != "Notification" {
		t.Fatalf("Type = %q", sobre.Type)
	}
	if sobre.Subject != "un pedido" {
		t.Fatalf("Subject = %q", sobre.Subject)
	}
	// 🥇 Y lo que publicaste esta dentro, y ademas como CADENA: hay que deserializar dos veces.
	if sobre.Message != `{"pedido":42}` {
		t.Fatalf("Message = %q", sobre.Message)
	}
	var deVerdad struct {
		Pedido int `json:"pedido"`
	}
	if err := json.Unmarshal([]byte(sobre.Message), &deVerdad); err != nil {
		t.Fatalf("el segundo nivel tampoco: %v", err)
	}
	if deVerdad.Pedido != 42 {
		t.Fatalf("pedido = %d", deVerdad.Pedido)
	}
}

// Y con el atributo puesto llega la carta, sin sobre.
func TestSNSConRawLlegaLaCarta(t *testing.T) {
	n, q := nubeConSNS(t)
	tema, urls := prepara(t, n, q, "pedidos", "facturacion")
	suscribir(t, n, tema, "facturacion", map[string]string{"RawMessageDelivery": "true"})

	if _, err := n.Publish(context.Background(), &sns.PublishInput{
		TopicArn: aws.String(tema), Message: aws.String(`{"pedido":42}`)}); err != nil {
		t.Fatalf("Publish: %v", err)
	}
	msgs := recibir(t, q, urls["facturacion"])
	if len(msgs) != 1 || msgs[0] != `{"pedido":42}` {
		t.Fatalf("llego %v", msgs)
	}
}

// El atributo también se puede cambiar después, con SetSubscriptionAttributes.
func TestSNSElAtributoSePuedeCambiarDespues(t *testing.T) {
	n, q := nubeConSNS(t)
	tema, urls := prepara(t, n, q, "pedidos", "facturacion")
	arn := suscribir(t, n, tema, "facturacion", nil)

	if _, err := n.SetSubscriptionAttributes(context.Background(),
		&sns.SetSubscriptionAttributesInput{
			SubscriptionArn: aws.String(arn),
			AttributeName:   aws.String("RawMessageDelivery"),
			AttributeValue:  aws.String("true")}); err != nil {
		t.Fatalf("SetSubscriptionAttributes: %v", err)
	}
	if _, err := n.Publish(context.Background(), &sns.PublishInput{
		TopicArn: aws.String(tema), Message: aws.String("sin sobre")}); err != nil {
		t.Fatalf("Publish: %v", err)
	}
	msgs := recibir(t, q, urls["facturacion"])
	if len(msgs) != 1 || msgs[0] != "sin sobre" {
		t.Fatalf("llego %v", msgs)
	}

	// Y se puede volver a leer.
	got, err := n.GetSubscriptionAttributes(context.Background(),
		&sns.GetSubscriptionAttributesInput{SubscriptionArn: aws.String(arn)})
	if err != nil {
		t.Fatalf("GetSubscriptionAttributes: %v", err)
	}
	if got.Attributes["RawMessageDelivery"] != "true" {
		t.Fatalf("el atributo volvio como %q", got.Attributes["RawMessageDelivery"])
	}
}

// 🥇 EL FILTRO SE EVALÚA EN LA SUSCRIPCIÓN, no en quien publica. Cada suscriptor dice de qué
// quiere enterarse, y quien publica no sabe ni cuántos hay.
func TestSNSElFiltroLoPoneElSuscriptor(t *testing.T) {
	n, q := nubeConSNS(t)
	tema, urls := prepara(t, n, q, "pedidos", "solo-es", "todo")
	suscribir(t, n, tema, "solo-es", map[string]string{
		"RawMessageDelivery": "true",
		"FilterPolicy":       `{"pais":["ES"]}`,
	})
	suscribir(t, n, tema, "todo", map[string]string{"RawMessageDelivery": "true"})

	publica := func(cuerpo, pais string) {
		t.Helper()
		if _, err := n.Publish(context.Background(), &sns.PublishInput{
			TopicArn: aws.String(tema), Message: aws.String(cuerpo),
			MessageAttributes: map[string]snstypes.MessageAttributeValue{
				"pais": {DataType: aws.String("String"), StringValue: aws.String(pais)},
			}}); err != nil {
			t.Fatalf("Publish: %v", err)
		}
	}
	publica("de españa", "ES")
	publica("de francia", "FR")

	soloES := recibir(t, q, urls["solo-es"])
	if len(soloES) != 1 || soloES[0] != "de españa" {
		t.Fatalf("la cola filtrada recibio %v", soloES)
	}
	todo := recibir(t, q, urls["todo"])
	if len(todo) != 2 {
		t.Fatalf("la cola sin filtro recibio %d mensajes", len(todo))
	}
}

// ⚠️ Y el motivo número uno de «mi filtro no deja pasar nada»: si el mensaje NO trae el
// atributo, no pasa. No es «da igual»: es que no coincide.
func TestSNSSinElAtributoElFiltroNoDejaPasar(t *testing.T) {
	n, q := nubeConSNS(t)
	tema, urls := prepara(t, n, q, "pedidos", "solo-es")
	suscribir(t, n, tema, "solo-es", map[string]string{
		"RawMessageDelivery": "true",
		"FilterPolicy":       `{"pais":["ES"]}`,
	})

	// Publicado SIN atributos.
	if _, err := n.Publish(context.Background(), &sns.PublishInput{
		TopicArn: aws.String(tema), Message: aws.String("sin atributos")}); err != nil {
		t.Fatalf("Publish: %v", err)
	}
	if msgs := recibir(t, q, urls["solo-es"]); len(msgs) != 0 {
		t.Fatalf("paso el filtro sin traer el atributo: %v", msgs)
	}
}

// 🥇 Un tema NO guarda nada. Publicar sin suscriptores no pierde el mensaje: es que nunca
// existió para nadie.
func TestSNSSinSuscriptoresElMensajeNoExiste(t *testing.T) {
	n, q := nubeConSNS(t)
	tema, urls := prepara(t, n, q, "pedidos", "tarde")

	// Se publica ANTES de suscribir.
	if _, err := n.Publish(context.Background(), &sns.PublishInput{
		TopicArn: aws.String(tema), Message: aws.String("nadie escuchaba")}); err != nil {
		t.Fatalf("Publish: %v", err)
	}
	suscribir(t, n, tema, "tarde", map[string]string{"RawMessageDelivery": "true"})

	if msgs := recibir(t, q, urls["tarde"]); len(msgs) != 0 {
		t.Fatalf("recibio algo publicado antes de existir: %v", msgs)
	}

	// Y a partir de ahora sí.
	if _, err := n.Publish(context.Background(), &sns.PublishInput{
		TopicArn: aws.String(tema), Message: aws.String("ahora sí")}); err != nil {
		t.Fatalf("Publish: %v", err)
	}
	if msgs := recibir(t, q, urls["tarde"]); len(msgs) != 1 {
		t.Fatalf("recibio %d", len(msgs))
	}
}

// Crear un tema dos veces devuelve el mismo. SNS es idempotente ahí, a propósito.
func TestSNSCrearDosVecesDaElMismoTema(t *testing.T) {
	n, _ := nubeConSNS(t)
	ctx := context.Background()
	a, err := n.CreateTopic(ctx, &sns.CreateTopicInput{Name: aws.String("pedidos")})
	if err != nil {
		t.Fatalf("CreateTopic: %v", err)
	}
	b, err := n.CreateTopic(ctx, &sns.CreateTopicInput{Name: aws.String("pedidos")})
	if err != nil {
		t.Fatalf("CreateTopic 2: %v", err)
	}
	if aws.ToString(a.TopicArn) != aws.ToString(b.TopicArn) {
		t.Fatalf("%s != %s", aws.ToString(a.TopicArn), aws.ToString(b.TopicArn))
	}
	lista, err := n.ListTopics(ctx, &sns.ListTopicsInput{})
	if err != nil {
		t.Fatalf("ListTopics: %v", err)
	}
	if len(lista.Topics) != 1 {
		t.Fatalf("hay %d temas", len(lista.Topics))
	}
}

// Desuscribirse deja de recibir. Y borrar el tema se lleva sus suscripciones por delante.
func TestSNSDesuscribirYBorrar(t *testing.T) {
	n, q := nubeConSNS(t)
	ctx := context.Background()
	tema, urls := prepara(t, n, q, "pedidos", "facturacion")
	arn := suscribir(t, n, tema, "facturacion", map[string]string{"RawMessageDelivery": "true"})

	if _, err := n.Unsubscribe(ctx, &sns.UnsubscribeInput{
		SubscriptionArn: aws.String(arn)}); err != nil {
		t.Fatalf("Unsubscribe: %v", err)
	}
	if _, err := n.Publish(ctx, &sns.PublishInput{
		TopicArn: aws.String(tema), Message: aws.String("ya no")}); err != nil {
		t.Fatalf("Publish: %v", err)
	}
	if msgs := recibir(t, q, urls["facturacion"]); len(msgs) != 0 {
		t.Fatalf("recibio despues de desuscribirse: %v", msgs)
	}

	// Y borrar el tema deja el registro limpio.
	suscribir(t, n, tema, "facturacion", nil)
	if _, err := n.DeleteTopic(ctx, &sns.DeleteTopicInput{TopicArn: aws.String(tema)}); err != nil {
		t.Fatalf("DeleteTopic: %v", err)
	}
	subs, err := n.ListSubscriptionsByTopic(ctx, &sns.ListSubscriptionsByTopicInput{
		TopicArn: aws.String(tema)})
	if err != nil {
		t.Fatalf("ListSubscriptionsByTopic: %v", err)
	}
	if len(subs.Subscriptions) != 0 {
		t.Fatalf("quedan %d suscripciones de un tema borrado", len(subs.Subscriptions))
	}
}

// Publicar en un tema que no existe da un error tipado, no un silencio.
func TestSNSPublicarEnUnTemaQueNoExiste(t *testing.T) {
	n, _ := nubeConSNS(t)
	_, err := n.Publish(context.Background(), &sns.PublishInput{
		TopicArn: aws.String("arn:aws:sns:us-east-1:000000000000:fantasma"),
		Message:  aws.String("hola")})
	if err == nil {
		t.Fatal("publico en un tema que no existe")
	}
	if !strings.Contains(err.Error(), "NotFound") && !strings.Contains(err.Error(), "no existe") {
		t.Fatalf("el error no dice que no existe: %v", err)
	}
}

// Las suscripciones se pueden listar, que es lo que hace cualquiera para comprobar quién está
// apuntado.
func TestSNSListarSuscripciones(t *testing.T) {
	n, q := nubeConSNS(t)
	tema, _ := prepara(t, n, q, "pedidos", "a", "b")
	suscribir(t, n, tema, "a", nil)
	suscribir(t, n, tema, "b", nil)

	r, err := n.ListSubscriptionsByTopic(context.Background(),
		&sns.ListSubscriptionsByTopicInput{TopicArn: aws.String(tema)})
	if err != nil {
		t.Fatalf("ListSubscriptionsByTopic: %v", err)
	}
	if len(r.Subscriptions) != 2 {
		t.Fatalf("hay %d suscripciones", len(r.Subscriptions))
	}
	for _, s := range r.Subscriptions {
		if aws.ToString(s.Protocol) != "sqs" {
			t.Fatalf("protocolo = %q", aws.ToString(s.Protocol))
		}
		if aws.ToString(s.TopicArn) != tema {
			t.Fatalf("tema = %q", aws.ToString(s.TopicArn))
		}
	}
}

// Un protocolo que no sabemos entregar se acepta, pero no llega nada. Y eso hay que decirlo.
func TestSNSOtroProtocoloSeApuntaPeroNoLlega(t *testing.T) {
	n, q := nubeConSNS(t)
	tema, urls := prepara(t, n, q, "pedidos", "facturacion")
	suscribir(t, n, tema, "facturacion", map[string]string{"RawMessageDelivery": "true"})

	if _, err := n.Subscribe(context.Background(), &sns.SubscribeInput{
		TopicArn: aws.String(tema),
		Protocol: aws.String("email"),
		Endpoint: aws.String("alguien@ejemplo.com")}); err != nil {
		t.Fatalf("Subscribe email: %v", err)
	}
	if _, err := n.Publish(context.Background(), &sns.PublishInput{
		TopicArn: aws.String(tema), Message: aws.String("hola")}); err != nil {
		t.Fatalf("Publish: %v", err)
	}
	// La cola sí recibe; el correo no existe aquí, y eso no rompe la entrega a los demás.
	if msgs := recibir(t, q, urls["facturacion"]); len(msgs) != 1 {
		t.Fatalf("la cola recibio %d", len(msgs))
	}
}

// ⚠️ Y el detalle de protocolo del episodio: SNS habla con un FORMULARIO, no con JSON.
func TestSNSHablaConUnFormulario(t *testing.T) {
	visto := make(chan string, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case visto <- r.Header.Get("Content-Type"):
		default:
		}
		w.Header().Set("Content-Type", "text/xml")
		_, _ = w.Write([]byte(`<?xml version="1.0"?><CreateTopicResponse ` +
			`xmlns="http://sns.amazonaws.com/doc/2010-03-31/"><CreateTopicResult>` +
			`<TopicArn>arn:aws:sns:us-east-1:000000000000:x</TopicArn></CreateTopicResult>` +
			`<ResponseMetadata><RequestId>1</RequestId></ResponseMetadata></CreateTopicResponse>`))
	}))
	defer srv.Close()

	c := sns.New(sns.Options{Region: "us-east-1",
		Credentials: credencialesFijas("prueba", "prueba"), BaseEndpoint: aws.String(srv.URL)})
	if _, err := c.CreateTopic(context.Background(),
		&sns.CreateTopicInput{Name: aws.String("x")}); err != nil {
		t.Fatalf("CreateTopic: %v", err)
	}
	tipo := <-visto
	if !strings.HasPrefix(tipo, "application/x-www-form-urlencoded") {
		t.Fatalf("SNS mando %q, y esperaba un formulario: es el protocolo Query de siempre", tipo)
	}
}
