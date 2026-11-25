package main

// El servicio SQS. Habla el mismo protocolo que DynamoDB —POST a "/", JSON, y la operacion en
// la cabecera— pero con otro prefijo:
//
//	X-Amz-Target: AmazonSQS.SendMessage
//
// ⚠️ Y aqui hay un dato que sorprende: SQS es MAS viejo que DynamoDB y durante años hablo el
// protocolo "Query" de AWS, con los parametros en el cuerpo tipo formulario. Amazon lo migro a
// JSON, y por eso hoy se parece a DynamoDB y no a S3. O sea: el protocolo de un servicio no es
// para siempre.
//
// Lo que se implementa, que es lo que se usa un dia normal:
//
//	CreateQueue · ListQueues · GetQueueUrl · DeleteQueue · PurgeQueue · GetQueueAttributes
//	SendMessage · SendMessageBatch
//	ReceiveMessage · DeleteMessage · DeleteMessageBatch · ChangeMessageVisibility
//
// ⛔ Lo que NO: colas FIFO, DelaySeconds, atributos de mensaje, sondeo largo de verdad
// (WaitTimeSeconds se acepta pero no bloquea) ni politicas de acceso.

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var errNoExisteCola = errors.New("la cola no existe")

const tipoSQS = "com.amazonaws.sqs#"

type ServicioSQS struct {
	colas    *Colas
	registro func(formato string, args ...any)
}

func NuevoSQS(c *Colas, registro func(string, ...any)) *ServicioSQS {
	if registro == nil {
		registro = func(string, ...any) {}
	}
	return &ServicioSQS{colas: c, registro: registro}
}

func (s *ServicioSQS) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	destino := r.Header.Get("X-Amz-Target")
	op := destino
	if i := strings.LastIndex(destino, "."); i >= 0 {
		op = destino[i+1:]
	}
	cuerpo, err := io.ReadAll(r.Body)
	if err != nil {
		s.fallo(w, "SerializationException", err.Error())
		return
	}
	s.registro("POST / [SQS %s] %d byte(s)", op, len(cuerpo))

	// ⚠️ La URL de la cola se arma con el Host de LA PETICION, no con una direccion fija de
	// arranque. Dentro de un contenedor el servidor escucha en 0.0.0.0, y devolver esa
	// direccion daria una URL que no le sirve a nadie. Asi, el cliente recibe la misma
	// direccion por la que entro.
	base := "http://" + r.Host

	switch op {
	case "CreateQueue":
		s.crearCola(w, cuerpo, base)
	case "ListQueues":
		s.listarColas(w, cuerpo, base)
	case "GetQueueUrl":
		s.urlDeCola(w, cuerpo, base)
	case "DeleteQueue":
		s.borrarCola(w, cuerpo)
	case "PurgeQueue":
		s.vaciarCola(w, cuerpo)
	case "GetQueueAttributes":
		s.atributos(w, cuerpo)
	case "SendMessage":
		s.enviar(w, cuerpo)
	case "SendMessageBatch":
		s.enviarLote(w, cuerpo)
	case "ReceiveMessage":
		s.recibir(w, cuerpo)
	case "DeleteMessage":
		s.borrarMensaje(w, cuerpo)
	case "DeleteMessageBatch":
		s.borrarLote(w, cuerpo)
	case "ChangeMessageVisibility":
		s.cambiarVisibilidad(w, cuerpo)
	default:
		s.fallo(w, "UnsupportedOperation",
			fmt.Sprintf("nubecita no implementa %q (episodio 3: S3, DynamoDB y SQS)", op))
	}
}

func (s *ServicioSQS) responder(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/x-amz-json-1.0")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(v)
}

func (s *ServicioSQS) fallo(w http.ResponseWriter, tipo, mensaje string) {
	s.registro("  -> 400 %s", tipo)
	w.Header().Set("Content-Type", "application/x-amz-json-1.0")
	w.WriteHeader(http.StatusBadRequest)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"__type": tipoSQS + tipo, "message": mensaje,
	})
}

// urlDeLaCola arma la URL con el numero de cuenta de mentira que usa todo el mundo en local.
func urlDeLaCola(base, nombre string) string { return base + "/000000000000/" + nombre }

// nombreDeURL saca el nombre del final de la URL de la cola. En SQS la cola se identifica por su
// URL entera, no por su nombre: el nombre es solo el ultimo tramo.
func nombreDeURL(u string) string {
	if i := strings.LastIndex(u, "/"); i >= 0 {
		return u[i+1:]
	}
	return u
}

// ── colas ────────────────────────────────────────────────────────────────

type peticionCrearCola struct {
	QueueName  string            `json:"QueueName"`
	Attributes map[string]string `json:"Attributes"`
}

// redrive es el `RedrivePolicy`, que en SQS viaja como una CADENA con JSON dentro. Sí: JSON
// dentro de un campo de texto de otro JSON. No es un error de lectura.
type redrive struct {
	MaxReceiveCount     any    `json:"maxReceiveCount"`
	DeadLetterTargetArn string `json:"deadLetterTargetArn"`
}

func (s *ServicioSQS) crearCola(w http.ResponseWriter, cuerpo []byte, base string) {
	var p peticionCrearCola
	if err := json.Unmarshal(cuerpo, &p); err != nil {
		s.fallo(w, "SerializationException", err.Error())
		return
	}
	if p.QueueName == "" {
		s.fallo(w, "InvalidParameterValue", "falta QueueName")
		return
	}
	var vis *time.Duration
	if v, hay := p.Attributes["VisibilityTimeout"]; hay {
		if n, err := strconv.Atoi(v); err == nil {
			d := time.Duration(n) * time.Second
			vis = &d
		}
	}
	maxRecibos, colaMuertos := 0, ""
	if v, hay := p.Attributes["RedrivePolicy"]; hay && v != "" {
		var rd redrive
		if err := json.Unmarshal([]byte(v), &rd); err != nil {
			s.fallo(w, "InvalidParameterValue", "RedrivePolicy no es JSON valido: "+err.Error())
			return
		}
		// ⚠️ maxReceiveCount llega unas veces como numero y otras como cadena, segun quien
		// arme la peticion. Aceptar solo uno de los dos es un fallo que solo aparece con el
		// cliente que no probaste.
		switch t := rd.MaxReceiveCount.(type) {
		case float64:
			maxRecibos = int(t)
		case string:
			maxRecibos, _ = strconv.Atoi(t)
		}
		colaMuertos = nombreDeURL(rd.DeadLetterTargetArn)
		if i := strings.LastIndex(colaMuertos, ":"); i >= 0 {
			colaMuertos = colaMuertos[i+1:]
		}
	}
	s.colas.Crear(p.QueueName, vis, maxRecibos, colaMuertos)
	s.registro("  -> 200 cola %q creada", p.QueueName)
	s.responder(w, map[string]any{"QueueUrl": urlDeLaCola(base, p.QueueName)})
}

func (s *ServicioSQS) listarColas(w http.ResponseWriter, cuerpo []byte, base string) {
	var p struct {
		QueueNamePrefix string `json:"QueueNamePrefix"`
	}
	_ = json.Unmarshal(cuerpo, &p)
	nombres := s.colas.Listar(p.QueueNamePrefix)
	urls := make([]string, 0, len(nombres))
	for _, n := range nombres {
		urls = append(urls, urlDeLaCola(base, n))
	}
	s.registro("  -> 200 %d cola(s)", len(urls))
	s.responder(w, map[string]any{"QueueUrls": urls})
}

func (s *ServicioSQS) urlDeCola(w http.ResponseWriter, cuerpo []byte, base string) {
	var p struct {
		QueueName string `json:"QueueName"`
	}
	_ = json.Unmarshal(cuerpo, &p)
	if _, ok := s.colas.Buscar(p.QueueName); !ok {
		s.fallo(w, "QueueDoesNotExist", "The specified queue does not exist.")
		return
	}
	s.responder(w, map[string]any{"QueueUrl": urlDeLaCola(base, p.QueueName)})
}

func (s *ServicioSQS) borrarCola(w http.ResponseWriter, cuerpo []byte) {
	var p struct {
		QueueUrl string `json:"QueueUrl"`
	}
	_ = json.Unmarshal(cuerpo, &p)
	if !s.colas.Borrar(nombreDeURL(p.QueueUrl)) {
		s.fallo(w, "QueueDoesNotExist", "The specified queue does not exist.")
		return
	}
	s.registro("  -> 200 cola borrada")
	s.responder(w, map[string]any{})
}

func (s *ServicioSQS) vaciarCola(w http.ResponseWriter, cuerpo []byte) {
	var p struct {
		QueueUrl string `json:"QueueUrl"`
	}
	_ = json.Unmarshal(cuerpo, &p)
	if !s.colas.Vaciar(nombreDeURL(p.QueueUrl)) {
		s.fallo(w, "QueueDoesNotExist", "The specified queue does not exist.")
		return
	}
	s.responder(w, map[string]any{})
}

func (s *ServicioSQS) atributos(w http.ResponseWriter, cuerpo []byte) {
	var p struct {
		QueueUrl string `json:"QueueUrl"`
	}
	_ = json.Unmarshal(cuerpo, &p)
	nombre := nombreDeURL(p.QueueUrl)
	visibles, enVuelo, ok := s.colas.Cuentas(nombre)
	if !ok {
		s.fallo(w, "QueueDoesNotExist", "The specified queue does not exist.")
		return
	}
	c, _ := s.colas.Buscar(nombre)
	attr := map[string]string{
		"ApproximateNumberOfMessages":           strconv.Itoa(visibles),
		"ApproximateNumberOfMessagesNotVisible": strconv.Itoa(enVuelo),
		"VisibilityTimeout":                     strconv.Itoa(int(c.Visibilidad.Seconds())),
		"QueueArn":                              "arn:aws:sqs:us-east-1:000000000000:" + nombre,
	}
	s.registro("  -> 200 %d visible(s), %d en vuelo", visibles, enVuelo)
	s.responder(w, map[string]any{"Attributes": attr})
}

// ── mensajes ─────────────────────────────────────────────────────────────

func (s *ServicioSQS) enviar(w http.ResponseWriter, cuerpo []byte) {
	var p struct {
		QueueUrl    string `json:"QueueUrl"`
		MessageBody string `json:"MessageBody"`
	}
	if err := json.Unmarshal(cuerpo, &p); err != nil {
		s.fallo(w, "SerializationException", err.Error())
		return
	}
	m, err := s.colas.Enviar(nombreDeURL(p.QueueUrl), p.MessageBody)
	if err != nil {
		s.fallo(w, "QueueDoesNotExist", "The specified queue does not exist.")
		return
	}
	s.registro("  -> 200 enviado %s", m.ID)
	// ⚠️ El MD5 no es decorativo: el SDK lo COMPRUEBA. Si lo calculas mal, el cliente da un
	// error de checksum y parece que se corrompio la red.
	s.responder(w, map[string]any{"MessageId": m.ID, "MD5OfMessageBody": m.MD5})
}

func (s *ServicioSQS) enviarLote(w http.ResponseWriter, cuerpo []byte) {
	var p struct {
		QueueUrl string `json:"QueueUrl"`
		Entries  []struct {
			Id          string `json:"Id"`
			MessageBody string `json:"MessageBody"`
		} `json:"Entries"`
	}
	if err := json.Unmarshal(cuerpo, &p); err != nil {
		s.fallo(w, "SerializationException", err.Error())
		return
	}
	nombre := nombreDeURL(p.QueueUrl)
	var bien []map[string]any
	var mal []map[string]any
	for _, e := range p.Entries {
		m, err := s.colas.Enviar(nombre, e.MessageBody)
		if err != nil {
			mal = append(mal, map[string]any{"Id": e.Id, "Code": "QueueDoesNotExist",
				"SenderFault": false, "Message": err.Error()})
			continue
		}
		bien = append(bien, map[string]any{"Id": e.Id, "MessageId": m.ID,
			"MD5OfMessageBody": m.MD5})
	}
	s.registro("  -> 200 lote: %d bien, %d mal", len(bien), len(mal))
	res := map[string]any{"Successful": bien}
	if mal != nil {
		res["Failed"] = mal
	}
	s.responder(w, res)
}

func (s *ServicioSQS) recibir(w http.ResponseWriter, cuerpo []byte) {
	var p struct {
		QueueUrl            string `json:"QueueUrl"`
		MaxNumberOfMessages int    `json:"MaxNumberOfMessages"`
		VisibilityTimeout   *int   `json:"VisibilityTimeout"`
	}
	if err := json.Unmarshal(cuerpo, &p); err != nil {
		s.fallo(w, "SerializationException", err.Error())
		return
	}
	if p.MaxNumberOfMessages <= 0 {
		p.MaxNumberOfMessages = 1
	}
	var vis *time.Duration
	if p.VisibilityTimeout != nil {
		d := time.Duration(*p.VisibilityTimeout) * time.Second
		vis = &d
	}
	msgs, err := s.colas.Recibir(nombreDeURL(p.QueueUrl), p.MaxNumberOfMessages, vis)
	if err != nil {
		s.fallo(w, "QueueDoesNotExist", "The specified queue does not exist.")
		return
	}
	salida := make([]map[string]any, 0, len(msgs))
	for _, m := range msgs {
		salida = append(salida, map[string]any{
			"MessageId":     m.ID,
			"ReceiptHandle": m.Recibo,
			"Body":          m.Cuerpo,
			"MD5OfBody":     m.MD5,
			"Attributes": map[string]string{
				"ApproximateReceiveCount": strconv.Itoa(m.Recibido),
				"SentTimestamp":           strconv.FormatInt(m.Enviado.UnixMilli(), 10),
			},
		})
	}
	s.registro("  -> 200 %d mensaje(s) entregados", len(salida))
	// ⚠️ Recibir cero mensajes NO es un error: es una cola vacia. Si devolvieras error, el
	// consumidor no distinguiria "no hay trabajo" de "se rompio la cola".
	res := map[string]any{}
	if len(salida) > 0 {
		res["Messages"] = salida
	}
	s.responder(w, res)
}

func (s *ServicioSQS) borrarMensaje(w http.ResponseWriter, cuerpo []byte) {
	var p struct {
		QueueUrl      string `json:"QueueUrl"`
		ReceiptHandle string `json:"ReceiptHandle"`
	}
	_ = json.Unmarshal(cuerpo, &p)
	borrado, err := s.colas.Quitar(nombreDeURL(p.QueueUrl), p.ReceiptHandle)
	if err != nil {
		s.fallo(w, "QueueDoesNotExist", "The specified queue does not exist.")
		return
	}
	// Borrar con un recibo caducado NO es un error en SQS. Se registra, porque casi siempre
	// significa que el consumidor tardo mas que la visibilidad y otro ya se lo llevo.
	s.registro("  -> 200 borrado=%v", borrado)
	s.responder(w, map[string]any{})
}

func (s *ServicioSQS) borrarLote(w http.ResponseWriter, cuerpo []byte) {
	var p struct {
		QueueUrl string `json:"QueueUrl"`
		Entries  []struct {
			Id            string `json:"Id"`
			ReceiptHandle string `json:"ReceiptHandle"`
		} `json:"Entries"`
	}
	_ = json.Unmarshal(cuerpo, &p)
	nombre := nombreDeURL(p.QueueUrl)
	var bien []map[string]any
	for _, e := range p.Entries {
		if _, err := s.colas.Quitar(nombre, e.ReceiptHandle); err != nil {
			s.fallo(w, "QueueDoesNotExist", "The specified queue does not exist.")
			return
		}
		bien = append(bien, map[string]any{"Id": e.Id})
	}
	s.responder(w, map[string]any{"Successful": bien})
}

func (s *ServicioSQS) cambiarVisibilidad(w http.ResponseWriter, cuerpo []byte) {
	var p struct {
		QueueUrl          string `json:"QueueUrl"`
		ReceiptHandle     string `json:"ReceiptHandle"`
		VisibilityTimeout int    `json:"VisibilityTimeout"`
	}
	_ = json.Unmarshal(cuerpo, &p)
	ok, err := s.colas.CambiarVisibilidad(nombreDeURL(p.QueueUrl), p.ReceiptHandle,
		time.Duration(p.VisibilityTimeout)*time.Second)
	if err != nil {
		s.fallo(w, "QueueDoesNotExist", "The specified queue does not exist.")
		return
	}
	if !ok {
		s.fallo(w, "ReceiptHandleIsInvalid", "The receipt handle is not valid.")
		return
	}
	s.registro("  -> 200 visibilidad a %d s", p.VisibilityTimeout)
	s.responder(w, map[string]any{})
}
