package main

// Los disparadores. Episodio 8, y es el que junta todo lo anterior.
//
// Hasta aqui habia servicios sueltos: un almacen, unas colas, unas funciones. Y en los dos
// episodios de Lambda quedo escrito el hueco con todas las letras: «no hay disparadores: nadie
// invoca la funcion sola». Esto lo tapa.
//
// Hay dos formas de que algo invoque a tu funcion sin que tu llames, y son MUY distintas:
//
//	SQS -> Lambda   un BUCLE que hace ReceiveMessage por ti y te invoca con lo que saca.
//	                Lo llaman «event source mapping». No es magia: es un for.
//
//	S3  -> Lambda   S3 AVISA cuando pasa algo. No hay bucle: hay una llamada.
//	                Se configura en el BUCKET, no en la funcion.
//
// 🥇 Y de la primera sale la regla que explica media vida operando colas: TU FUNCION NO BORRA EL
// MENSAJE. Lo borra el disparador, y solo si fuiste bien. Por eso un consumidor de Lambda no
// llama nunca a DeleteMessage — y por eso, si tu funcion falla, el mensaje vuelve.
//
// 🥇 Y de ahi sale la segunda: si de un lote de diez falla UNO y no dices cual, se reintentan
// LOS DIEZ. Los nueve buenos se procesan dos veces. Decir cual es lo que hace
// `batchItemFailures`, y es la diferencia entre un consumidor que funciona y uno que cobra dos
// veces.

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

// ── el registro de disparadores ──────────────────────────────────────────

// Mapeo es un disparador de cola: una funcion, una cola, y un bucle.
type Mapeo struct {
	UUID    string
	Funcion string
	Cola    string
	Lote    int
	Estado  string

	parar context.CancelFunc
	hecho chan struct{}
}

type Disparadores struct {
	fns    *Funciones
	lan    *Lanzador
	colas  *Colas
	espera time.Duration // cuanto se espera cuando la cola esta vacia
	region string

	registro func(formato string, args ...any)

	mu     sync.Mutex
	mapeos map[string]*Mapeo
	sec    int

	// ⚠️ Este grupo existe para las PRUEBAS, y no es trampa: sirve para esperar a que las
	// notificaciones asincronas terminen en vez de dormir un rato y cruzar los dedos. Una
	// prueba que duerme es una prueba que un dia falla en otra maquina.
	enVuelo sync.WaitGroup
}

func NuevosDisparadores(fns *Funciones, lan *Lanzador, colas *Colas, espera time.Duration,
	registro func(string, ...any)) *Disparadores {
	if registro == nil {
		registro = func(string, ...any) {}
	}
	if espera <= 0 {
		espera = 200 * time.Millisecond
	}
	return &Disparadores{fns: fns, lan: lan, colas: colas, espera: espera,
		region: "us-east-1", registro: registro, mapeos: map[string]*Mapeo{}}
}

// Esperar bloquea hasta que no queda ninguna notificacion en vuelo.
func (d *Disparadores) Esperar() { d.enVuelo.Wait() }

// invocar es la puerta por la que entran TODAS las invocaciones que no vienen de una peticion
// HTTP. Hace lo mismo que el plano de datos: pone el plazo y se asegura de que haya runtime.
func (d *Disparadores) invocar(ctx context.Context, nombre string, evento []byte) (Resultado, error) {
	fn, err := d.fns.Buscar(nombre)
	if err != nil {
		return Resultado{}, err
	}
	ctx, cancelar := context.WithTimeout(ctx, time.Duration(fn.Segundos)*time.Second)
	defer cancelar()
	d.lan.Asegurar(ctx, fn)
	return d.fns.Invocar(ctx, fn, evento)
}

// ── SQS -> Lambda: el bucle ──────────────────────────────────────────────

func (d *Disparadores) Crear(funcion, cola string, lote int, activo bool) (*Mapeo, error) {
	if _, err := d.fns.Buscar(funcion); err != nil {
		return nil, err
	}
	if _, hay := d.colas.Buscar(cola); !hay {
		return nil, errNoExisteCola
	}
	if lote <= 0 {
		lote = 10 // el mismo valor por defecto que Amazon
	}

	d.mu.Lock()
	d.sec++
	m := &Mapeo{
		UUID:    fmt.Sprintf("00000000-0000-4000-9000-%012d", d.sec),
		Funcion: funcion, Cola: cola, Lote: lote, Estado: "Disabled",
	}
	d.mapeos[m.UUID] = m
	d.mu.Unlock()

	if activo {
		d.Arrancar(m)
	}
	d.registro("  -> disparador %s: %s -> %s (lote %d, %s)", corto(m.UUID), cola, funcion,
		lote, m.Estado)
	return m, nil
}

// Arrancar pone en marcha el bucle de un mapeo.
//
// ⚠️ La gorrutina que sale de aqui NO vive dentro de una peticion HTTP: vive lo que viva el
// proceso. Asi que se le da su propio contexto y una forma de pararla, y `Borrar` la espera. Una
// gorrutina que nadie sabe parar es una fuga con buenos modales.
func (d *Disparadores) Arrancar(m *Mapeo) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if m.parar != nil {
		return // ya esta corriendo
	}
	ctx, parar := context.WithCancel(context.Background())
	m.parar, m.hecho, m.Estado = parar, make(chan struct{}), "Enabled"
	go d.sondear(ctx, m)
}

func (d *Disparadores) sondear(ctx context.Context, m *Mapeo) {
	defer close(m.hecho)
	d.registro("  -> sondeando %s para %s", m.Cola, m.Funcion)
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		msgs, err := d.colas.Recibir(m.Cola, m.Lote, nil)
		if err != nil {
			// La cola ya no existe: no tiene sentido seguir preguntando.
			d.registro("  -> el disparador %s se para: %v", corto(m.UUID), err)
			return
		}
		if len(msgs) == 0 {
			// Vacia: se espera un poco. En Amazon esto es el sondeo largo; aqui, una siesta.
			select {
			case <-ctx.Done():
				return
			case <-time.After(d.espera):
			}
			continue
		}

		d.entregar(ctx, m, msgs)
	}
}

// entregar es el corazon del episodio: invoca con el lote y decide QUE SE BORRA.
func (d *Disparadores) entregar(ctx context.Context, m *Mapeo, msgs []*Mensaje) {
	evento, err := eventoSQS(msgs, m.Cola, d.region)
	if err != nil {
		d.registro("  -> no puedo armar el evento: %v", err)
		return
	}
	d.registro("  -> [%s] invoco %s con %d mensaje(s)", corto(m.UUID), m.Funcion, len(msgs))

	res, err := d.invocar(ctx, m.Funcion, evento)

	switch {
	case err != nil:
		// Ni siquiera se pudo ejecutar. No se borra NADA: cuando venza la visibilidad, los
		// mensajes vuelven y otro lo intentara.
		d.registro("  -> [%s] no se pudo invocar (%v): no borro nada", corto(m.UUID), err)
	case res.Fallo:
		// 🥇 Tu funcion devolvio error. Tampoco se borra nada, y ESO es lo que hace que un
		// mensaje se reintente sin que tu escribas una linea de reintentos.
		d.registro("  -> [%s] la funcion fallo: los %d mensajes vuelven a la cola",
			corto(m.UUID), len(msgs))
	default:
		fallidos := idsFallidos(res.Cuerpo)
		borrados := 0
		for _, msg := range msgs {
			if fallidos[msg.ID] {
				continue // este lo pidio ella misma: se queda
			}
			if ok, _ := d.colas.Quitar(m.Cola, msg.Recibo); ok {
				borrados++
			}
		}
		if len(fallidos) > 0 {
			d.registro("  -> [%s] borro %d y dejo %d que la funcion marco como fallidos",
				corto(m.UUID), borrados, len(fallidos))
		} else {
			d.registro("  -> [%s] borro los %d", corto(m.UUID), borrados)
		}
	}
}

// idsFallidos lee `{"batchItemFailures":[{"itemIdentifier":"..."}]}` de la respuesta.
//
// 🥇 ⚠️ Y aqui esta la trampa que se lleva por delante a mucha gente: si la respuesta NO tiene
// esa forma —porque devolviste otra cosa, o porque escribiste mal el nombre del campo— este
// conjunto sale VACIO, y entonces se borra el lote entero. O sea: una errata en
// `itemIdentifier` no da error, da perdida de mensajes.
func idsFallidos(cuerpo []byte) map[string]bool {
	var r struct {
		BatchItemFailures []struct {
			ItemIdentifier string `json:"itemIdentifier"`
		} `json:"batchItemFailures"`
	}
	if err := json.Unmarshal(cuerpo, &r); err != nil {
		return nil
	}
	if len(r.BatchItemFailures) == 0 {
		return nil
	}
	fuera := make(map[string]bool, len(r.BatchItemFailures))
	for _, f := range r.BatchItemFailures {
		if f.ItemIdentifier != "" {
			fuera[f.ItemIdentifier] = true
		}
	}
	return fuera
}

// eventoSQS arma el JSON que le llega a tu manejador. Es el formato de verdad: si escribes tu
// funcion contra esto, funciona contra Amazon sin tocar nada.
func eventoSQS(msgs []*Mensaje, cola, region string) ([]byte, error) {
	type registro struct {
		MessageID     string            `json:"messageId"`
		ReceiptHandle string            `json:"receiptHandle"`
		Body          string            `json:"body"`
		Attributes    map[string]string `json:"attributes"`
		MD5OfBody     string            `json:"md5OfBody"`
		EventSource   string            `json:"eventSource"`
		EventSourceAR string            `json:"eventSourceARN"`
		AWSRegion     string            `json:"awsRegion"`
	}
	regs := make([]registro, 0, len(msgs))
	for _, m := range msgs {
		regs = append(regs, registro{
			MessageID: m.ID, ReceiptHandle: m.Recibo, Body: m.Cuerpo,
			Attributes: map[string]string{
				"ApproximateReceiveCount": fmt.Sprint(m.Recibido),
				"SentTimestamp":           fmt.Sprint(m.Enviado.UnixMilli()),
			},
			MD5OfBody:   m.MD5,
			EventSource: "aws:sqs", EventSourceAR: arnCola(cola, region), AWSRegion: region,
		})
	}
	return json.Marshal(map[string]any{"Records": regs})
}

func arnCola(nombre, region string) string {
	return "arn:aws:sqs:" + region + ":000000000000:" + nombre
}

// nombreDeARN saca el ultimo tramo de un ARN. Vale para colas y para temas.
func nombreDeARN(arn string) string {
	if i := strings.LastIndex(arn, ":"); i >= 0 {
		return arn[i+1:]
	}
	return arn
}

func (d *Disparadores) Listar() []*Mapeo {
	d.mu.Lock()
	defer d.mu.Unlock()
	fuera := make([]*Mapeo, 0, len(d.mapeos))
	for _, m := range d.mapeos {
		fuera = append(fuera, m)
	}
	ordenarPor(fuera, func(m *Mapeo) string { return m.UUID })
	return fuera
}

func (d *Disparadores) Buscar(uuid string) (*Mapeo, bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	m, ok := d.mapeos[uuid]
	return m, ok
}

// Borrar para el bucle y ESPERA a que termine. Volver antes seria mentir: el bucle podria
// seguir sacando mensajes de una cola que el que llamo cree que ya nadie mira.
func (d *Disparadores) Borrar(uuid string) (*Mapeo, bool) {
	d.mu.Lock()
	m, ok := d.mapeos[uuid]
	if ok {
		delete(d.mapeos, uuid)
	}
	d.mu.Unlock()
	if !ok {
		return nil, false
	}
	d.pararMapeo(m)
	return m, true
}

func (d *Disparadores) pararMapeo(m *Mapeo) {
	d.mu.Lock()
	parar, hecho := m.parar, m.hecho
	m.parar, m.Estado = nil, "Disabled"
	d.mu.Unlock()
	if parar != nil {
		parar()
		<-hecho
	}
}

// Apagar para todos los bucles. Se llama al cerrar.
func (d *Disparadores) Apagar() {
	if d == nil {
		return
	}
	for _, m := range d.Listar() {
		d.pararMapeo(m)
	}
	d.enVuelo.Wait()
}

// ── S3 -> Lambda: las notificaciones ─────────────────────────────────────

// Aqui NO hay bucle. S3 avisa cuando pasa algo, y avisar es una llamada.
//
// ⚠️ Y la configuracion vive en el BUCKET, no en la funcion. Eso sorprende siempre: buscas en
// tu funcion quien la llama y no lo encuentras, porque esta escrito al otro lado.

type reglaFiltro struct {
	Nombre string `xml:"Name"`
	Valor  string `xml:"Value"`
}

type filtroXML struct {
	Clave struct {
		Reglas []reglaFiltro `xml:"FilterRule"`
	} `xml:"S3Key"`
}

// ⚠️ Los nombres de estas etiquetas NO son los del SDK. En Go el campo se llama
// `LambdaFunctionArn` y en el XML la etiqueta es `CloudFunction`, a secas; y `QueueArn` viaja
// como `Queue`. Es historia: cuando se diseño esto, Lambda todavia se llamaba «cloud function».
// Se descubre mirando el XML que manda el SDK de verdad, no leyendo la documentacion.
type configLambdaXML struct {
	ID      string     `xml:"Id,omitempty"`
	ARN     string     `xml:"CloudFunction"`
	Eventos []string   `xml:"Event"`
	Filtro  *filtroXML `xml:"Filter,omitempty"`
}

type configColaXML struct {
	ID      string     `xml:"Id,omitempty"`
	ARN     string     `xml:"Queue"`
	Eventos []string   `xml:"Event"`
	Filtro  *filtroXML `xml:"Filter,omitempty"`
}

type notificacionXML struct {
	XMLName xml.Name          `xml:"NotificationConfiguration"`
	Lambdas []configLambdaXML `xml:"CloudFunctionConfiguration"`
	Colas   []configColaXML   `xml:"QueueConfiguration"`
}

// Notificador guarda que hay que avisar por cada bucket, y avisa.
type Notificador struct {
	disp     *Disparadores
	colas    *Colas
	registro func(formato string, args ...any)

	mu        sync.Mutex
	porBucket map[string]notificacionXML
}

func NuevoNotificador(d *Disparadores, colas *Colas, registro func(string, ...any)) *Notificador {
	if registro == nil {
		registro = func(string, ...any) {}
	}
	return &Notificador{disp: d, colas: colas, registro: registro,
		porBucket: map[string]notificacionXML{}}
}

func (n *Notificador) Poner(bucket string, cfg notificacionXML) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.porBucket[bucket] = cfg
}

func (n *Notificador) Config(bucket string) notificacionXML {
	n.mu.Lock()
	defer n.mu.Unlock()
	return n.porBucket[bucket]
}

// coincide decide si esta configuracion quiere enterarse de este evento y esta clave.
//
// ⚠️ El filtro se evalua AQUI, en S3, no dentro de tu funcion. Si tu funcion se ejecuta con un
// fichero que no querias, el filtro esta mal escrito — no es que «llegue de todo».
func coincide(eventos []string, filtro *filtroXML, evento, clave string) bool {
	quiere := false
	for _, e := range eventos {
		// Los eventos admiten comodin al final: `s3:ObjectCreated:*`.
		if e == evento || (strings.HasSuffix(e, "*") && strings.HasPrefix(evento, strings.TrimSuffix(e, "*"))) {
			quiere = true
			break
		}
	}
	if !quiere {
		return false
	}
	if filtro == nil {
		return true
	}
	for _, r := range filtro.Clave.Reglas {
		switch strings.ToLower(r.Nombre) {
		case "prefix":
			if !strings.HasPrefix(clave, r.Valor) {
				return false
			}
		case "suffix":
			if !strings.HasSuffix(clave, r.Valor) {
				return false
			}
		}
	}
	return true
}

// Aviso dispara lo que haya configurado para este bucket.
//
// 🥇 Se hace en OTRA gorrutina a proposito, y esa es la respuesta a una pregunta que se hace
// todo el mundo: si mi funcion falla, ¿falla mi PutObject? No. Cuando S3 te contesta, tu objeto
// ya esta guardado y la notificacion todavia no ha salido. Son dos cosas distintas, y por eso
// una notificacion perdida no deja el objeto a medias.
func (n *Notificador) Aviso(bucket, clave, etag string, tam int, evento string) {
	cfg := n.Config(bucket)
	if len(cfg.Lambdas) == 0 && len(cfg.Colas) == 0 {
		return
	}
	cuerpo, err := eventoS3(bucket, clave, etag, tam, evento, n.disp.region)
	if err != nil {
		n.registro("  -> no puedo armar la notificacion: %v", err)
		return
	}

	for _, c := range cfg.Lambdas {
		if !coincide(c.Eventos, c.Filtro, evento, clave) {
			continue
		}
		funcion := nombreDeARN(c.ARN)
		n.disp.enVuelo.Add(1)
		go func() {
			defer n.disp.enVuelo.Done()
			n.registro("  -> [S3] %s %s/%s invoca %s", evento, bucket, clave, funcion)
			if _, err := n.disp.invocar(context.Background(), funcion, cuerpo); err != nil {
				// Y no pasa nada mas. En Amazon habria reintentos y una cola de fallos; aqui
				// se dice en voz alta y se sigue.
				n.registro("  -> [S3] la invocacion de %s fallo: %v", funcion, err)
			}
		}()
	}

	for _, c := range cfg.Colas {
		if !coincide(c.Eventos, c.Filtro, evento, clave) {
			continue
		}
		cola := nombreDeARN(c.ARN)
		n.disp.enVuelo.Add(1)
		go func() {
			defer n.disp.enVuelo.Done()
			n.registro("  -> [S3] %s %s/%s encola en %s", evento, bucket, clave, cola)
			if _, err := n.colas.Enviar(cola, string(cuerpo)); err != nil {
				n.registro("  -> [S3] no puedo encolar en %s: %v", cola, err)
			}
		}()
	}
}

// eventoS3 arma el JSON que recibe tu funcion. Es el formato de verdad, con su `Records` y todo.
func eventoS3(bucket, clave, etag string, tam int, evento, region string) ([]byte, error) {
	type registro struct {
		EventVersion string `json:"eventVersion"`
		EventSource  string `json:"eventSource"`
		AWSRegion    string `json:"awsRegion"`
		EventTime    string `json:"eventTime"`
		EventName    string `json:"eventName"`
		S3           struct {
			Bucket struct {
				Name string `json:"name"`
				ARN  string `json:"arn"`
			} `json:"bucket"`
			Object struct {
				Key  string `json:"key"`
				Size int    `json:"size"`
				ETag string `json:"eTag"`
			} `json:"object"`
		} `json:"s3"`
	}
	var r registro
	r.EventVersion, r.EventSource, r.AWSRegion = "2.1", "aws:s3", region
	r.EventTime = time.Now().UTC().Format(time.RFC3339)
	// El nombre del evento viaja SIN el `s3:` de delante. Es un detalle tonto y es de los que
	// hacen que tu `if` no entre nunca.
	r.EventName = strings.TrimPrefix(evento, "s3:")
	r.S3.Bucket.Name = bucket
	r.S3.Bucket.ARN = "arn:aws:s3:::" + bucket
	r.S3.Object.Key = clave
	r.S3.Object.Size = tam
	r.S3.Object.ETag = strings.Trim(etag, `"`)
	return json.Marshal(map[string]any{"Records": []registro{r}})
}

// ── las rutas del plano de datos ─────────────────────────────────────────

type fichaMapeo struct {
	UUID           string `json:"UUID"`
	EventSourceArn string `json:"EventSourceArn"`
	FunctionArn    string `json:"FunctionArn"`
	BatchSize      int    `json:"BatchSize"`
	State          string `json:"State"`
}

func (d *Disparadores) ficha(m *Mapeo) fichaMapeo {
	return fichaMapeo{UUID: m.UUID, EventSourceArn: arnCola(m.Cola, d.region),
		FunctionArn: arnDe(m.Funcion), BatchSize: m.Lote, State: m.Estado}
}

// Rutas se enchufa al mux de Lambda: los disparadores son parte de su API.
func (s *ServicioLambda) rutasDisparadores(mux *http.ServeMux) {
	mux.HandleFunc("POST /2015-03-31/event-source-mappings", s.crearMapeo)
	mux.HandleFunc("GET /2015-03-31/event-source-mappings", s.listarMapeos)
	mux.HandleFunc("DELETE /2015-03-31/event-source-mappings/{uuid}", s.borrarMapeo)
}

func (s *ServicioLambda) crearMapeo(w http.ResponseWriter, r *http.Request) {
	if s.disp == nil {
		s.fallo(w, 501, "ServiceException", "esta nubecita arranco sin disparadores")
		return
	}
	var c struct {
		EventSourceArn string `json:"EventSourceArn"`
		FunctionName   string `json:"FunctionName"`
		BatchSize      int    `json:"BatchSize"`
		Enabled        *bool  `json:"Enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		s.fallo(w, 400, "InvalidParameterValueException", err.Error())
		return
	}
	s.registro("POST /2015-03-31/event-source-mappings [%s -> %s]", c.EventSourceArn,
		c.FunctionName)

	activo := c.Enabled == nil || *c.Enabled
	m, err := s.disp.Crear(c.FunctionName, nombreDeARN(c.EventSourceArn), c.BatchSize, activo)
	if err != nil {
		s.fallo(w, 400, "InvalidParameterValueException", err.Error())
		return
	}
	s.json(w, 202, s.disp.ficha(m))
}

func (s *ServicioLambda) listarMapeos(w http.ResponseWriter, r *http.Request) {
	s.registro("GET /2015-03-31/event-source-mappings")
	fichas := []fichaMapeo{}
	if s.disp != nil {
		for _, m := range s.disp.Listar() {
			fichas = append(fichas, s.disp.ficha(m))
		}
	}
	s.json(w, 200, map[string]any{"EventSourceMappings": fichas})
}

func (s *ServicioLambda) borrarMapeo(w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("uuid")
	s.registro("DELETE /2015-03-31/event-source-mappings/%s", uuid)
	if s.disp == nil {
		s.fallo(w, 404, "ResourceNotFoundException", "no hay disparadores")
		return
	}
	m, ok := s.disp.Borrar(uuid)
	if !ok {
		s.fallo(w, 404, "ResourceNotFoundException", "no existe el mapeo "+uuid)
		return
	}
	f := s.disp.ficha(m)
	f.State = "Deleting"
	s.json(w, 202, f)
}
