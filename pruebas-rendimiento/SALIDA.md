# Pruebas ejecutadas — Vídeo 116 — ¿Por qué mi programa en Go va lento? (todo medido)

> Salida **literal** de esta máquina. La genera `probar.py` y se vuelve a generar sola.
> ⛔ Lo que no esté aquí no puede salir en pantalla (canon: ninguna cifra inventada,
> y lo que se muestra como código se ejecuta antes).

**Entorno citable en el vídeo:** contenedor `golang:1.26` · `go version go1.26.7 linux/amd64`

---

## 1. La regla cero: la optimizacion obvia que no cambia nada

**La pregunta:** «Saca el len() del bucle». ¿Cuanto mejora eso?

`regla/bench_test.go`

```go
package regla

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
```

```console
$ go test -bench=. -benchmem -count=3 ./regla
goos: linux
goarch: amd64
pkg: pruebas/regla
cpu: Intel(R) Core(TM) Ultra 7 270K Plus
BenchmarkIngenua-24       	    3732	    316838 ns/op	       0 B/op	       0 allocs/op
BenchmarkIngenua-24       	    3746	    316023 ns/op	       0 B/op	       0 allocs/op
BenchmarkIngenua-24       	    3774	    320264 ns/op	       0 B/op	       0 allocs/op
BenchmarkOptimizada-24    	    4906	    317061 ns/op	       0 B/op	       0 allocs/op
BenchmarkOptimizada-24    	    3793	    316162 ns/op	       0 B/op	       0 allocs/op
BenchmarkOptimizada-24    	    3735	    316564 ns/op	       0 B/op	       0 allocs/op
PASS
ok  	pruebas/regla	7.716s
[codigo de salida: 0]
```

> Si sale igual, ese es el gancho del video: la optimizacion que todo el mundo pide en una revision de codigo y no hace nada.

---

## 2. Concatenar texto: lo que importa son las ASIGNACIONES

**La pregunta:** Cuanto cuesta de verdad construir un texto con +=?

`texto/bench_test.go`

```go
package texto

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
```

```console
$ go test -bench=. -benchmem -count=2 ./texto
goos: linux
goarch: amd64
pkg: pruebas/texto
cpu: Intel(R) Core(TM) Ultra 7 270K Plus
BenchmarkConcatMas-24                 	     290	   3640356 ns/op	38794744 B/op	    2003 allocs/op
BenchmarkConcatMas-24                 	     332	   4320710 ns/op	38794832 B/op	    2003 allocs/op
BenchmarkConcatBuilder-24             	   50460	     20954 ns/op	  153000 B/op	      18 allocs/op
BenchmarkConcatBuilder-24             	   57298	     19784 ns/op	  153000 B/op	      18 allocs/op
BenchmarkConcatBuilderReservado-24    	  212791	      6531 ns/op	   40960 B/op	       1 allocs/op
BenchmarkConcatBuilderReservado-24    	  213531	      5547 ns/op	   40960 B/op	       1 allocs/op
BenchmarkConcatBuffer-24              	   56928	     25160 ns/op	  171969 B/op	      12 allocs/op
BenchmarkConcatBuffer-24              	   49755	     22196 ns/op	  171968 B/op	      12 allocs/op
PASS
ok  	pruebas/texto	11.617s
[codigo de salida: 0]
```

> Mirar la columna de allocs/op mas que la de ns/op: es la que explica el resto.

---

## 3. Reservar la capacidad al crear el slice

**La pregunta:** Cuanto ahorra un make([]T, 0, n)?

`capacidad/bench_test.go`

```go
package capacidad

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
```

```console
$ go test -bench=. -benchmem -count=2 ./capacidad
goos: linux
goarch: amd64
pkg: pruebas/capacidad
cpu: Intel(R) Core(TM) Ultra 7 270K Plus
BenchmarkSinReservar-24    	   31471	     40580 ns/op	  357630 B/op	      19 allocs/op
BenchmarkSinReservar-24    	   29812	     38230 ns/op	  357631 B/op	      19 allocs/op
BenchmarkReservando-24     	  159566	      7763 ns/op	   81920 B/op	       1 allocs/op
BenchmarkReservando-24     	  136078	      8456 ns/op	   81920 B/op	       1 allocs/op
PASS
ok  	pruebas/capacidad	5.776s
[codigo de salida: 0]
```

---

## 4. Escape analysis: que se va al monton y por que

**La pregunta:** Como sé si algo se escapa al monton?

`escapes/main.go`

```go
package main

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
```

```console
$ sh -c cd escapes && go build -gcflags=-m . 2>&1 | head -20

[codigo de salida: 0]
```

> Es GRATIS de ejecutar y casi nadie lo ha visto nunca. El caso de fmt es el que sorprende: recibir `any` obliga al valor a vivir en el monton.

---

## 5. pprof: perfilar, encontrar el punto caliente y volver a medir

**La pregunta:** Como encuentro DONDE se va el tiempo?

`perfil/main.go`

```go
package main

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
```

`pp/bench_test.go`

```go
package pp

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
```

```console
$ sh -c cd perfil && go run . && go tool pprof -top -nodecount=8 cpu.pprof 2>&1 | head -18
palabras distintas: 8
File: perfil
Build ID: e1037e11240d18bfb9ec33afacd9e4ec44abf1ff
Type: cpu
Time: 2026-08-23 13:36:50 UTC
Duration: 1.46s, Total samples = 1620ms (111.31%)
Showing nodes accounting for 1020ms, 62.96% of 1620ms total
Showing top 8 nodes out of 72
      flat  flat%   sum%        cum   cum%
     240ms 14.81% 14.81%      500ms 30.86%  runtime.mallocgcTiny
     210ms 12.96% 27.78%      710ms 43.83%  runtime.concatstrings
     130ms  8.02% 35.80%      130ms  8.02%  runtime.memmove
     100ms  6.17% 41.98%      100ms  6.17%  aeshashbody
     100ms  6.17% 48.15%      600ms 37.04%  runtime.mallocgc
      90ms  5.56% 53.70%     1510ms 93.21%  main.normalizar
      90ms  5.56% 59.26%       90ms  5.56%  runtime.nextFreeFast (inline)
      60ms  3.70% 62.96%      250ms 15.43%  runtime.mapassign_faststr
[codigo de salida: 0]
```

```console
$ go test -bench=. -benchmem -count=2 ./pp
goos: linux
goarch: amd64
pkg: pruebas/pp
cpu: Intel(R) Core(TM) Ultra 7 270K Plus
BenchmarkAntes-24      	     321	   3703054 ns/op	 1856727 B/op	  352005 allocs/op
BenchmarkAntes-24      	     326	   3713746 ns/op	 1856723 B/op	  352005 allocs/op
BenchmarkDespues-24    	    1485	    796759 ns/op	  288714 B/op	   36005 allocs/op
BenchmarkDespues-24    	    1524	    797918 ns/op	  288714 B/op	   36005 allocs/op
PASS
ok  	pruebas/pp	5.720s
[codigo de salida: 0]
```

> El ciclo entero: perfilar, leer el top, arreglar ESE punto y medir otra vez.

---

## 6. El recolector: asignar menos contra tocar GOGC

**La pregunta:** Se arregla subiendo GOGC?

`gc/main.go`

```go
package main

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
```

```console
$ go run ./gc
MISMO recolector (GOGC=100), cambiando el codigo
  asigna mucho (puntero por vuelta)     109ms  ciclos de GC: 151   asignado:   457 MB
  asigna poco (reutiliza el valor)        1ms  ciclos de GC: 0     asignado:     0 MB
[codigo de salida: 0]
```

```console
$ go run ./gc gogc
MISMO codigo (el que asigna mucho), tocando solo el recolector
  GOGC=100                              115ms  ciclos de GC: 158   asignado:   457 MB
  GOGC=400                               83ms  ciclos de GC: 32    asignado:   457 MB
  GOGC=1600                              91ms  ciclos de GC: 7     asignado:   457 MB
[codigo de salida: 0]
```

> La comparacion que decide el capitulo: cambiar el CODIGO contra cambiar el RECOLECTOR. Se dice lo que salga, aunque contradiga lo esperado.

---

## 7. El cierre honesto: la CPU no es donde se va el tiempo

**La pregunta:** Y si optimizo el codigo y sigue lento?

`io/main.go`

```go
package main

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
```

```console
$ go run ./io
Una peticion que formatea 5.000 filas:
  1 consulta                      base de datos: 41ms      CPU: 280µs       total: 41ms

La misma, con el clasico N+1 (una consulta por fila... 50 de ellas):
  50 consultas (N+1)              base de datos: 2.033s    CPU: 280µs       total: 2.033s

Optimizar el formateo al doble de rapido ahorra microsegundos.
Quitar el N+1 ahorra dos segundos.
[codigo de salida: 0]
```

> La marca de la casa. Casi todos los programas de Go no son lentos por la CPU: son lentos por una consulta de mas.

---

## 8. Como NO mentirte con un banco de pruebas

**La pregunta:** Puede el compilador borrar lo que estoy midiendo?

`mentira/bench_test.go`

```go
package mentira

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
```

```console
$ go test -bench=. -count=2 ./mentira
goos: linux
goarch: amd64
pkg: pruebas/mentira
cpu: Intel(R) Core(TM) Ultra 7 270K Plus
BenchmarkMiente-24     	1000000000	         0.09880 ns/op
BenchmarkMiente-24     	1000000000	         0.09736 ns/op
BenchmarkHonesto-24    	356225702	         3.375 ns/op
BenchmarkHonesto-24    	356313454	         3.388 ns/op
PASS
ok  	pruebas/mentira	3.326s
[codigo de salida: 0]
```

> Paso de verdad en la entrega 4 de la serie de entrevistas: el primer benchmark daba 0,13 ns/op porque el compilador se lo habia comido entero.
