package main

// El registro de funciones de Lambda, y —lo interesante— el ENCUENTRO entre quien invoca y
// quien ejecuta.
//
// Esto es lo que casi nadie sabe de Lambda: cuando llamas a una funcion, tu peticion NO le
// llega a tu codigo. Le llega a Lambda, que la deja esperando hasta que un runtime la venga a
// buscar. Y el runtime la viene a buscar preguntando, en un bucle, "¿hay algo para mi?".
//
// O sea que hay DOS mitades que se tienen que encontrar:
//
//	quien invoca   →  deja una invocacion y se queda esperando la respuesta
//	el runtime     →  pide trabajo, lo hace, y devuelve el resultado
//
// En Go eso es un canal. Literalmente un canal sin buffer: un lado manda, el otro recibe, y
// ninguno sigue hasta que el otro esta. Se llama encuentro (rendezvous) y es la primitiva mas
// vieja de la concurrencia.
//
// 🥇 Y de aqui sale la explicacion del arranque en frio, que se suele contar como si fuera
// magia: el arranque en frio es el tiempo que quien invoca pasa esperando en ese canal porque
// todavia no hay NADIE al otro lado. No es una penalizacion que te cobran: es que el receptor
// aun no ha llegado.

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

var (
	errNoExisteFuncion = errors.New("la funcion no existe")
	errYaExisteFuncion = errors.New("la funcion ya existe")
)

// Resultado es lo que el runtime devuelve de una invocacion.
//
// ⚠️ `Fallo` no quiere decir "hubo un error de red". Quiere decir que TU CODIGO devolvio un
// error. Y eso, en Lambda, viaja con codigo HTTP 200. Ver `lambda.go`.
type Resultado struct {
	Cuerpo []byte
	Fallo  bool
}

// Invocacion es una llamada esperando a que alguien la ejecute.
type Invocacion struct {
	ID     string
	Evento []byte
	Limite time.Time

	// ⚠️ Con buffer 1, y esto NO es un detalle de estilo. Quien invoca puede irse antes de
	// tiempo (cortar la conexion, agotar su contexto). Si el canal fuera sin buffer, el
	// runtime se quedaria bloqueado para siempre intentando entregar una respuesta que ya no
	// espera nadie: una gorrutina perdida por cada invocacion abandonada, y el proceso
	// creciendo despacito hasta que se cae de madrugada.
	//
	// Con hueco para uno, quien contesta deja la respuesta y se va, la espere alguien o no.
	Respuesta chan Resultado
}

// Funcion es una funcion registrada. `espera` es el canal por el que se encuentran.
type Funcion struct {
	Nombre    string
	Runtime   string
	Manejador string
	Memoria   int32
	Segundos  int32
	Creada    time.Time
	// Imagen es el contenedor que sabe ejecutar esta funcion. Vacia si nadie nos lo dijo: en
	// ese caso el runtime lo levantas tu, como en el episodio 5.
	Imagen string
	// Entorno son las variables que se le ponen al contenedor. Es como se configura una Lambda
	// de verdad: el codigo no cambia entre entornos, cambian estas.
	Entorno map[string]string

	// ⚠️ Estos dos numeros existen solo para el lanzador (episodio 7), y son atomicos porque
	// los toca cualquier gorrutina que atienda una peticion. Un candado aqui seria un candado
	// en el camino caliente de TODAS las invocaciones.
	//
	//	libres  runtimes esperando trabajo AHORA MISMO. Si es cero, hay que levantar otro.
	//	ultima  cuando se invoco por ultima vez. Si hace mucho, sobran contenedores.
	libres atomic.Int32
	ultima atomic.Int64

	// ⚠️ Sin buffer, a proposito. Una invocacion no se "encola" en la funcion: se queda en la
	// mano de quien invoca hasta que un runtime la recoge. Si tuviera buffer, `Invoke`
	// devolveria antes de que nadie fuera a ejecutar nada, y estariamos mintiendo sobre
	// cuando empieza el trabajo.
	espera chan *Invocacion
}

// Funciones es el registro. Nada de esto toca disco: al matar el proceso, no queda rastro.
type Funciones struct {
	mu      sync.Mutex
	porNom  map[string]*Funcion
	enVuelo map[string]*Invocacion // por ID de peticion, mientras el runtime la trabaja
	sec     int64
	ahora   func() time.Time
}

func NuevasFunciones() *Funciones {
	return &Funciones{
		porNom:  map[string]*Funcion{},
		enVuelo: map[string]*Invocacion{},
		ahora:   func() time.Time { return time.Now().UTC() },
	}
}

func (f *Funciones) Crear(nombre, runtime, manejador string, memoria, segundos int32) (*Funcion, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, hay := f.porNom[nombre]; hay {
		return nil, errYaExisteFuncion
	}
	fn := &Funcion{
		Nombre: nombre, Runtime: runtime, Manejador: manejador,
		Memoria: memoria, Segundos: segundos, Creada: f.ahora(),
		espera: make(chan *Invocacion),
	}
	f.porNom[nombre] = fn
	return fn, nil
}

func (f *Funciones) Buscar(nombre string) (*Funcion, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	fn, hay := f.porNom[nombre]
	if !hay {
		return nil, errNoExisteFuncion
	}
	return fn, nil
}

func (f *Funciones) Listar() []*Funcion {
	f.mu.Lock()
	defer f.mu.Unlock()
	fuera := make([]*Funcion, 0, len(f.porNom))
	for _, fn := range f.porNom {
		fuera = append(fuera, fn)
	}
	return fuera
}

func (f *Funciones) Borrar(nombre string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, hay := f.porNom[nombre]; !hay {
		return errNoExisteFuncion
	}
	delete(f.porNom, nombre)
	return nil
}

// Invocar deja una invocacion en la funcion y espera la respuesta.
//
// 🥇 Las DOS esperas de esta funcion son las dos mitades del arranque en frio:
//
//	la primera  →  esperando a que un runtime venga a buscar trabajo
//	la segunda  →  esperando a que ese runtime termine
//
// Si tu funcion tarda 20 ms pero la llamada tarda 900, casi todo se fue en la primera.
func (f *Funciones) Invocar(ctx context.Context, fn *Funcion, evento []byte) (Resultado, error) {
	fn.ultima.Store(f.ahora().UnixNano())
	inv := &Invocacion{
		ID:        f.siguienteID(),
		Evento:    evento,
		Limite:    f.ahora().Add(time.Duration(fn.Segundos) * time.Second),
		Respuesta: make(chan Resultado, 1),
	}

	// ⚠️ Los dos `select` llevan `ctx.Done()`. Sin eso, un cliente que cuelga deja aqui una
	// gorrutina parada para siempre — y en un servidor eso no se ve, no da error y no se va.
	select {
	case fn.espera <- inv:
	case <-ctx.Done():
		return Resultado{}, ctx.Err()
	}

	select {
	case res := <-inv.Respuesta:
		return res, nil
	case <-ctx.Done():
		return Resultado{}, ctx.Err()
	}
}

// Pedir es la otra mitad: la llamada que hace el runtime para que le den trabajo. Se queda
// bloqueada —esto es el sondeo largo— hasta que llega una invocacion o el runtime se va.
//
// ⚠️ Fijate en que NO se toca el mutex mientras se espera en el canal. Bloquearse con el
// candado cogido pararia todo el registro: nadie podria ni crear una funcion ni listar. Es el
// fallo clasico de mezclar candados y canales, y aqui se evita porque el canal se saca del
// mapa ANTES de esperar.
func (f *Funciones) Pedir(ctx context.Context, fn *Funcion) (*Invocacion, error) {
	// Mientras esta gorrutina espera en el canal, este runtime esta LIBRE. En cuanto recoge
	// trabajo, deja de estarlo. El lanzador mira este numero para saber si hace falta otro
	// contenedor, y no hace falta preguntarle nada a Docker.
	fn.libres.Add(1)
	defer fn.libres.Add(-1)

	select {
	case inv := <-fn.espera:
		f.mu.Lock()
		f.enVuelo[inv.ID] = inv
		f.mu.Unlock()
		return inv, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// Contestar entrega el resultado de una invocacion que estaba en vuelo.
//
// Devuelve false si ese ID no existe o si ya se contesto. Que no exista es normal: el runtime
// puede contestar tarde, cuando la invocacion ya vencio.
func (f *Funciones) Contestar(id string, res Resultado) bool {
	f.mu.Lock()
	inv, hay := f.enVuelo[id]
	delete(f.enVuelo, id)
	f.mu.Unlock()
	if !hay {
		return false
	}
	inv.Respuesta <- res // no bloquea: el canal tiene hueco para uno
	return true
}

// Libres son los runtimes de esta funcion que estan pidiendo trabajo ahora mismo.
func (fn *Funcion) Libres() int { return int(fn.libres.Load()) }

// Ultima es cuando se invoco por ultima vez. El cero significa «nunca».
func (fn *Funcion) Ultima() time.Time {
	n := fn.ultima.Load()
	if n == 0 {
		return time.Time{}
	}
	return time.Unix(0, n).UTC()
}

func (f *Funciones) siguienteID() string {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.sec++
	// Un identificador con la pinta de los de AWS. No es un UUID de verdad y no hace falta:
	// solo tiene que ser distinto dentro de este proceso.
	return "00000000-0000-4000-8000-" + pad12(f.sec)
}

func pad12(n int64) string {
	d := []byte("000000000000")
	for i := len(d) - 1; i >= 0 && n > 0; i-- {
		d[i] = byte('0' + n%10)
		n /= 10
	}
	return string(d)
}
