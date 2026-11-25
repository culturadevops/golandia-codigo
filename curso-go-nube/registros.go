package main

// CloudWatch Logs: el sitio donde acaba lo que imprime tu funcion.
//
// Esto cierra un hueco que estaba DECLARADO desde el episodio 7: «lo que imprima la funcion se
// queda en su contenedor». Se quedaba ahi porque nadie lo sacaba. Aqui esta el que lo saca.
//
// 🥇 Y lo primero que se aprende escribiendolo es que un registro NO es una cola, aunque las dos
// guarden cosas en orden y las dos se lean por el principio:
//
//	una cola      leer SACA el mensaje. Lo lee UNO. El puntero lo lleva el servidor.
//	un registro   leer no borra nada. Lo leen TODOS. El puntero lo lleva el LECTOR.
//
// De esa diferencia sale todo lo demas: por eso un registro tiene `nextToken` y una cola tiene
// `ReceiptHandle`; por eso dos consumidores de una cola se reparten el trabajo y dos lectores de
// un registro ven exactamente lo mismo.
//
// La jerarquia son tres niveles y conviene tenerla clara antes de leer nada:
//
//	grupo    /aws/lambda/saludar        lo que TU nombras. Aqui vive la retencion.
//	flujo    2026/08/25/[$LATEST]a1b2   una fuente concreta: un contenedor, una maquina.
//	apunte   {marca, mensaje}           una linea, con la hora en milisegundos.
//
// ⚠️ Y el flujo no es un capricho de nomenclatura: es lo que hace que dos contenedores de la
// misma funcion escribiendo a la vez no se pisen. Cada uno tiene el suyo, y el orden dentro de
// un flujo si esta garantizado. ENTRE flujos, no.

import (
	"errors"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	errNoExisteGrupo   = errors.New("el grupo de registro no existe")
	errNoExisteFlujo   = errors.New("el flujo de registro no existe")
	errYaExisteGrupo   = errors.New("el grupo de registro ya existe")
	errYaExisteFlujo   = errors.New("el flujo de registro ya existe")
	errFueraDeOrden    = errors.New("los apuntes de un lote tienen que ir ordenados por marca de tiempo")
	errTokenEquivocado = errors.New("el sequenceToken no es el que toca")
)

// ventanaVieja y ventanaFutura son las dos que usa CloudWatch de verdad. No estan aqui por
// completismo: son la explicacion de «puse un log y no aparece». Si tu contenedor tiene la hora
// mal, o si reenvias un fichero de ayer, los apuntes se RECHAZAN y la llamada contesta 200.
const (
	ventanaVieja  = 14 * 24 * time.Hour
	ventanaFutura = 2 * time.Hour
	// Los topes del lote, tal cual los documenta Amazon. El de bytes cuenta 26 de sobrecarga
	// por apunte, que es el numero mas raro de toda la API y sale en la factura.
	maxApuntesLote = 10000
	maxBytesLote   = 1048576
	sobrecarga     = 26
)

// Apunte es una linea del registro. La marca va en milisegundos desde epoca, que es como viaja
// por la API: ni RFC3339, ni segundos, ni nanosegundos.
type Apunte struct {
	Marca   int64
	Mensaje string
	// orden desempata cuando dos apuntes comparten marca. Sin el, dos lineas escritas en el
	// mismo milisegundo saldrian en cualquier orden, y en un registro eso se nota: son las
	// dos mitades de la misma traza.
	orden int64
}

type Flujo struct {
	Nombre  string
	Creado  time.Time
	apuntes []Apunte
	// token es el `sequenceToken`. Vease `Poner` para la historia completa, que es buena.
	token  string
	subida time.Time
}

type Grupo struct {
	Nombre    string
	Creado    time.Time
	Retencion int32 // dias; 0 = para siempre
	flujos    map[string]*Flujo
}

type Registros struct {
	mu     sync.RWMutex
	grupos map[string]*Grupo
	sec    int64
	// tope de apuntes por flujo: esto es memoria, no disco. Sin tope, un contenedor en bucle
	// se come la maquina y el fallo aparece en OTRA prueba.
	tope int
}

func NuevosRegistros() *Registros {
	return &Registros{grupos: map[string]*Grupo{}, tope: 10000}
}

// ── grupos ───────────────────────────────────────────────────────────────

func (r *Registros) CrearGrupo(nombre string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, hay := r.grupos[nombre]; hay {
		return errYaExisteGrupo
	}
	r.grupos[nombre] = &Grupo{Nombre: nombre, Creado: time.Now().UTC(),
		flujos: map[string]*Flujo{}}
	return nil
}

func (r *Registros) BorrarGrupo(nombre string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, hay := r.grupos[nombre]; !hay {
		return errNoExisteGrupo
	}
	delete(r.grupos, nombre)
	return nil
}

// Retener pone (o quita, con 0) los dias de retencion del grupo.
//
// ⚠️ Y aqui va un «no» honesto: esto GUARDA el numero y no lo aplica. Borrar por antiguedad
// necesitaria un reloj corriendo por dentro, y un doble de pruebas que borra tus registros
// mientras depuras no ayuda a nadie. El campo esta para que tu codigo de infraestructura pueda
// leerlo y comprobarlo.
func (r *Registros) Retener(nombre string, dias int32) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	g, hay := r.grupos[nombre]
	if !hay {
		return errNoExisteGrupo
	}
	g.Retencion = dias
	return nil
}

// FichaGrupo es lo que se enseña por la API de un grupo.
type FichaGrupo struct {
	Nombre    string
	Creado    time.Time
	Retencion int32
	Flujos    int
	Bytes     int64
}

func (r *Registros) ListarGrupos(prefijo string) []FichaGrupo {
	r.mu.RLock()
	defer r.mu.RUnlock()
	fuera := make([]FichaGrupo, 0, len(r.grupos))
	for _, g := range r.grupos {
		if !strings.HasPrefix(g.Nombre, prefijo) {
			continue
		}
		f := FichaGrupo{Nombre: g.Nombre, Creado: g.Creado, Retencion: g.Retencion,
			Flujos: len(g.flujos)}
		for _, fl := range g.flujos {
			for _, a := range fl.apuntes {
				f.Bytes += int64(len(a.Mensaje))
			}
		}
		fuera = append(fuera, f)
	}
	ordenarPor(fuera, func(x FichaGrupo) string { return x.Nombre })
	return fuera
}

// ── flujos ───────────────────────────────────────────────────────────────

func (r *Registros) CrearFlujo(grupo, flujo string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	g, hay := r.grupos[grupo]
	if !hay {
		return errNoExisteGrupo
	}
	if _, hay := g.flujos[flujo]; hay {
		return errYaExisteFlujo
	}
	g.flujos[flujo] = &Flujo{Nombre: flujo, Creado: time.Now().UTC()}
	return nil
}

func (r *Registros) BorrarFlujo(grupo, flujo string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	g, hay := r.grupos[grupo]
	if !hay {
		return errNoExisteGrupo
	}
	if _, hay := g.flujos[flujo]; !hay {
		return errNoExisteFlujo
	}
	delete(g.flujos, flujo)
	return nil
}

// Asegurar crea el grupo y el flujo si no estaban. Es lo que hace Lambda de verdad con
// `/aws/lambda/<funcion>`: nadie lo crea a mano, aparece porque alguien escribio.
func (r *Registros) Asegurar(grupo, flujo string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	g, hay := r.grupos[grupo]
	if !hay {
		g = &Grupo{Nombre: grupo, Creado: time.Now().UTC(), flujos: map[string]*Flujo{}}
		r.grupos[grupo] = g
	}
	if _, hay := g.flujos[flujo]; !hay {
		g.flujos[flujo] = &Flujo{Nombre: flujo, Creado: time.Now().UTC()}
	}
}

// FichaFlujo es lo que se enseña por la API de un flujo.
type FichaFlujo struct {
	Nombre  string
	Creado  time.Time
	Primero int64
	Ultimo  int64
	Subida  time.Time
	Token   string
	Apuntes int
}

func (r *Registros) ListarFlujos(grupo, prefijo string) ([]FichaFlujo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	g, hay := r.grupos[grupo]
	if !hay {
		return nil, errNoExisteGrupo
	}
	fuera := make([]FichaFlujo, 0, len(g.flujos))
	for _, fl := range g.flujos {
		if !strings.HasPrefix(fl.Nombre, prefijo) {
			continue
		}
		f := FichaFlujo{Nombre: fl.Nombre, Creado: fl.Creado, Subida: fl.subida,
			Token: fl.token, Apuntes: len(fl.apuntes)}
		if len(fl.apuntes) > 0 {
			f.Primero = fl.apuntes[0].Marca
			f.Ultimo = fl.apuntes[len(fl.apuntes)-1].Marca
		}
		fuera = append(fuera, f)
	}
	ordenarPor(fuera, func(x FichaFlujo) string { return x.Nombre })
	return fuera, nil
}

// ── escribir ─────────────────────────────────────────────────────────────

// Rechazo cuenta lo que no entro, y por que. Va en la respuesta de `PutLogEvents`.
type Rechazo struct {
	Viejos int
	Nuevos int
}

// Poner mete un lote de apuntes en un flujo.
//
// 🥇 Y aqui esta la historia del `sequenceToken`, que es la mejor de este servicio. Durante
// nueve años, para escribir habia que mandar el token que te devolvio la escritura ANTERIOR. Si
// te equivocabas te contestaba `InvalidSequenceTokenException`... con el token bueno DENTRO del
// error. O sea: el error te daba exactamente el dato que te faltaba, igual que el error de
// permisos del episodio 10 y al reves que el de la firma del episodio 4.
//
// ⚠️ Servia para una cosa concreta: que dos escritores del MISMO flujo no se pisaran, porque el
// segundo tenia que haber leido lo que hizo el primero. En 2023 Amazon dejo de exigirlo, y hoy
// se acepta y se ignora. Aqui se implementa entero porque es lo que enseña por que existia: un
// contador de version, con otro nombre.
//
// Se acepta que venga vacio (como hoy) y se rechaza si viene EQUIVOCADO (como antes).
func (r *Registros) Poner(grupo, flujo, token string, apuntes []Apunte, ahora time.Time) (
	string, Rechazo, error) {
	if len(apuntes) == 0 {
		return "", Rechazo{}, nil
	}
	if len(apuntes) > maxApuntesLote {
		return "", Rechazo{}, errors.New("un lote no puede pasar de " +
			strconv.Itoa(maxApuntesLote) + " apuntes")
	}
	var bytes int
	for _, a := range apuntes {
		bytes += len(a.Mensaje) + sobrecarga
	}
	if bytes > maxBytesLote {
		return "", Rechazo{}, errors.New("un lote no puede pasar de " +
			strconv.Itoa(maxBytesLote) + " bytes, contando " + strconv.Itoa(sobrecarga) +
			" de sobrecarga por apunte")
	}
	// ⚠️ Ordenados, y lo comprueba el servidor. No es quisquilloseria: si no, el registro
	// dejaria de estar ordenado y `GetLogEvents` con su `nextToken` no podria funcionar, que
	// es toda la idea de un registro.
	for i := 1; i < len(apuntes); i++ {
		if apuntes[i].Marca < apuntes[i-1].Marca {
			return "", Rechazo{}, errFueraDeOrden
		}
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	g, hay := r.grupos[grupo]
	if !hay {
		return "", Rechazo{}, errNoExisteGrupo
	}
	fl, hay := g.flujos[flujo]
	if !hay {
		return "", Rechazo{}, errNoExisteFlujo
	}
	if token != "" && token != fl.token {
		return fl.token, Rechazo{}, errTokenEquivocado
	}

	// La ventana. Lo que cae fuera NO es un error: se cuenta y se tira, y la llamada contesta
	// que si. Por eso «puse un log y no aparece» es tan dificil de depurar.
	var rech Rechazo
	desde := ahora.Add(-ventanaVieja).UnixMilli()
	hasta := ahora.Add(ventanaFutura).UnixMilli()
	for _, a := range apuntes {
		switch {
		case a.Marca < desde:
			rech.Viejos++
		case a.Marca > hasta:
			rech.Nuevos++
		default:
			r.sec++
			a.orden = r.sec
			fl.apuntes = append(fl.apuntes, a)
		}
	}
	// Estable por marca, y el desempate es el orden de llegada. Un `sort.SliceStable` sobre la
	// marca sola bastaria hoy, pero el campo `orden` deja el criterio escrito en el dato en vez
	// de en el algoritmo.
	sort.SliceStable(fl.apuntes, func(i, j int) bool {
		if fl.apuntes[i].Marca != fl.apuntes[j].Marca {
			return fl.apuntes[i].Marca < fl.apuntes[j].Marca
		}
		return fl.apuntes[i].orden < fl.apuntes[j].orden
	})
	if r.tope > 0 && len(fl.apuntes) > r.tope {
		fl.apuntes = fl.apuntes[len(fl.apuntes)-r.tope:]
	}
	fl.subida = ahora.UTC()
	r.sec++
	fl.token = "t-" + pad12(r.sec)
	return fl.token, rech, nil
}

// Anexar es el atajo que usa el lanzador: crea lo que falte y mete una linea. No pasa por el
// token porque no hay dos escritores compitiendo: cada contenedor tiene su flujo.
func (r *Registros) Anexar(grupo, flujo string, marca int64, mensaje string) {
	r.Asegurar(grupo, flujo)
	_, _, _ = r.Poner(grupo, flujo, "", []Apunte{{Marca: marca, Mensaje: mensaje}}, time.Now())
}

// ── leer ─────────────────────────────────────────────────────────────────

// Pagina es lo que devuelve una lectura: los apuntes y los dos punteros.
type Pagina struct {
	Apuntes  []Apunte
	Adelante string
	Atras    string
}

// Leer saca una pagina de un flujo.
//
// 🥇 Y aqui esta la trampa que se lleva a mas gente por delante de toda esta API: `nextToken`
// NUNCA viene vacio. Ni cuando no hay mas. Un bucle «mientras haya token, pide otra pagina» no
// termina jamas: se queda dando vueltas devolviendo cero apuntes para siempre.
//
// La forma correcta de parar es comparar: si el token que te devuelven es EL MISMO que
// mandaste, ya no hay mas. Suena a detalle y es la diferencia entre un script que termina y uno
// que se queda colgado en produccion.
func (r *Registros) Leer(grupo, flujo string, desde, hasta *int64, limite int,
	token string, desdeElPrincipio bool) (Pagina, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	g, hay := r.grupos[grupo]
	if !hay {
		return Pagina{}, errNoExisteGrupo
	}
	fl, hay := g.flujos[flujo]
	if !hay {
		return Pagina{}, errNoExisteFlujo
	}
	if limite <= 0 || limite > 10000 {
		limite = 10000
	}

	// El filtro por tiempo primero: los tokens hablan de posiciones DENTRO de lo filtrado.
	candidatos := make([]Apunte, 0, len(fl.apuntes))
	for _, a := range fl.apuntes {
		if desde != nil && a.Marca < *desde {
			continue
		}
		// ⚠️ `endTime` es EXCLUSIVO en CloudWatch. Pedir «hasta las 12:00:00.000» no trae el
		// apunte de las 12:00:00.000. Es de las cosas que solo se descubren cuando falta uno.
		if hasta != nil && a.Marca >= *hasta {
			continue
		}
		candidatos = append(candidatos, a)
	}

	inicio := 0
	if token != "" {
		n, err := posicionDeToken(token)
		if err != nil {
			return Pagina{}, err
		}
		inicio = n
	} else if !desdeElPrincipio && len(candidatos) > limite {
		// Sin token y sin `startFromHead`, CloudWatch da el FINAL, no el principio. Lo que
		// quieres ver de un registro casi siempre es lo ultimo.
		inicio = len(candidatos) - limite
	}
	if inicio < 0 {
		inicio = 0
	}
	if inicio > len(candidatos) {
		inicio = len(candidatos)
	}
	fin := inicio + limite
	if fin > len(candidatos) {
		fin = len(candidatos)
	}

	pag := Pagina{
		Apuntes:  append([]Apunte(nil), candidatos[inicio:fin]...),
		Adelante: tokenDePosicion("f", fin),
		Atras:    tokenDePosicion("b", inicio),
	}
	return pag, nil
}

// Filtrar busca por texto en TODOS los flujos de un grupo a la vez.
//
// ⚠️ Y el «patron» de CloudWatch no son expresiones regulares, igual que el comodin de los
// permisos del episodio 10 no lo era. Aqui se implementa lo unico que usa todo el mundo: que la
// linea contenga ese texto. Lo demas —los patrones con comillas, los `?` de «o esto o lo
// otro»— es un lenguaje entero y no cabe en un doble de pruebas.
func (r *Registros) Filtrar(grupo, prefijoFlujo, patron string, desde, hasta *int64,
	limite int) ([]FilaFiltro, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	g, hay := r.grupos[grupo]
	if !hay {
		return nil, errNoExisteGrupo
	}
	if limite <= 0 || limite > 10000 {
		limite = 10000
	}
	var filas []FilaFiltro
	for _, fl := range g.flujos {
		if !strings.HasPrefix(fl.Nombre, prefijoFlujo) {
			continue
		}
		for _, a := range fl.apuntes {
			if desde != nil && a.Marca < *desde {
				continue
			}
			if hasta != nil && a.Marca >= *hasta {
				continue
			}
			if patron != "" && !strings.Contains(a.Mensaje, patron) {
				continue
			}
			filas = append(filas, FilaFiltro{Flujo: fl.Nombre, Apunte: a})
		}
	}
	// ⚠️ Entre flujos hay que ordenar aqui: cada flujo esta ordenado por dentro, pero el
	// conjunto no lo esta hasta que alguien lo ordena. Es la misma diferencia que hay entre
	// «cada particion esta ordenada» y «el registro esta ordenado».
	sort.SliceStable(filas, func(i, j int) bool {
		if filas[i].Apunte.Marca != filas[j].Apunte.Marca {
			return filas[i].Apunte.Marca < filas[j].Apunte.Marca
		}
		return filas[i].Apunte.orden < filas[j].Apunte.orden
	})
	if len(filas) > limite {
		filas = filas[:limite]
	}
	return filas, nil
}

// FilaFiltro es un apunte con el flujo del que salio: al buscar en varios flujos a la vez, sin
// eso no sabrias de donde viene cada linea.
type FilaFiltro struct {
	Flujo  string
	Apunte Apunte
}

// ── los tokens ───────────────────────────────────────────────────────────

// Los de verdad son opacos y llevan dentro mucho mas que un numero. Aqui son «f/12» o «b/0»
// porque un token opaco de mentira no enseña nada, y uno legible deja ver en las pruebas que el
// puntero lo lleva el LECTOR y no el servidor.

func tokenDePosicion(sentido string, n int) string {
	return sentido + "/" + strconv.Itoa(n)
}

func posicionDeToken(t string) (int, error) {
	i := strings.IndexByte(t, '/')
	if i < 0 {
		return 0, errors.New("nextToken no vale: " + t)
	}
	n, err := strconv.Atoi(t[i+1:])
	if err != nil {
		return 0, errors.New("nextToken no vale: " + t)
	}
	return n, nil
}

// nombreDeFlujo arma el nombre que usa Lambda de verdad: la fecha, la version y algo que
// distinga al contenedor. No es decorativo: por eso se puede mirar «los registros de ayer» sin
// abrir ninguno.
func nombreDeFlujo(cuando time.Time, id string) string {
	return cuando.UTC().Format("2006/01/02") + "/[$LATEST]" + corto(id)
}

// grupoDeFuncion es la convencion de Lambda, y es una convencion, no una regla del servicio de
// registros: quien decide que sus registros van a `/aws/lambda/<nombre>` es Lambda.
func grupoDeFuncion(nombre string) string { return "/aws/lambda/" + nombre }
