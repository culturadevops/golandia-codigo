package main

// El estado de SNS: temas y suscripciones. Episodio 9.
//
// Una cola y un tema parecen lo mismo y no lo son, y la diferencia cabe en una frase:
//
//	COLA   el mensaje se lo lleva UNO. Quien llega primero.
//	TEMA   el mensaje se lo llevan TODOS los que estaban apuntados.
//
// De ahi sale todo lo demas. Una cola guarda; un tema NO guarda nada. Si publicas en un tema sin
// suscriptores, el mensaje no se pierde: es que nunca existio. Y por eso el patron de verdad no
// es «un tema» ni «una cola», sino un tema QUE REPARTE A COLAS: el tema hace la copia, y cada
// cola guarda la suya hasta que su consumidor pueda.
//
// ⚠️ Y la trampa que se lleva a todo el mundo la primera vez: cuando un tema entrega a una cola,
// tu mensaje llega ENVUELTO en otro JSON. Tu consumidor lee el cuerpo del mensaje de la cola y
// se encuentra un sobre, no la carta. Se arregla con `RawMessageDelivery`, y hay que saberlo.

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

var (
	errNoExisteTema  = errors.New("el tema no existe")
	errNoExisteSuscr = errors.New("la suscripcion no existe")
)

// Suscripcion es alguien apuntado a un tema.
type Suscripcion struct {
	ARN      string
	TemaARN  string
	Protocol string // "sqs" es el unico que se entrega de verdad aqui
	Destino  string // el ARN de la cola
	// 🥇 Si esta en false —que es lo normal— el mensaje llega ENVUELTO. Ver `Publicar`.
	Crudo bool
	// Filtro por atributos. Vacio quiere decir «todo».
	Filtro map[string][]string
}

type Tema struct {
	Nombre string
	ARN    string
	Creado time.Time
}

type Temas struct {
	mu     sync.Mutex
	temas  map[string]*Tema        // por ARN
	suscr  map[string]*Suscripcion // por ARN de suscripcion
	region string
	cuenta string
	sec    int
	reloj  func() time.Time
}

func NuevosTemas() *Temas {
	return &Temas{temas: map[string]*Tema{}, suscr: map[string]*Suscripcion{},
		region: "us-east-1", cuenta: "000000000000", reloj: func() time.Time { return time.Now().UTC() }}
}

func (ts *Temas) arnTema(nombre string) string {
	return "arn:aws:sns:" + ts.region + ":" + ts.cuenta + ":" + nombre
}

// Crear es idempotente, como en SNS de verdad: crear dos veces el mismo tema devuelve el mismo.
func (ts *Temas) Crear(nombre string) *Tema {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	arn := ts.arnTema(nombre)
	if t, ya := ts.temas[arn]; ya {
		return t
	}
	t := &Tema{Nombre: nombre, ARN: arn, Creado: ts.reloj()}
	ts.temas[arn] = t
	return t
}

func (ts *Temas) Buscar(arn string) (*Tema, bool) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	t, ok := ts.temas[arn]
	return t, ok
}

func (ts *Temas) Listar() []*Tema {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	fuera := make([]*Tema, 0, len(ts.temas))
	for _, t := range ts.temas {
		fuera = append(fuera, t)
	}
	sort.Slice(fuera, func(i, j int) bool { return fuera[i].ARN < fuera[j].ARN })
	return fuera
}

// Borrar quita el tema Y sus suscripciones. Dejarlas seria dejar apuntada gente a una lista que
// ya no existe.
func (ts *Temas) Borrar(arn string) bool {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	if _, hay := ts.temas[arn]; !hay {
		return false
	}
	delete(ts.temas, arn)
	for k, s := range ts.suscr {
		if s.TemaARN == arn {
			delete(ts.suscr, k)
		}
	}
	return true
}

func (ts *Temas) Suscribir(temaARN, protocolo, destino string) (*Suscripcion, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	if _, hay := ts.temas[temaARN]; !hay {
		return nil, errNoExisteTema
	}
	ts.sec++
	s := &Suscripcion{
		ARN:     fmt.Sprintf("%s:%08d-0000-4000-8000-000000000000", temaARN, ts.sec),
		TemaARN: temaARN, Protocol: protocolo, Destino: destino,
	}
	ts.suscr[s.ARN] = s
	return s, nil
}

func (ts *Temas) BuscarSuscripcion(arn string) (*Suscripcion, bool) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	s, ok := ts.suscr[arn]
	return s, ok
}

func (ts *Temas) Desuscribir(arn string) bool {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	if _, hay := ts.suscr[arn]; !hay {
		return false
	}
	delete(ts.suscr, arn)
	return true
}

// Suscripciones devuelve las de un tema, o todas si el ARN va vacio.
func (ts *Temas) Suscripciones(temaARN string) []*Suscripcion {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	fuera := []*Suscripcion{}
	for _, s := range ts.suscr {
		if temaARN == "" || s.TemaARN == temaARN {
			fuera = append(fuera, s)
		}
	}
	sort.Slice(fuera, func(i, j int) bool { return fuera[i].ARN < fuera[j].ARN })
	return fuera
}

// PonerAtributo cambia una opcion de la suscripcion. Las dos que importan son estas.
func (ts *Temas) PonerAtributo(arn, nombre, valor string) error {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	s, hay := ts.suscr[arn]
	if !hay {
		return errNoExisteSuscr
	}
	switch nombre {
	case "RawMessageDelivery":
		s.Crudo = strings.EqualFold(valor, "true")
	case "FilterPolicy":
		f, err := leerFiltro(valor)
		if err != nil {
			return err
		}
		s.Filtro = f
	}
	return nil
}

// leerFiltro entiende la forma sencilla de la politica de filtrado:
//
//	{"tipo": ["pedido", "devolucion"], "pais": ["ES"]}
//
// ⛔ La de verdad admite mucho mas (numeric, prefix, anything-but...). Aqui solo la lista de
// valores exactos, que es la que se usa el noventa por ciento de las veces.
func leerFiltro(crudo string) (map[string][]string, error) {
	if strings.TrimSpace(crudo) == "" {
		return nil, nil
	}
	var bruto map[string][]any
	if err := json.Unmarshal([]byte(crudo), &bruto); err != nil {
		return nil, fmt.Errorf("la politica de filtrado no es un JSON que entienda: %w", err)
	}
	fuera := map[string][]string{}
	for k, vs := range bruto {
		for _, v := range vs {
			if t, ok := v.(string); ok {
				fuera[k] = append(fuera[k], t)
			}
		}
	}
	return fuera, nil
}

// AtributoMensaje es un atributo tal como viaja: con su tipo al lado.
type AtributoMensaje struct {
	Tipo  string
	Valor string
}

// pasaElFiltro decide si esta suscripcion quiere este mensaje.
//
// 🥇 Y fijate DONDE se decide: en la suscripcion, no en quien publica. Quien publica no sabe ni
// cuantos hay apuntados. Cada suscriptor dice de que quiere enterarse, y el tema descarta antes
// de entregar. Por eso filtrar aqui es gratis y filtrar dentro de tu consumidor no lo es: lo
// que no pasa el filtro no llega a tu cola, y lo que no llega a tu cola no se factura.
func pasaElFiltro(s *Suscripcion, attrs map[string]AtributoMensaje) bool {
	if len(s.Filtro) == 0 {
		return true
	}
	for clave, admitidos := range s.Filtro {
		a, hay := attrs[clave]
		if !hay {
			// ⚠️ Si el mensaje NO trae el atributo, NO pasa. No es «da igual»: es que no
			// coincide. Es el motivo numero uno de «mi filtro no deja pasar nada».
			return false
		}
		ok := false
		for _, v := range admitidos {
			if a.Valor == v {
				ok = true
				break
			}
		}
		if !ok {
			return false
		}
	}
	return true
}

// sobre es el JSON que SNS pone alrededor de tu mensaje cuando entrega a una cola.
//
// 🥇 ESTA es la trampa del episodio. Tu consumidor lee el cuerpo del mensaje de la cola y se
// encuentra esto, no lo que publicaste. Lo que publicaste esta dentro, en `Message`, y ademas
// como CADENA — asi que si publicaste un JSON, tienes que deserializar dos veces.
type sobre struct {
	Type             string                   `json:"Type"`
	MessageId        string                   `json:"MessageId"`
	TopicArn         string                   `json:"TopicArn"`
	Subject          string                   `json:"Subject,omitempty"`
	Message          string                   `json:"Message"`
	Timestamp        string                   `json:"Timestamp"`
	MessageAttribute map[string]sobreAtributo `json:"MessageAttributes,omitempty"`
}

type sobreAtributo struct {
	Type  string `json:"Type"`
	Value string `json:"Value"`
}

// Entrega es lo que hay que hacer con un mensaje para un suscriptor.
type Entrega struct {
	Cola   string // nombre de la cola destino
	Cuerpo string
}

// Publicar decide QUE se entrega y A QUIEN. No entrega: eso lo hace quien tenga las colas
// delante, para que este fichero no dependa de SQS.
//
// 🥇 Y aqui esta la diferencia con una cola, en una linea: esto devuelve una LISTA. Un mensaje
// entra, y salen tantas copias como suscriptores lo quieran. Si no hay ninguno, sale una lista
// vacia — y ese mensaje no se ha perdido: es que nunca existio para nadie.
func (ts *Temas) Publicar(temaARN, mensaje, asunto string,
	attrs map[string]AtributoMensaje) (string, []Entrega, error) {
	ts.mu.Lock()
	t, hay := ts.temas[temaARN]
	if !hay {
		ts.mu.Unlock()
		return "", nil, errNoExisteTema
	}
	ts.sec++
	id := fmt.Sprintf("%s-%d", t.Nombre, ts.sec)
	var suscr []*Suscripcion
	for _, s := range ts.suscr {
		if s.TemaARN == temaARN {
			suscr = append(suscr, s)
		}
	}
	ahora := ts.reloj()
	ts.mu.Unlock()

	sort.Slice(suscr, func(i, j int) bool { return suscr[i].ARN < suscr[j].ARN })

	entregas := []Entrega{}
	for _, s := range suscr {
		if !pasaElFiltro(s, attrs) {
			continue
		}
		if !strings.EqualFold(s.Protocol, "sqs") {
			// ⛔ Los demas protocolos (correo, HTTP, SMS) no se entregan aqui. Se acepta la
			// suscripcion y se dice en voz alta que no va a llegar nada.
			continue
		}
		cuerpo := mensaje
		if !s.Crudo {
			cuerpo = envolver(id, temaARN, asunto, mensaje, attrs, ahora)
		}
		entregas = append(entregas, Entrega{Cola: nombreDeARN(s.Destino), Cuerpo: cuerpo})
	}
	return id, entregas, nil
}

func envolver(id, temaARN, asunto, mensaje string, attrs map[string]AtributoMensaje,
	ahora time.Time) string {
	s := sobre{
		Type: "Notification", MessageId: id, TopicArn: temaARN, Subject: asunto,
		Message: mensaje, Timestamp: ahora.Format("2006-01-02T15:04:05.000Z"),
	}
	if len(attrs) > 0 {
		s.MessageAttribute = map[string]sobreAtributo{}
		for k, v := range attrs {
			s.MessageAttribute[k] = sobreAtributo{Type: v.Tipo, Value: v.Valor}
		}
	}
	crudo, err := json.Marshal(s)
	if err != nil {
		return mensaje
	}
	return string(crudo)
}
