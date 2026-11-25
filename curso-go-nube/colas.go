package main

// El estado de SQS: colas y mensajes. Y aqui vive el concepto que hace a SQS distinto de
// cualquier lista de la que hayas sacado elementos: la VISIBILIDAD.
//
// Un mensaje no se saca de la cola cuando lo recibes. Se vuelve INVISIBLE durante un rato. Si
// dentro de ese rato lo borras, desaparece. Si no lo borras —porque tu proceso murio, o tardo
// demasiado— vuelve a aparecer y se lo lleva otro.
//
// De ahi salen las dos consecuencias que definen como se escribe un consumidor:
//
//	1. La entrega es AL MENOS UNA VEZ, nunca exactamente una. Tu consumidor tiene que ser
//	   idempotente: procesar dos veces el mismo mensaje no puede cobrar dos veces.
//	2. Un mensaje que falla siempre se reintenta para siempre. Por eso existe la cola de
//	   mensajes muertos: a la N-esima entrega, en vez de darlo otra vez, se aparta.

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"sort"
	"sync"
	"time"
)

type Mensaje struct {
	ID        string
	Cuerpo    string
	MD5       string
	Recibido  int       // cuantas veces se ha entregado ya
	VisibleEn time.Time // antes de este instante, no se entrega a nadie
	Recibo    string    // el ReceiptHandle vigente: cambia en CADA entrega
	Enviado   time.Time
}

type Cola struct {
	Nombre      string
	Visibilidad time.Duration
	MaxRecibos  int    // del RedrivePolicy: a partir de aqui, a la cola de muertos
	ColaMuertos string // nombre de la cola de muertos, si hay
	mensajes    []*Mensaje
	siguiente   int // para generar identificadores estables y legibles
}

type Colas struct {
	mu    sync.Mutex // aqui un Mutex normal: recibir MODIFICA (marca invisible), no solo lee
	colas map[string]*Cola
	reloj func() time.Time
	seq   int
}

// ⚠️ Aqui NO se guarda la direccion publica a proposito: la URL de la cola la arma quien atiende
// la peticion, con el Host que trajo el cliente. Guardarla al arrancar da URLs equivocadas en
// cuanto el servidor vive detras de un contenedor o de un proxy.
func NuevasColas() *Colas {
	return &Colas{colas: map[string]*Cola{}, reloj: time.Now}
}

const visibilidadPorDefecto = 30 * time.Second

// ── colas ────────────────────────────────────────────────────────────────

// Crear recibe la visibilidad como PUNTERO, y no es capricho: en SQS `0` es un valor legitimo
// («que reaparezca al instante») y hay que poder distinguirlo de «no me lo has dicho». Con un
// `time.Duration` a secas, `0` y «sin especificar» son el mismo valor y el cero se pierde.
func (cs *Colas) Crear(nombre string, visibilidad *time.Duration, maxRecibos int,
	colaMuertos string) *Cola {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	if c, ya := cs.colas[nombre]; ya {
		// SQS es idempotente al crear si los atributos coinciden. Aqui se simplifica: se
		// devuelve la que hay. Lo que NO se hace es fallar en silencio.
		return c
	}
	vis := visibilidadPorDefecto
	if visibilidad != nil {
		vis = *visibilidad
	}
	c := &Cola{Nombre: nombre, Visibilidad: vis, MaxRecibos: maxRecibos,
		ColaMuertos: colaMuertos}
	cs.colas[nombre] = c
	return c
}

func (cs *Colas) Buscar(nombre string) (*Cola, bool) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	c, ok := cs.colas[nombre]
	return c, ok
}

func (cs *Colas) Borrar(nombre string) bool {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	if _, ok := cs.colas[nombre]; !ok {
		return false
	}
	delete(cs.colas, nombre)
	return true
}

func (cs *Colas) Listar(prefijo string) []string {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	fuera := make([]string, 0, len(cs.colas))
	for n := range cs.colas {
		if prefijo == "" || len(n) >= len(prefijo) && n[:len(prefijo)] == prefijo {
			fuera = append(fuera, n)
		}
	}
	sort.Strings(fuera)
	return fuera
}

func (cs *Colas) Vaciar(nombre string) bool {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	c, ok := cs.colas[nombre]
	if !ok {
		return false
	}
	c.mensajes = nil
	return true
}

// ── mensajes ─────────────────────────────────────────────────────────────

func (cs *Colas) Enviar(nombre, cuerpo string) (*Mensaje, error) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	c, ok := cs.colas[nombre]
	if !ok {
		return nil, errNoExisteCola
	}
	suma := md5.Sum([]byte(cuerpo))
	c.siguiente++
	cs.seq++
	m := &Mensaje{
		ID:      fmt.Sprintf("%s-%d", nombre, c.siguiente),
		Cuerpo:  cuerpo,
		MD5:     hex.EncodeToString(suma[:]),
		Enviado: cs.reloj().UTC(),
		// VisibleEn en cero: visible desde ya. (Aqui iria el retraso de entrega si se
		// implementara DelaySeconds.)
	}
	c.mensajes = append(c.mensajes, m)
	return m, nil
}

// Recibir entrega hasta `cuantos` mensajes visibles y los esconde durante `visibilidad`.
//
// 🥇 Este metodo es el corazon de SQS, y hace TRES cosas que sorprenden:
//   - marca el mensaje como invisible en vez de sacarlo;
//   - genera un RECIBO NUEVO en cada entrega — el recibo no es el identificador del mensaje,
//     y el de la entrega anterior deja de valer;
//   - si el mensaje ya se entrego demasiadas veces, NO lo entrega: lo aparta a la cola de
//     muertos. Esa es toda la magia del "dead letter queue".
func (cs *Colas) Recibir(nombre string, cuantos int, visibilidad *time.Duration) ([]*Mensaje, error) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	c, ok := cs.colas[nombre]
	if !ok {
		return nil, errNoExisteCola
	}
	vis := c.Visibilidad
	if visibilidad != nil {
		vis = *visibilidad
	}
	ahora := cs.reloj().UTC()
	var salen []*Mensaje
	var quedan []*Mensaje
	for _, m := range c.mensajes {
		if len(salen) >= cuantos || m.VisibleEn.After(ahora) {
			quedan = append(quedan, m)
			continue
		}
		if c.MaxRecibos > 0 && m.Recibido+1 > c.MaxRecibos {
			if muertos, hay := cs.colas[c.ColaMuertos]; hay {
				m.VisibleEn = time.Time{}
				m.Recibido = 0
				m.Recibo = ""
				muertos.mensajes = append(muertos.mensajes, m)
			}
			// Si no hay cola de muertos configurada, el mensaje se descarta. Es lo que hace
			// SQS y es mejor que dar vueltas para siempre — pero por eso conviene tenerla.
			continue
		}
		m.Recibido++
		m.VisibleEn = ahora.Add(vis)
		cs.seq++
		m.Recibo = fmt.Sprintf("recibo-%d-%d", cs.seq, m.Recibido)
		salen = append(salen, m)
		quedan = append(quedan, m)
	}
	c.mensajes = quedan
	return salen, nil
}

// Quitar borra por RECIBO, no por identificador. Y devuelve si encontro algo: borrar con un
// recibo caducado no es un error en SQS, pero saberlo sirve para el registro.
func (cs *Colas) Quitar(nombre, recibo string) (bool, error) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	c, ok := cs.colas[nombre]
	if !ok {
		return false, errNoExisteCola
	}
	for i, m := range c.mensajes {
		if m.Recibo != "" && m.Recibo == recibo {
			c.mensajes = append(c.mensajes[:i], c.mensajes[i+1:]...)
			return true, nil
		}
	}
	return false, nil
}

// CambiarVisibilidad es lo que usa un consumidor que va a tardar mas de lo previsto: pide mas
// tiempo antes de que el mensaje reaparezca. Con 0, lo devuelve a la cola de inmediato.
func (cs *Colas) CambiarVisibilidad(nombre, recibo string, d time.Duration) (bool, error) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	c, ok := cs.colas[nombre]
	if !ok {
		return false, errNoExisteCola
	}
	for _, m := range c.mensajes {
		if m.Recibo != "" && m.Recibo == recibo {
			m.VisibleEn = cs.reloj().UTC().Add(d)
			return true, nil
		}
	}
	return false, nil
}

// Cuentas devuelve lo que SQS llama "aproximado", y aqui se ve por que lo llama asi: son dos
// numeros que dependen del instante exacto en que preguntas.
func (cs *Colas) Cuentas(nombre string) (visibles, enVuelo int, ok bool) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	c, hay := cs.colas[nombre]
	if !hay {
		return 0, 0, false
	}
	ahora := cs.reloj().UTC()
	for _, m := range c.mensajes {
		if m.VisibleEn.After(ahora) {
			enVuelo++
		} else {
			visibles++
		}
	}
	return visibles, enVuelo, true
}
