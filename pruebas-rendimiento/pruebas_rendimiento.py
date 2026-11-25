# -*- coding: utf-8 -*-
"""pruebas_rendimiento.py — mide de verdad lo que enseña el vídeo #116.

Ninguna cifra sale de un blog: se construyen los casos y se miden aquí.
Salida literal en `recursos/pruebas-rendimiento/SALIDA.md`.

    python canales/golandia/investigacion/pruebas_rendimiento.py
"""
import sys
from pathlib import Path
sys.path.insert(0, str(Path(__file__).resolve().parent))
import probar
from probar import episodio

probar.RAIZ = Path(r"<repo>")

# ── 1. LA REGLA CERO: la «optimización obvia» que no cambia nada ──────────────────────
REGLA_CERO = r'''package regla

import "testing"

// El caso que todo el mundo ha visto en una revision de codigo: «saca el len() del bucle,
// que se recalcula en cada vuelta». Suena a mejora. Vamos a medirlo.

func SumaIngenua(xs []int) int {
	t := 0
	for i := 0; i < len(xs); i++ { // len() en la condicion
		t += xs[i]
	}
	return t
}

func SumaOptimizada(xs []int) int {
	t := 0
	n := len(xs) // «lo saco fuera para que no se recalcule»
	for i := 0; i < n; i++ {
		t += xs[i]
	}
	return t
}

var datos = make([]int, 100000)
var sink int

func BenchmarkIngenua(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sink = SumaIngenua(datos)
	}
}

func BenchmarkOptimizada(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sink = SumaOptimizada(datos)
	}
}
'''

# ── 2. Concatenar texto: el caso donde lo que importa son las ASIGNACIONES ────────────
TEXTO = r'''package texto

import (
	"bytes"
	"strings"
	"testing"
)

const N = 2000

func ConMas() string {
	s := ""
	for i := 0; i < N; i++ {
		s += "linea de registro\n" // cada vuelta crea una cadena NUEVA
	}
	return s
}

func ConBuilder() string {
	var b strings.Builder
	for i := 0; i < N; i++ {
		b.WriteString("linea de registro\n")
	}
	return b.String()
}

func ConBuilderReservado() string {
	var b strings.Builder
	b.Grow(N * len("linea de registro\n")) // se sabe cuanto va a ocupar: se pide de una vez
	for i := 0; i < N; i++ {
		b.WriteString("linea de registro\n")
	}
	return b.String()
}

func ConBuffer() string {
	var b bytes.Buffer
	for i := 0; i < N; i++ {
		b.WriteString("linea de registro\n")
	}
	return b.String()
}

var s string

func BenchmarkConcatMas(b *testing.B) {
	for i := 0; i < b.N; i++ {
		s = ConMas()
	}
}
func BenchmarkConcatBuilder(b *testing.B) {
	for i := 0; i < b.N; i++ {
		s = ConBuilder()
	}
}
func BenchmarkConcatBuilderReservado(b *testing.B) {
	for i := 0; i < b.N; i++ {
		s = ConBuilderReservado()
	}
}
func BenchmarkConcatBuffer(b *testing.B) {
	for i := 0; i < b.N; i++ {
		s = ConBuffer()
	}
}
'''

# ── 3. Reservar capacidad al crear el slice ───────────────────────────────────────────
CAPACIDAD = r'''package capacidad

import "testing"

const N = 10000

func SinReservar() []int {
	var xs []int
	for i := 0; i < N; i++ {
		xs = append(xs, i) // se muda de array unas cuantas veces
	}
	return xs
}

func Reservando() []int {
	xs := make([]int, 0, N) // una sola peticion de memoria
	for i := 0; i < N; i++ {
		xs = append(xs, i)
	}
	return xs
}

var sink []int

func BenchmarkSinReservar(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sink = SinReservar()
	}
}
func BenchmarkReservando(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sink = Reservando()
	}
}
'''

# ── 4. Escape analysis: qué se va al montón y por qué ─────────────────────────────────
ESCAPES = r'''package main

import "fmt"

type Punto struct{ X, Y int }

// Se queda en la PILA: no sale de la funcion.
func enPila() int {
	p := Punto{1, 2}
	return p.X + p.Y
}

// Se ESCAPA al monton: se devuelve un puntero a algo local.
func seEscapa() *Punto {
	p := Punto{1, 2}
	return &p
}

// Tambien se escapa, y esta sorprende: fmt.Println recibe `any`, y eso obliga
// a que el valor viva en el monton aunque no salga de aqui.
func porCulpaDeFmt() {
	p := Punto{3, 4}
	fmt.Println(p.X)
	_ = p
}

func main() {
	_ = enPila()
	_ = seEscapa()
	porCulpaDeFmt()
}
'''

# ── 5. pprof de verdad: perfilar, encontrar el punto caliente y volver a medir ─────────
PPROF_LENTO = r'''package main

import (
	"fmt"
	"os"
	"runtime/pprof"
	"strings"
)

// Un trabajo realista: normalizar y contar palabras de un texto grande.
// Tiene un punto caliente escondido a proposito.
func normalizar(palabras []string) map[string]int {
	cuenta := map[string]int{}
	for _, p := range palabras {
		// EL PUNTO CALIENTE: se construye la clave concatenando en cada vuelta.
		clave := ""
		for _, r := range strings.ToLower(p) {
			if r >= 'a' && r <= 'z' {
				clave += string(r)
			}
		}
		cuenta[clave]++
	}
	return cuenta
}

func main() {
	f, _ := os.Create("cpu.pprof")
	pprof.StartCPUProfile(f)
	defer func() { pprof.StopCPUProfile(); f.Close() }()

	base := strings.Fields(strings.Repeat("El Rapido Zorro Marron Salta Sobre El Perro Perezoso ", 4000))

	// ⚠️ El perfilador de CPU muestrea a 100 Hz. Con una sola pasada el programa duraba 5 ms
	// y el perfil salia con «Total samples = 0»: cero muestras, nada que mirar. Se repite el
	// trabajo hasta pasar del segundo, que es lo minimo para que el `top` diga algo.
	var c map[string]int
	for i := 0; i < 400; i++ {
		c = normalizar(base)
	}
	fmt.Println("palabras distintas:", len(c))
}
'''

PPROF_RAPIDO = r'''package main

import (
	"fmt"
	"strings"
)

// La misma funcion, con el punto caliente arreglado: un Builder con capacidad
// en vez de concatenar caracter a caracter.
func normalizar(palabras []string) map[string]int {
	cuenta := map[string]int{}
	var b strings.Builder
	for _, p := range palabras {
		b.Reset()
		b.Grow(len(p))
		for i := 0; i < len(p); i++ {
			c := p[i]
			if c >= 'A' && c <= 'Z' {
				c += 'a' - 'A'
			}
			if c >= 'a' && c <= 'z' {
				b.WriteByte(c)
			}
		}
		cuenta[b.String()]++
	}
	return cuenta
}

func main() {
	base := strings.Fields(strings.Repeat("El Rapido Zorro Marron Salta Sobre El Perro Perezoso ", 4000))
	c := normalizar(base)
	fmt.Println("palabras distintas:", len(c))
}
'''

PPROF_BENCH = r'''package pp

import (
	"strings"
	"testing"
)

var base = strings.Fields(strings.Repeat("El Rapido Zorro Marron Salta Sobre El Perro Perezoso ", 4000))

func lento(palabras []string) map[string]int {
	cuenta := map[string]int{}
	for _, p := range palabras {
		clave := ""
		for _, r := range strings.ToLower(p) {
			if r >= 'a' && r <= 'z' {
				clave += string(r)
			}
		}
		cuenta[clave]++
	}
	return cuenta
}

func rapido(palabras []string) map[string]int {
	cuenta := map[string]int{}
	var b strings.Builder
	for _, p := range palabras {
		b.Reset()
		b.Grow(len(p))
		for i := 0; i < len(p); i++ {
			c := p[i]
			if c >= 'A' && c <= 'Z' {
				c += 'a' - 'A'
			}
			if c >= 'a' && c <= 'z' {
				b.WriteByte(c)
			}
		}
		cuenta[b.String()]++
	}
	return cuenta
}

var sink map[string]int

func BenchmarkAntes(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sink = lento(base)
	}
}
func BenchmarkDespues(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sink = rapido(base)
	}
}
'''

# ── 6. El recolector: GOGC, GOMEMLIMIT, y si asignar menos gana a tocarlo ─────────────
GC = r'''package main

import (
	"fmt"
	"os"
	"runtime"
	"runtime/debug"
	"time"
)

type Registro struct {
	ID     int
	Nombre string
	Datos  [64]byte
}

// ⚠️ El sumidero global es OBLIGATORIO, no adorno. Sin el, el escape analysis ve que el
// puntero no sale de la funcion, lo deja en la PILA y no hay ni una asignacion en el monton:
// la primera version de esta prueba daba 0 ciclos de GC y 0 MB, o sea que no medía nada.
var sumidero *Registro

// Version que ASIGNA MUCHO: un puntero nuevo por vuelta, y que se escapa de verdad.
func asignaMucho(n int) int {
	total := 0
	for i := 0; i < n; i++ {
		r := &Registro{ID: i, Nombre: "linea"}
		sumidero = r // fuerza el monton: sin esto el compilador lo borra
		total += r.ID
	}
	return total
}

// Version que ASIGNA POCO: reutiliza el mismo valor, sin puntero.
func asignaPoco(n int) int {
	total := 0
	var r Registro
	for i := 0; i < n; i++ {
		r.ID = i
		r.Nombre = "linea"
		total += r.ID
	}
	return total
}

func medir(nombre string, f func(int) int, n int) {
	runtime.GC()
	var antes, despues runtime.MemStats
	runtime.ReadMemStats(&antes)
	t := time.Now()
	f(n)
	d := time.Since(t)
	runtime.ReadMemStats(&despues)
	fmt.Printf("  %-34s %8v  ciclos de GC: %-4d  asignado: %5d MB\n",
		nombre, d.Round(time.Millisecond),
		despues.NumGC-antes.NumGC,
		(despues.TotalAlloc-antes.TotalAlloc)/1024/1024)
}

func main() {
	const N = 5000000
	modo := ""
	if len(os.Args) > 1 {
		modo = os.Args[1]
	}
	switch modo {
	case "gogc":
		fmt.Println("MISMO codigo (el que asigna mucho), tocando solo el recolector")
		for _, g := range []int{100, 400, 1600} {
			debug.SetGCPercent(g)
			medir(fmt.Sprintf("GOGC=%d", g), asignaMucho, N)
		}
	default:
		debug.SetGCPercent(100) // el valor por defecto
		fmt.Println("MISMO recolector (GOGC=100), cambiando el codigo")
		medir("asigna mucho (puntero por vuelta)", asignaMucho, N)
		medir("asigna poco (reutiliza el valor)", asignaPoco, N)
	}
}
'''

# ── 7. El cierre honesto: optimizar la CPU no sirve si esperas a la red ───────────────
HONESTO = r'''package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"
)

var _ = sql.ErrNoRows

// Un manejador realista: pide un dato y lo formatea.
// El formateo es el codigo «lento»; la peticion es la espera real.
func trabajoCPU(n int) string {
	var b strings.Builder
	for i := 0; i < n; i++ {
		fmt.Fprintf(&b, "fila %d;", i)
	}
	return b.String()
}

func main() {
	// La «base de datos»: 40 ms por consulta, que es de lo normal tirando a bueno.
	bd := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(40 * time.Millisecond)
		w.Write([]byte("ok"))
	}))
	defer bd.Close()

	medir := func(nombre string, filas int, consultas int) {
		t := time.Now()
		for i := 0; i < consultas; i++ {
			res, _ := http.Get(bd.URL)
			res.Body.Close()
		}
		tBD := time.Since(t)
		t2 := time.Now()
		_ = trabajoCPU(filas)
		tCPU := time.Since(t2)
		fmt.Printf("  %-30s  base de datos: %-8v  CPU: %-10v  total: %v\n",
			nombre, tBD.Round(time.Millisecond),
			tCPU.Round(10*time.Microsecond),
			(tBD + tCPU).Round(time.Millisecond))
	}

	fmt.Println("Una peticion que formatea 5.000 filas:")
	medir("1 consulta", 5000, 1)
	fmt.Println()
	fmt.Println("La misma, con el clasico N+1 (una consulta por fila... 50 de ellas):")
	medir("50 consultas (N+1)", 5000, 50)
	fmt.Println()
	fmt.Println("Optimizar el formateo al doble de rapido ahorra microsegundos.")
	fmt.Println("Quitar el N+1 ahorra dos segundos.")
}
'''

MENTIRA = r'''package mentira

import "testing"

type Pedido struct {
	ID    int64
	Notas [256]byte
}

// Funcion normal: pequena y por tanto INCRUSTABLE. Es el caso realista.
func total(p Pedido) int64 { return p.ID * 2 }

// EL BANCO QUE MIENTE: no se usa el resultado. El compilador incrusta la funcion,
// ve que nadie mira el valor, y borra el trabajo entero. El numero no mide nada.
func BenchmarkMiente(b *testing.B) {
	p := Pedido{ID: 42}
	for i := 0; i < b.N; i++ {
		total(p)
	}
}

// EL BANCO HONESTO: el resultado va a una variable de paquete, que el compilador
// no puede demostrar que sobre. Ahora si se ejecuta.
var sumidero int64

func BenchmarkHonesto(b *testing.B) {
	p := Pedido{ID: 42}
	for i := 0; i < b.N; i++ {
		sumidero = total(p)
	}
}
'''

episodio(
    "pruebas-rendimiento",
    "Vídeo 116 — ¿Por qué mi programa en Go va lento? (todo medido)",
    [
        {"nombre": "1. La regla cero: la optimizacion obvia que no cambia nada",
         "pregunta": "«Saca el len() del bucle». ¿Cuanto mejora eso?",
         "archivos": {"regla/bench_test.go": REGLA_CERO},
         "correr": [(["go", "test", "-bench=.", "-benchmem", "-count=3", "./regla"], False)],
         "nota": "Si sale igual, ese es el gancho del video: la optimizacion que todo el mundo "
                 "pide en una revision de codigo y no hace nada."},
        {"nombre": "2. Concatenar texto: lo que importa son las ASIGNACIONES",
         "pregunta": "Cuanto cuesta de verdad construir un texto con +=?",
         "archivos": {"texto/bench_test.go": TEXTO},
         "correr": [(["go", "test", "-bench=.", "-benchmem", "-count=2", "./texto"], False)],
         "nota": "Mirar la columna de allocs/op mas que la de ns/op: es la que explica el resto."},
        {"nombre": "3. Reservar la capacidad al crear el slice",
         "pregunta": "Cuanto ahorra un make([]T, 0, n)?",
         "archivos": {"capacidad/bench_test.go": CAPACIDAD},
         "correr": [(["go", "test", "-bench=.", "-benchmem", "-count=2", "./capacidad"], False)]},
        {"nombre": "4. Escape analysis: que se va al monton y por que",
         "pregunta": "Como sé si algo se escapa al monton?",
         "archivos": {"escapes/main.go": ESCAPES},
         "correr": [(["sh", "-c", "cd escapes && go build -gcflags=-m . 2>&1 | head -20"], False)],
         "nota": "Es GRATIS de ejecutar y casi nadie lo ha visto nunca. El caso de fmt es el que "
                 "sorprende: recibir `any` obliga al valor a vivir en el monton."},
        {"nombre": "5. pprof: perfilar, encontrar el punto caliente y volver a medir",
         "pregunta": "Como encuentro DONDE se va el tiempo?",
         "archivos": {"perfil/main.go": PPROF_LENTO, "pp/bench_test.go": PPROF_BENCH},
         "correr": [(["sh", "-c", "cd perfil && go run . && go tool pprof -top -nodecount=8 cpu.pprof 2>&1 | head -18"], False),
                    (["go", "test", "-bench=.", "-benchmem", "-count=2", "./pp"], False)],
         "nota": "El ciclo entero: perfilar, leer el top, arreglar ESE punto y medir otra vez."},
        {"nombre": "6. El recolector: asignar menos contra tocar GOGC",
         "pregunta": "Se arregla subiendo GOGC?",
         "archivos": {"gc/main.go": GC},
         "correr": [(["go", "run", "./gc"], False),
                    (["go", "run", "./gc", "gogc"], False)],
         "nota": "La comparacion que decide el capitulo: cambiar el CODIGO contra cambiar el "
                 "RECOLECTOR. Se dice lo que salga, aunque contradiga lo esperado."},
        {"nombre": "7. El cierre honesto: la CPU no es donde se va el tiempo",
         "pregunta": "Y si optimizo el codigo y sigue lento?",
         "archivos": {"io/main.go": HONESTO},
         "correr": [(["go", "run", "./io"], False)],
         "nota": "La marca de la casa. Casi todos los programas de Go no son lentos por la CPU: "
                 "son lentos por una consulta de mas."},
        {"nombre": "8. Como NO mentirte con un banco de pruebas",
         "pregunta": "Puede el compilador borrar lo que estoy midiendo?",
         "archivos": {"mentira/bench_test.go": MENTIRA},
         "correr": [(["go", "test", "-bench=.", "-count=2", "./mentira"], False)],
         "nota": "Paso de verdad en la entrega 4 de la serie de entrevistas: el primer benchmark daba 0,13 ns/op porque el compilador se lo habia comido entero."},
    ],
    gomod="module pruebas\n\ngo 1.24\n")
