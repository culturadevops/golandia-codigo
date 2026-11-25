package main

// La persistencia. Episodio 6, y el que cierra la serie.
//
// Hasta aqui nubecita no tocaba disco a proposito: matabas el proceso y no quedaba nada, que es
// justo lo que quieres de un doble de pruebas. Pero hay dos casos en los que eso estorba:
//
//	1. la maquina de desarrollo: arrancas veinte veces al dia y vuelves a subir los mismos
//	   ficheros de ejemplo;
//	2. la prueba que necesita empezar CON DATOS YA PUESTOS, y no con quince llamadas de
//	   preparacion que tardan mas que la prueba.
//
// Asi que se guarda. Y guardar estado abre una pregunta que casi nadie se hace en voz alta:
//
// 🥇 ¿QUE ES ESTADO Y QUE NO? Un objeto de S3 es estado. Un mensaje de SQS, tambien. Pero
// ¿el rato que a un mensaje le queda de estar invisible? ¿La invocacion que un runtime tenia en
// la mano cuando mataste el proceso? Eso no es estado: es una CONVERSACION a medias, y una
// conversacion a medias no se guarda, se abandona. Ver `Exportar` de las colas.
//
// Dos banderas, y hacen cosas distintas a proposito:
//
//	-datos   estado.json    lee al arrancar Y escribe cuando algo cambia
//	-semilla semilla.json   lee al arrancar y NO escribe nunca (para pruebas)

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// versionInstantanea sube cuando cambia el formato del fichero.
//
// ⚠️ Y esto no es burocracia. Sin un numero de version, un fichero viejo se carga a medias: los
// campos que existen entran y los que faltan se quedan en su valor cero, sin decir nada. Una
// cola sin visibilidad, con cero segundos, entrega el mismo mensaje en bucle — y nadie va a
// sospechar del fichero de estado.
const versionInstantanea = 1

var errVersionDesconocida = errors.New("no se de que version es este fichero de estado")

// ── lo que se guarda ─────────────────────────────────────────────────────

type Instantanea struct {
	Version   int        `json:"version"`
	Creada    time.Time  `json:"creada"`
	Buckets   []bucketG  `json:"buckets"`
	Tablas    []tablaG   `json:"tablas"`
	Colas     []colaG    `json:"colas"`
	Funciones []funcionG `json:"funciones"`
}

type objetoG struct {
	Clave string `json:"clave"`
	// ⚠️ Un []byte en JSON no sale como texto: sale en BASE64. Asi que el fichero no se lee a
	// simple vista y ocupa un tercio mas de lo que ocupaba el contenido. Es el precio de meter
	// bytes crudos en un formato de texto, y conviene saberlo antes de abrir el fichero y
	// pensar que se ha corrompido.
	Datos      []byte    `json:"datos"`
	Tipo       string    `json:"tipo"`
	Modificado time.Time `json:"modificado"`
	// ⛔ Ni ETag ni Tam. Los dos se calculan a partir de Datos, y guardar algo que se puede
	// recalcular es guardarse la oportunidad de que un dia no cuadren.
}

type bucketG struct {
	Nombre  string    `json:"nombre"`
	Creado  time.Time `json:"creado"`
	Objetos []objetoG `json:"objetos"`
}

type claveG struct {
	Nombre string `json:"nombre"`
	Tipo   string `json:"tipo"`
}

type tablaG struct {
	Nombre    string     `json:"nombre"`
	PK        claveG     `json:"pk"`
	SK        *claveG    `json:"sk,omitempty"`
	Creada    time.Time  `json:"creada"`
	Elementos []Elemento `json:"elementos"`
	// ⛔ Tampoco se guarda el mapa de claves compuestas: se vuelve a construir al cargar, con
	// el mismo codigo que lo construye al escribir. Un indice guardado es un indice que se
	// puede desincronizar.
}

type mensajeG struct {
	ID       string    `json:"id"`
	Cuerpo   string    `json:"cuerpo"`
	Recibido int       `json:"recibido"`
	Enviado  time.Time `json:"enviado"`
	// 🥇 ⛔ NO se guardan `VisibleEn` ni `Recibo`, y esta es LA decision del episodio.
	//
	// Un mensaje invisible lo esta porque alguien lo tiene en la mano ahora mismo. Si matas el
	// proceso, ese alguien ya no existe y su recibo no vale nada. Guardar «le quedan 22
	// segundos de invisibilidad» seria guardar una promesa hecha a alguien que se fue.
	//
	// Asi que al restaurar, TODO mensaje vuelve visible. Es la misma regla que aplica SQS
	// cuando un consumidor se cae: si nadie lo borro, vuelve a la cola.
}

type colaG struct {
	Nombre string `json:"nombre"`
	// La visibilidad como texto ("30s") y no como nanosegundos. El fichero se abre a mano mas
	// veces de las que uno cree.
	Visibilidad string     `json:"visibilidad"`
	MaxRecibos  int        `json:"maxRecibos,omitempty"`
	ColaMuertos string     `json:"colaMuertos,omitempty"`
	Siguiente   int        `json:"siguiente"`
	Mensajes    []mensajeG `json:"mensajes"`
}

type funcionG struct {
	Nombre    string            `json:"nombre"`
	Imagen    string            `json:"imagen,omitempty"`
	Entorno   map[string]string `json:"entorno,omitempty"`
	Runtime   string            `json:"runtime"`
	Manejador string            `json:"manejador"`
	Memoria   int32             `json:"memoria"`
	Segundos  int32             `json:"segundos"`
	Creada    time.Time         `json:"creada"`
	// ⛔ Ni los canales ni las invocaciones en vuelo. Un canal no es un dato: es un sitio donde
	// dos gorrutinas se esperan, y al arrancar de nuevo no hay nadie esperando.
}

// ordenarPor deja una lista en orden estable por una clave de texto.
//
// 🥇 ⚠️ Y no es cosmetica: el recorrido de un mapa en Go es ALEATORIO a proposito. Sin ordenar,
// dos exportaciones del MISMO estado salen con los elementos en distinto orden, el JSON sale
// distinto, y la comparacion de «¿ha cambiado algo?» dice que si SIEMPRE. Resultado: el fichero
// se reescribe entero en cada peticion, incluso sin tocar nada.
func ordenarPor[T any](s []T, clave func(T) string) {
	sort.Slice(s, func(i, j int) bool { return clave(s[i]) < clave(s[j]) })
}

// ── exportar ─────────────────────────────────────────────────────────────

func (a *Almacen) Exportar() []bucketG {
	a.mu.RLock()
	defer a.mu.RUnlock()
	fuera := make([]bucketG, 0, len(a.buckets))
	for _, b := range a.buckets {
		g := bucketG{Nombre: b.Nombre, Creado: b.Creado,
			Objetos: make([]objetoG, 0, len(b.objetos))}
		for _, o := range b.objetos {
			g.Objetos = append(g.Objetos, objetoG{Clave: o.Clave, Datos: o.Datos,
				Tipo: o.Tipo, Modificado: o.Modificado})
		}
		ordenarPor(g.Objetos, func(x objetoG) string { return x.Clave })
		fuera = append(fuera, g)
	}
	ordenarPor(fuera, func(x bucketG) string { return x.Nombre })
	return fuera
}

func (ts *Tablas) Exportar() []tablaG {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	fuera := make([]tablaG, 0, len(ts.tablas))
	for _, t := range ts.tablas {
		g := tablaG{Nombre: t.Nombre, Creada: t.Creada,
			PK:        claveG{t.PK.Nombre, t.PK.Tipo},
			Elementos: make([]Elemento, 0, len(t.elementos))}
		if t.SK != nil {
			g.SK = &claveG{t.SK.Nombre, t.SK.Tipo}
		}
		for _, e := range t.elementos {
			g.Elementos = append(g.Elementos, e)
		}
		fuera = append(fuera, g)
	}
	ordenarPor(fuera, func(x tablaG) string { return x.Nombre })
	return fuera
}

func (cs *Colas) Exportar() []colaG {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	fuera := make([]colaG, 0, len(cs.colas))
	for _, c := range cs.colas {
		g := colaG{Nombre: c.Nombre, Visibilidad: c.Visibilidad.String(),
			MaxRecibos: c.MaxRecibos, ColaMuertos: c.ColaMuertos, Siguiente: c.siguiente,
			Mensajes: make([]mensajeG, 0, len(c.mensajes))}
		for _, m := range c.mensajes {
			g.Mensajes = append(g.Mensajes, mensajeG{ID: m.ID, Cuerpo: m.Cuerpo,
				Recibido: m.Recibido, Enviado: m.Enviado})
		}
		fuera = append(fuera, g)
	}
	ordenarPor(fuera, func(x colaG) string { return x.Nombre })
	return fuera
}

func (f *Funciones) Exportar() []funcionG {
	f.mu.Lock()
	defer f.mu.Unlock()
	fuera := make([]funcionG, 0, len(f.porNom))
	for _, fn := range f.porNom {
		fuera = append(fuera, funcionG{Nombre: fn.Nombre, Imagen: fn.Imagen,
			Entorno: fn.Entorno, Runtime: fn.Runtime, Manejador: fn.Manejador, Memoria: fn.Memoria,
			Segundos: fn.Segundos, Creada: fn.Creada})
	}
	ordenarPor(fuera, func(x funcionG) string { return x.Nombre })
	return fuera
}

// ── importar ─────────────────────────────────────────────────────────────

func (a *Almacen) Importar(gs []bucketG) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.buckets = map[string]*Bucket{}
	for _, g := range gs {
		b := &Bucket{Nombre: g.Nombre, Creado: g.Creado, objetos: map[string]*Objeto{}}
		for _, o := range g.Objetos {
			suma := md5.Sum(o.Datos)
			b.objetos[o.Clave] = &Objeto{Clave: o.Clave, Datos: o.Datos, Tipo: o.Tipo,
				// El ETag y el tamaño se RECALCULAN. Si el fichero mintiera, aqui deja de
				// mentir.
				ETag: `"` + hex.EncodeToString(suma[:]) + `"`,
				Tam:  len(o.Datos), Modificado: o.Modificado}
		}
		a.buckets[g.Nombre] = b
	}
}

func (ts *Tablas) Importar(gs []tablaG) error {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	ts.tablas = map[string]*Tabla{}
	for _, g := range gs {
		t := &Tabla{Nombre: g.Nombre, Creada: g.Creada,
			PK: Clave{g.PK.Nombre, g.PK.Tipo}, elementos: map[string]Elemento{}}
		if g.SK != nil {
			t.SK = &Clave{g.SK.Nombre, g.SK.Tipo}
		}
		for _, e := range g.Elementos {
			// La clave compuesta se vuelve a calcular con el MISMO codigo de siempre. Si un
			// elemento del fichero no tiene la clave que dice la tabla, se entera aqui y no
			// tres consultas mas tarde.
			k, err := t.claveDe(e)
			if err != nil {
				return fmt.Errorf("tabla %s: %w", g.Nombre, err)
			}
			t.elementos[k] = e
		}
		ts.tablas[g.Nombre] = t
	}
	return nil
}

func (cs *Colas) Importar(gs []colaG) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.colas = map[string]*Cola{}
	for _, g := range gs {
		vis, err := time.ParseDuration(g.Visibilidad)
		if err != nil {
			return fmt.Errorf("cola %s: la visibilidad %q no es una duracion: %w",
				g.Nombre, g.Visibilidad, err)
		}
		c := &Cola{Nombre: g.Nombre, Visibilidad: vis, MaxRecibos: g.MaxRecibos,
			ColaMuertos: g.ColaMuertos, siguiente: g.Siguiente}
		for _, m := range g.Mensajes {
			suma := md5.Sum([]byte(m.Cuerpo))
			c.mensajes = append(c.mensajes, &Mensaje{
				ID: m.ID, Cuerpo: m.Cuerpo, MD5: hex.EncodeToString(suma[:]),
				Recibido: m.Recibido, Enviado: m.Enviado,
				// 🥇 Visible desde ya, y sin recibo. Ver el comentario de `mensajeG`.
				VisibleEn: time.Time{}, Recibo: "",
			})
		}
		cs.colas[g.Nombre] = c
	}
	return nil
}

func (f *Funciones) Importar(gs []funcionG) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.porNom = map[string]*Funcion{}
	f.enVuelo = map[string]*Invocacion{}
	for _, g := range gs {
		f.porNom[g.Nombre] = &Funcion{Nombre: g.Nombre, Imagen: g.Imagen,
			Entorno: g.Entorno, Runtime: g.Runtime, Manejador: g.Manejador, Memoria: g.Memoria,
			Segundos: g.Segundos, Creada: g.Creada,
			// El canal se crea nuevo. Es lo unico que NO puede venir de un fichero.
			espera: make(chan *Invocacion)}
	}
}

// ── la guarda: quien escribe, y cuando ───────────────────────────────────

// Guarda escribe el estado a disco. Y hace tres cosas que parecen detalles y no lo son:
// agrupa, compara y escribe entero.
type Guarda struct {
	ruta     string
	cada     time.Duration
	reunir   func() Instantanea
	registro func(formato string, args ...any)

	// ⚠️ Buffer 1, y otra vez por el mismo motivo que en Lambda: es un AVISO, no una cola. Si
	// llegan mil cambios mientras se esta escribiendo, el aviso pendiente vale por los mil.
	aviso chan struct{}
	hecho chan struct{}

	ultimo []byte // lo ultimo que se escribio: si no ha cambiado, no se reescribe
	veces  int    // cuantas escrituras de verdad ha habido (lo miran las pruebas)
	saltos int    // cuantas veces se decidio no escribir
}

func NuevaGuarda(ruta string, cada time.Duration, reunir func() Instantanea,
	registro func(string, ...any)) *Guarda {
	if registro == nil {
		registro = func(string, ...any) {}
	}
	return &Guarda{ruta: ruta, cada: cada, reunir: reunir, registro: registro,
		aviso: make(chan struct{}, 1), hecho: make(chan struct{})}
}

// Cambio avisa de que puede haber algo que guardar. No bloquea nunca.
func (g *Guarda) Cambio() {
	select {
	case g.aviso <- struct{}{}:
	default: // ya habia un aviso sin atender: uno vale por todos
	}
}

// Envolver marca un cambio DESPUES de cada peticion.
//
// 🥇 Y aqui esta la decision que ahorra la mitad de los errores de este fichero: no se intenta
// adivinar que operaciones escriben. Se avisa siempre, y quien escribe compara el resultado con
// lo ultimo que guardo; si es igual, no escribe.
//
// Adivinar parece mas listo y es una trampa: `ReceiveMessage` de SQS suena a lectura y MODIFICA
// el estado —marca el mensaje invisible y le cambia el recibo—. Una lista de «operaciones que
// escriben» se equivoca ahi, y se equivoca en silencio.
func (g *Guarda) Envolver(siguiente http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		siguiente.ServeHTTP(w, r)
		g.Cambio()
	})
}

// Arrancar lanza al que escribe. Devuelve cuando el contexto muere Y el ultimo estado esta en
// disco: por eso `Esperar` existe.
func (g *Guarda) Arrancar(ctx context.Context) {
	go func() {
		defer close(g.hecho)
		for {
			select {
			case <-ctx.Done():
				g.Escribir() // la ultima, al salir
				return
			case <-g.aviso:
				// Agrupar: se espera un poco por si vienen mas cambios. Con esto, una rafaga
				// de mil peticiones es UNA escritura, no mil.
				t := time.NewTimer(g.cada)
				select {
				case <-t.C:
				case <-ctx.Done():
					t.Stop()
					g.Escribir()
					return
				}
				g.Escribir()
			}
		}
	}()
}

func (g *Guarda) Esperar() { <-g.hecho }

// Escribir vuelca el estado, si ha cambiado.
func (g *Guarda) Escribir() {
	inst := g.reunir()
	inst.Version = versionInstantanea
	inst.Creada = time.Now().UTC()

	crudo, err := json.MarshalIndent(inst, "", "  ")
	if err != nil {
		g.registro("no puedo serializar el estado: %v", err)
		return
	}

	// ⚠️ La comparacion NO puede incluir la fecha, o cambiaria siempre y escribiriamos cada
	// vez. Asi que se compara con la fecha fuera.
	inst.Creada = time.Time{}
	sinFecha, _ := json.Marshal(inst)
	if bytes.Equal(sinFecha, g.ultimo) {
		g.saltos++
		return
	}

	if err := escribirEntero(g.ruta, crudo); err != nil {
		g.registro("no puedo guardar el estado en %s: %v", g.ruta, err)
		return
	}
	g.ultimo = sinFecha
	g.veces++
	g.registro("estado guardado en %s (%d bytes)", g.ruta, len(crudo))
}

// escribirEntero escribe a un temporal y luego renombra.
//
// ⚠️ Escribir «encima» del fichero bueno tiene un momento en el que el fichero existe y esta a
// medias. Si el proceso muere ahi —o si alguien lo lee justo entonces— el estado que tenias se
// ha convertido en basura. Renombrar dentro del mismo sistema de ficheros es atomico: o esta el
// viejo entero, o esta el nuevo entero. Nunca medio.
func escribirEntero(ruta string, datos []byte) error {
	dir := filepath.Dir(ruta)
	tmp, err := os.CreateTemp(dir, filepath.Base(ruta)+".*.tmp")
	if err != nil {
		return err
	}
	nombre := tmp.Name()
	if _, err := tmp.Write(datos); err != nil {
		tmp.Close()
		os.Remove(nombre)
		return err
	}
	// Sincronizar antes de renombrar: si no, el renombrado puede llegar al disco antes que el
	// contenido y quedarte un fichero nuevo y vacio.
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		os.Remove(nombre)
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(nombre)
		return err
	}
	if err := os.Rename(nombre, ruta); err != nil {
		os.Remove(nombre)
		return err
	}
	return nil
}

// ── cargar ───────────────────────────────────────────────────────────────

// Cargar lee una instantanea. Si el fichero no existe devuelve (nil, nil): arrancar vacio es
// normal la primera vez, no es un error.
func Cargar(ruta string) (*Instantanea, error) {
	crudo, err := os.ReadFile(ruta)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var inst Instantanea
	// ⚠️ DisallowUnknownFields: si el fichero trae un campo que no conocemos, es que lo escribio
	// una version mas nueva. Mejor negarse que cargar la mitad.
	dec := json.NewDecoder(bytes.NewReader(crudo))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&inst); err != nil {
		return nil, fmt.Errorf("%s no es un estado de nubecita: %w", ruta, err)
	}
	if inst.Version != versionInstantanea {
		return nil, fmt.Errorf("%w: %s dice version %d y yo entiendo la %d",
			errVersionDesconocida, ruta, inst.Version, versionInstantanea)
	}
	return &inst, nil
}

// Aplicar mete una instantanea en los almacenes que ya estan montados.
func Aplicar(inst *Instantanea, a *Almacen, ts *Tablas, cs *Colas, f *Funciones) error {
	if inst == nil {
		return nil
	}
	a.Importar(inst.Buckets)
	if err := ts.Importar(inst.Tablas); err != nil {
		return err
	}
	if err := cs.Importar(inst.Colas); err != nil {
		return err
	}
	f.Importar(inst.Funciones)
	return nil
}

// Reunir arma la instantanea de todo lo que hay ahora mismo.
func Reunir(a *Almacen, ts *Tablas, cs *Colas, f *Funciones) Instantanea {
	return Instantanea{
		Buckets: a.Exportar(), Tablas: ts.Exportar(),
		Colas: cs.Exportar(), Funciones: f.Exportar(),
	}
}
