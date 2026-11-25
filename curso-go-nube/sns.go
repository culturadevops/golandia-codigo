package main

// El servicio SNS. Episodio 9.
//
// ⚠️ Y aqui hay una sorpresa de protocolo que no me esperaba: SNS NO habla como DynamoDB ni como
// SQS. Habla el protocolo «Query» de toda la vida de AWS: un formulario en el cuerpo, la
// operacion en un campo llamado `Action`, y la respuesta en texto etiquetado.
//
//	POST /
//	Content-Type: application/x-www-form-urlencoded
//	Action=Publish&TopicArn=arn:aws:sns:...&Message=hola
//
// O sea que es el CUARTO protocolo distinto dentro de la misma nube:
//
//	S3         metodo y ruta, respuesta en texto etiquetado
//	DynamoDB   POST a "/", JSON, operacion en una cabecera
//	SQS        igual que DynamoDB (lo migraron; SNS no)
//	SNS        formulario en el cuerpo, operacion en un campo, respuesta etiquetada
//
// Es exactamente el protocolo que tenia SQS antes de que Amazon lo modernizara. SNS se quedo
// donde estaba, y por eso hoy son primos que no se parecen.

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type ServicioSNS struct {
	temas    *Temas
	colas    *Colas // para entregar de verdad
	registro func(formato string, args ...any)
}

func NuevoSNS(t *Temas, c *Colas, registro func(string, ...any)) *ServicioSNS {
	if registro == nil {
		registro = func(string, ...any) {}
	}
	return &ServicioSNS{temas: t, colas: c, registro: registro}
}

func (s *ServicioSNS) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		s.fallo(w, 400, "InvalidParameterValue", err.Error())
		return
	}
	accion := r.PostForm.Get("Action")
	s.registro("POST / [SNS %s]", accion)

	switch accion {
	case "CreateTopic":
		s.crearTema(w, r)
	case "ListTopics":
		s.listarTemas(w, r)
	case "DeleteTopic":
		s.borrarTema(w, r)
	case "Subscribe":
		s.suscribir(w, r)
	case "Unsubscribe":
		s.desuscribir(w, r)
	case "ListSubscriptionsByTopic":
		s.listarSuscripciones(w, r)
	case "SetSubscriptionAttributes":
		s.ponerAtributo(w, r)
	case "GetSubscriptionAttributes":
		s.leerAtributos(w, r)
	case "Publish":
		s.publicar(w, r)
	default:
		s.fallo(w, 400, "InvalidAction", "no implemento "+accion)
	}
}

// ── las operaciones ──────────────────────────────────────────────────────

func (s *ServicioSNS) crearTema(w http.ResponseWriter, r *http.Request) {
	nombre := r.PostForm.Get("Name")
	if nombre == "" {
		s.fallo(w, 400, "InvalidParameterValue", "hace falta un Name")
		return
	}
	t := s.temas.Crear(nombre)
	s.registro("  -> %s", t.ARN)
	s.responder(w, "CreateTopicResponse",
		"<CreateTopicResult><TopicArn>"+xmlEscapa(t.ARN)+"</TopicArn></CreateTopicResult>")
}

func (s *ServicioSNS) listarTemas(w http.ResponseWriter, r *http.Request) {
	var b strings.Builder
	b.WriteString("<ListTopicsResult><Topics>")
	for _, t := range s.temas.Listar() {
		fmt.Fprintf(&b, "<member><TopicArn>%s</TopicArn></member>", xmlEscapa(t.ARN))
	}
	b.WriteString("</Topics></ListTopicsResult>")
	s.responder(w, "ListTopicsResponse", b.String())
}

func (s *ServicioSNS) borrarTema(w http.ResponseWriter, r *http.Request) {
	arn := r.PostForm.Get("TopicArn")
	// ⚠️ SNS contesta bien aunque el tema no exista. Borrar es idempotente, igual que en S3:
	// si devolvieras un error, un reintento de red fallaria despues de haber borrado bien.
	s.temas.Borrar(arn)
	s.responder(w, "DeleteTopicResponse", "")
}

func (s *ServicioSNS) suscribir(w http.ResponseWriter, r *http.Request) {
	tema := r.PostForm.Get("TopicArn")
	protocolo := r.PostForm.Get("Protocol")
	destino := r.PostForm.Get("Endpoint")

	sus, err := s.temas.Suscribir(tema, protocolo, destino)
	if err != nil {
		s.fallo(w, 404, "NotFound", err.Error())
		return
	}
	// Los atributos pueden venir ya en la propia llamada, en un mapa aplanado:
	//   Attributes.entry.1.key=RawMessageDelivery&Attributes.entry.1.value=true
	for k, v := range atributosDeFormulario(r.PostForm, "Attributes") {
		if err := s.temas.PonerAtributo(sus.ARN, k, v); err != nil {
			s.fallo(w, 400, "InvalidParameter", err.Error())
			return
		}
	}
	if !strings.EqualFold(protocolo, "sqs") {
		s.registro("  -> ⚠️ protocolo %q: se apunta, pero aqui no se entrega nada", protocolo)
	}
	s.registro("  -> %s", sus.ARN)
	s.responder(w, "SubscribeResponse", "<SubscribeResult><SubscriptionArn>"+
		xmlEscapa(sus.ARN)+"</SubscriptionArn></SubscribeResult>")
}

func (s *ServicioSNS) desuscribir(w http.ResponseWriter, r *http.Request) {
	s.temas.Desuscribir(r.PostForm.Get("SubscriptionArn"))
	s.responder(w, "UnsubscribeResponse", "")
}

func (s *ServicioSNS) listarSuscripciones(w http.ResponseWriter, r *http.Request) {
	var b strings.Builder
	b.WriteString("<ListSubscriptionsByTopicResult><Subscriptions>")
	for _, x := range s.temas.Suscripciones(r.PostForm.Get("TopicArn")) {
		fmt.Fprintf(&b, "<member><SubscriptionArn>%s</SubscriptionArn>"+
			"<Protocol>%s</Protocol><Endpoint>%s</Endpoint><TopicArn>%s</TopicArn>"+
			"<Owner>000000000000</Owner></member>",
			xmlEscapa(x.ARN), xmlEscapa(x.Protocol), xmlEscapa(x.Destino), xmlEscapa(x.TemaARN))
	}
	b.WriteString("</Subscriptions></ListSubscriptionsByTopicResult>")
	s.responder(w, "ListSubscriptionsByTopicResponse", b.String())
}

func (s *ServicioSNS) ponerAtributo(w http.ResponseWriter, r *http.Request) {
	arn := r.PostForm.Get("SubscriptionArn")
	nombre := r.PostForm.Get("AttributeName")
	valor := r.PostForm.Get("AttributeValue")
	s.registro("  -> %s = %s", nombre, valor)
	if err := s.temas.PonerAtributo(arn, nombre, valor); err != nil {
		s.fallo(w, 404, "NotFound", err.Error())
		return
	}
	s.responder(w, "SetSubscriptionAttributesResponse", "")
}

func (s *ServicioSNS) leerAtributos(w http.ResponseWriter, r *http.Request) {
	sus, hay := s.temas.BuscarSuscripcion(r.PostForm.Get("SubscriptionArn"))
	if !hay {
		s.fallo(w, 404, "NotFound", errNoExisteSuscr.Error())
		return
	}
	var b strings.Builder
	b.WriteString("<GetSubscriptionAttributesResult><Attributes>")
	for _, e := range [][2]string{
		{"SubscriptionArn", sus.ARN},
		{"TopicArn", sus.TemaARN},
		{"Protocol", sus.Protocol},
		{"Endpoint", sus.Destino},
		{"RawMessageDelivery", strconv.FormatBool(sus.Crudo)},
	} {
		fmt.Fprintf(&b, "<entry><key>%s</key><value>%s</value></entry>",
			xmlEscapa(e[0]), xmlEscapa(e[1]))
	}
	b.WriteString("</Attributes></GetSubscriptionAttributesResult>")
	s.responder(w, "GetSubscriptionAttributesResponse", b.String())
}

// publicar es el corazon: una entrada, N salidas.
func (s *ServicioSNS) publicar(w http.ResponseWriter, r *http.Request) {
	tema := r.PostForm.Get("TopicArn")
	mensaje := r.PostForm.Get("Message")
	asunto := r.PostForm.Get("Subject")
	attrs := atributosMensaje(r.PostForm)

	id, entregas, err := s.temas.Publicar(tema, mensaje, asunto, attrs)
	if err != nil {
		s.fallo(w, 404, "NotFound", err.Error())
		return
	}
	// 🥇 Y aqui se ve la diferencia con una cola, sin comentarios que la expliquen: un bucle.
	for _, e := range entregas {
		if _, err := s.colas.Enviar(e.Cola, e.Cuerpo); err != nil {
			s.registro("  -> no puedo entregar en %s: %v", e.Cola, err)
			continue
		}
	}
	s.registro("  -> %s repartido a %d suscriptor(es)", id, len(entregas))

	s.responder(w, "PublishResponse",
		"<PublishResult><MessageId>"+xmlEscapa(id)+"</MessageId></PublishResult>")
}

// ── el formulario ────────────────────────────────────────────────────────

// atributosMensaje saca los atributos del formulario aplanado que manda el SDK:
//
//	MessageAttributes.entry.1.Name=tipo
//	MessageAttributes.entry.1.Value.DataType=String
//	MessageAttributes.entry.1.Value.StringValue=pedido
//
// ⚠️ Ese aplanado es lo que hace que el protocolo Query envejezca mal: un mapa dentro de una
// lista dentro de un formulario. Es legible, y es un dolor.
func atributosMensaje(form map[string][]string) map[string]AtributoMensaje {
	nombres := map[string]string{}
	tipos := map[string]string{}
	valores := map[string]string{}
	for clave, vs := range form {
		if len(vs) == 0 || !strings.HasPrefix(clave, "MessageAttributes.entry.") {
			continue
		}
		resto := strings.TrimPrefix(clave, "MessageAttributes.entry.")
		i := strings.Index(resto, ".")
		if i < 0 {
			continue
		}
		n, campo := resto[:i], resto[i+1:]
		switch campo {
		case "Name":
			nombres[n] = vs[0]
		case "Value.DataType":
			tipos[n] = vs[0]
		case "Value.StringValue":
			valores[n] = vs[0]
		}
	}
	fuera := map[string]AtributoMensaje{}
	for n, nombre := range nombres {
		fuera[nombre] = AtributoMensaje{Tipo: tipos[n], Valor: valores[n]}
	}
	return fuera
}

// atributosDeFormulario lee un mapa aplanado tipo `Attributes.entry.1.key/value`.
func atributosDeFormulario(form map[string][]string, prefijo string) map[string]string {
	claves := map[string]string{}
	valores := map[string]string{}
	for clave, vs := range form {
		if len(vs) == 0 || !strings.HasPrefix(clave, prefijo+".entry.") {
			continue
		}
		resto := strings.TrimPrefix(clave, prefijo+".entry.")
		i := strings.Index(resto, ".")
		if i < 0 {
			continue
		}
		n, campo := resto[:i], resto[i+1:]
		switch campo {
		case "key":
			claves[n] = vs[0]
		case "value":
			valores[n] = vs[0]
		}
	}
	fuera := map[string]string{}
	for n, k := range claves {
		fuera[k] = valores[n]
	}
	return fuera
}

// ── las respuestas ───────────────────────────────────────────────────────

// responder envuelve el resultado con la metadata que el SDK espera.
//
// ⚠️ El texto etiquetado se escribe A MANO en vez de dejar que la biblioteca lo genere desde una
// struct, y no es pereza: el protocolo Query pide nombres de etiqueta muy concretos
// (`<Topics><member>...`), y describirlos con etiquetas de campo sale mas largo y menos claro
// que escribirlos. Lo que NO se hace a mano es escapar: eso siempre con la biblioteca.
func (s *ServicioSNS) responder(w http.ResponseWriter, raiz, resultado string) {
	w.Header().Set("Content-Type", "text/xml")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(xml.Header))
	_, _ = fmt.Fprintf(w, `<%s xmlns="http://sns.amazonaws.com/doc/2010-03-31/">%s`+
		`<ResponseMetadata><RequestId>00000000-0000-4000-8000-000000000000</RequestId>`+
		`</ResponseMetadata></%s>`, raiz, resultado, raiz)
}

func (s *ServicioSNS) fallo(w http.ResponseWriter, estado int, codigo, mensaje string) {
	s.registro("  -> %d %s", estado, codigo)
	w.Header().Set("Content-Type", "text/xml")
	w.WriteHeader(estado)
	_, _ = w.Write([]byte(xml.Header))
	_, _ = fmt.Fprintf(w, `<ErrorResponse xmlns="http://sns.amazonaws.com/doc/2010-03-31/">`+
		`<Error><Type>Sender</Type><Code>%s</Code><Message>%s</Message></Error>`+
		`<RequestId>00000000-0000-4000-8000-000000000000</RequestId></ErrorResponse>`,
		xmlEscapa(codigo), xmlEscapa(mensaje))
}

func xmlEscapa(s string) string {
	var b strings.Builder
	_ = xml.EscapeText(&b, []byte(s))
	return b.String()
}
