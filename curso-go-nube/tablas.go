package main

// El estado de DynamoDB: tablas y elementos. Igual que en S3, es lo unico con candado.
//
// Lo que hay que entender ANTES de escribir una linea, y que cambia como diseñas tus tablas:
//
//	· Un elemento es un MAPA de atributos con TIPO: {"nombre": {"S":"ana"}, "edad": {"N":"31"}}.
//	· Los numeros viajan como CADENA. No es un descuido: asi el numero no pasa por el float de
//	  ningun lenguaje y no pierde precision por el camino.
//	· La CLAVE se elige al crear la tabla y no se cambia nunca. Es de particion, o de particion
//	  mas orden. Todo lo demas del elemento es libre: no hay esquema.
//	· `Query` solo sabe buscar POR LA CLAVE. Si preguntas por otra cosa, es `Scan`, y `Scan` lee
//	  la tabla entera. Por eso la clave se elige una vez y con cuidado.

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"math/big"
	"sort"
	"strings"
	"sync"
	"time"
)

// AV es un AttributeValue tal como viaja por el cable: {"S":"hola"} · {"N":"42"} · {"BOOL":true}
type AV map[string]any

// Elemento es una fila: un mapa de nombre de atributo a valor con tipo.
type Elemento map[string]AV

type Clave struct {
	Nombre string // el nombre del atributo
	Tipo   string // S | N | B
}

type Tabla struct {
	Nombre    string
	PK        Clave
	SK        *Clave // nil si la tabla solo tiene clave de particion
	Creada    time.Time
	elementos map[string]Elemento // clave compuesta serializada -> elemento
}

type Tablas struct {
	mu     sync.RWMutex
	tablas map[string]*Tabla
	reloj  func() time.Time
}

func NuevasTablas() *Tablas {
	return &Tablas{tablas: map[string]*Tabla{}, reloj: time.Now}
}

// ── comparar valores, que es lo que hace que el orden exista ─────────────

// escalar saca el valor comparable de un AV de clave. Solo S, N y B pueden ser clave.
func escalar(v AV) (tipo string, s string, n *big.Float, b []byte, ok bool) {
	if x, hay := v["S"]; hay {
		t, _ := x.(string)
		return "S", t, nil, nil, true
	}
	if x, hay := v["N"]; hay {
		t, _ := x.(string)
		f, _, err := big.ParseFloat(t, 10, 200, big.ToNearestEven)
		if err != nil {
			return "N", t, nil, nil, false
		}
		return "N", t, f, nil, true
	}
	if x, hay := v["B"]; hay {
		t, _ := x.(string)
		raw, err := base64.StdEncoding.DecodeString(t)
		if err != nil {
			return "B", t, nil, nil, false
		}
		return "B", t, nil, raw, true
	}
	return "", "", nil, nil, false
}

// compara devuelve -1, 0 o 1. ⚠️ Las cadenas se comparan POR BYTES, igual que DynamoDB. Por eso
// las fechas se guardan en ISO-8601: "2026-02-01" va despues de "2026-01-31" por puro orden de
// bytes, sin que nadie sepa que son fechas.
func compara(a, b AV) (int, bool) {
	ta, sa, na, ba, oka := escalar(a)
	tb, sb, nb, bb, okb := escalar(b)
	if !oka || !okb || ta != tb {
		return 0, false
	}
	switch ta {
	case "S":
		return strings.Compare(sa, sb), true
	case "N":
		return na.Cmp(nb), true
	case "B":
		return bytes.Compare(ba, bb), true
	}
	return 0, false
}

// serie convierte un valor de clave en algo que sirve de clave de mapa, CON su tipo delante:
// sin el tipo, el numero 1 y la cadena "1" colisionarian.
func serie(v AV) (string, bool) {
	t, s, _, _, ok := escalar(v)
	if !ok {
		return "", false
	}
	return t + ":" + s, true
}

func (t *Tabla) claveDe(e Elemento) (string, error) {
	pk, hay := e[t.PK.Nombre]
	if !hay {
		return "", fmt.Errorf("falta la clave de particion %q", t.PK.Nombre)
	}
	sp, ok := serie(pk)
	if !ok {
		return "", fmt.Errorf("la clave de particion %q no es S, N ni B", t.PK.Nombre)
	}
	if t.SK == nil {
		return sp, nil
	}
	sk, hay := e[t.SK.Nombre]
	if !hay {
		return "", fmt.Errorf("falta la clave de orden %q", t.SK.Nombre)
	}
	ss, ok := serie(sk)
	if !ok {
		return "", fmt.Errorf("la clave de orden %q no es S, N ni B", t.SK.Nombre)
	}
	return sp + "\x00" + ss, nil
}

// ── operaciones sobre las tablas ─────────────────────────────────────────

func (ts *Tablas) Crear(nombre string, pk Clave, sk *Clave) (*Tabla, bool) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	if _, ya := ts.tablas[nombre]; ya {
		return nil, false
	}
	t := &Tabla{Nombre: nombre, PK: pk, SK: sk, Creada: ts.reloj().UTC(),
		elementos: map[string]Elemento{}}
	ts.tablas[nombre] = t
	return t, true
}

func (ts *Tablas) Buscar(nombre string) (*Tabla, bool) {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	t, ok := ts.tablas[nombre]
	return t, ok
}

func (ts *Tablas) Borrar(nombre string) (*Tabla, bool) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	t, ok := ts.tablas[nombre]
	if ok {
		delete(ts.tablas, nombre)
	}
	return t, ok
}

// Listar devuelve los nombres ORDENADOS: un mapa de Go no tiene orden y sin ordenar la salida
// cambia en cada llamada.
func (ts *Tablas) Listar() []string {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	fuera := make([]string, 0, len(ts.tablas))
	for n := range ts.tablas {
		fuera = append(fuera, n)
	}
	sort.Strings(fuera)
	return fuera
}

func (ts *Tablas) Cuenta(nombre string) int {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	if t, ok := ts.tablas[nombre]; ok {
		return len(t.elementos)
	}
	return 0
}

// ── operaciones sobre los elementos ──────────────────────────────────────

// Poner guarda el elemento y devuelve el que habia antes (nil si no habia). Devolverlo es lo que
// permite implementar ReturnValues=ALL_OLD sin volver a leer.
func (ts *Tablas) Poner(tabla string, e Elemento, condicion func(Elemento) bool) (Elemento, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	t, ok := ts.tablas[tabla]
	if !ok {
		return nil, errNoExisteTabla
	}
	k, err := t.claveDe(e)
	if err != nil {
		return nil, err
	}
	anterior := t.elementos[k]
	if condicion != nil && !condicion(anterior) {
		return nil, errCondicion
	}
	t.elementos[k] = e
	return anterior, nil
}

func (ts *Tablas) Obtener(tabla string, clave Elemento) (Elemento, error) {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	t, ok := ts.tablas[tabla]
	if !ok {
		return nil, errNoExisteTabla
	}
	k, err := t.claveDe(clave)
	if err != nil {
		return nil, err
	}
	return t.elementos[k], nil
}

func (ts *Tablas) Quitar(tabla string, clave Elemento, condicion func(Elemento) bool) (Elemento, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	t, ok := ts.tablas[tabla]
	if !ok {
		return nil, errNoExisteTabla
	}
	k, err := t.claveDe(clave)
	if err != nil {
		return nil, err
	}
	anterior := t.elementos[k]
	if condicion != nil && !condicion(anterior) {
		return nil, errCondicion
	}
	delete(t.elementos, k)
	return anterior, nil
}

// Recorrer entrega los elementos ordenados por clave de particion y, dentro, por clave de orden.
// Es lo que necesitan Scan y Query, y lo que hace que el resultado sea estable entre corridas.
//
// ⚠️ Se ordena COMPARANDO LOS VALORES, no la cadena con la que se indexan. Ordenar por la cadena
// pondria el 10 antes que el 9, porque "10" es menor que "9" byte a byte. En una clave de orden
// numerica eso devuelve la pagina equivocada y no lo avisa nadie.
func (ts *Tablas) Recorrer(tabla string) ([]Elemento, *Tabla, error) {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	t, ok := ts.tablas[tabla]
	if !ok {
		return nil, nil, errNoExisteTabla
	}
	fuera := make([]Elemento, 0, len(t.elementos))
	for _, e := range t.elementos {
		fuera = append(fuera, e)
	}
	sort.SliceStable(fuera, func(i, j int) bool { return t.antes(fuera[i], fuera[j]) })
	return fuera, t, nil
}

// antes dice si el elemento a va antes que el b: primero por particion, luego por orden.
func (t *Tabla) antes(a, b Elemento) bool {
	if c, ok := compara(a[t.PK.Nombre], b[t.PK.Nombre]); ok && c != 0 {
		return c < 0
	}
	if t.SK == nil {
		return false
	}
	c, ok := compara(a[t.SK.Nombre], b[t.SK.Nombre])
	return ok && c < 0
}
