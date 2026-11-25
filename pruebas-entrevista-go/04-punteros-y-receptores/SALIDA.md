# Pruebas ejecutadas — Episodio 4/8 — receptor por valor o por puntero

> Salida **literal** de esta máquina. La genera `probar.py` y se vuelve a generar sola.
> ⛔ Lo que no esté aquí no puede salir en pantalla (canon: ninguna cifra inventada,
> y lo que se muestra como código se ejecuta antes).

**Entorno citable en el vídeo:** contenedor `golang:1.26` · `go version go1.26.7 linux/amd64`

---

## 1. El metodo que no cambia nada

**La pregunta:** Llamo al metodo tres veces y el campo sigue a cero. Por que?

`p01/main.go`

```go
package main

import "fmt"

type Contador struct{ N int }

func (c Contador) SumaValor()   { c.N++ } // recibe una COPIA
func (c *Contador) SumaPuntero() { c.N++ }

func main() {
	c := Contador{}
	for i := 0; i < 3; i++ {
		c.SumaValor()
	}
	fmt.Println("tras 3 SumaValor():  ", c.N, "  <-- no cambio nada")

	for i := 0; i < 3; i++ {
		c.SumaPuntero()
	}
	fmt.Println("tras 3 SumaPuntero():", c.N)
}
```

```console
$ go run ./p01
tras 3 SumaValor():   0   <-- no cambio nada
tras 3 SumaPuntero(): 3
[codigo de salida: 0]
```

---

## 2. Donde duele de verdad: la interfaz que no se cumple

**La pregunta:** El metodo existe. Por que el compilador dice que no implementa la interfaz?

`p02/main.go`

```go
package main

import "fmt"

type Guardable interface{ Guardar() error }

type Usuario struct{ Nombre string }

// El metodo esta en *Usuario, no en Usuario.
func (u *Usuario) Guardar() error { return nil }

func persistir(g Guardable) { fmt.Println("guardando", g) }

func main() {
	u := Usuario{Nombre: "ana"}
	persistir(u) // <-- pasamos el VALOR, no el puntero
}
```

```console
$ go build -o /tmp/salida-p02 ./p02   # ⚠️ falla A PROPÓSITO
# pruebas/p02
p02/main.go:16:12: cannot use u (variable of struct type Usuario) as Guardable value in argument to persistir: Usuario does not implement Guardable (method Guardar has pointer receiver)
[codigo de salida: 1]
```

> Este es EL error que hay que reconocer de memoria en una entrevista: el conjunto de
> metodos de T y el de *T no son el mismo.

---

## 3. Cuanto cuesta de verdad copiar (medido, no supuesto)

**La pregunta:** El receptor por valor es mas lento?

`p03/bench_test.go`

```go
package p03

import "testing"

// Un struct realista de API: no es enorme, pero tampoco dos enteros.
type Pedido struct {
	ID           int64
	Cliente      string
	Direccion    string
	Lineas       [16]int64
	Total        float64
	Notas        string
	Etiquetas    [8]string
	Observacion  [128]byte
}

//go:noinline
func (p Pedido) TotalValor() float64 { return p.Total }

//go:noinline
func (p *Pedido) TotalPuntero() float64 { return p.Total }

var sink float64

func BenchmarkReceptorValor(b *testing.B) {
	p := Pedido{Total: 42}
	for i := 0; i < b.N; i++ {
		sink = p.TotalValor()
	}
}

func BenchmarkReceptorPuntero(b *testing.B) {
	p := Pedido{Total: 42}
	for i := 0; i < b.N; i++ {
		sink = p.TotalPuntero()
	}
}
```

`p03tam/main.go`

```go
package main

import (
	"fmt"
	"unsafe"
)

type Pedido struct {
	ID          int64
	Cliente     string
	Direccion   string
	Lineas      [16]int64
	Total       float64
	Notas       string
	Etiquetas   [8]string
	Observacion [128]byte
}

func main() {
	fmt.Println("tamano del struct Pedido:", unsafe.Sizeof(Pedido{}), "bytes")
	fmt.Println("tamano de un puntero:    ", unsafe.Sizeof(&Pedido{}), "bytes")
}
```

```console
$ go run ./p03tam
tamano del struct Pedido: 448 bytes
tamano de un puntero:     8 bytes
[codigo de salida: 0]
```

```console
$ go test -bench=. -benchmem -count=3 ./p03
goos: linux
goarch: amd64
pkg: pruebas/p03
cpu: Intel(R) Core(TM) Ultra 7 270K Plus
BenchmarkReceptorValor-24      	346153282	         3.464 ns/op	       0 B/op	       0 allocs/op
BenchmarkReceptorValor-24      	348139666	         3.432 ns/op	       0 B/op	       0 allocs/op
BenchmarkReceptorValor-24      	349844726	         3.464 ns/op	       0 B/op	       0 allocs/op
BenchmarkReceptorPuntero-24    	1000000000	         0.4277 ns/op	       0 B/op	       0 allocs/op
BenchmarkReceptorPuntero-24    	1000000000	         0.4302 ns/op	       0 B/op	       0 allocs/op
BenchmarkReceptorPuntero-24    	1000000000	         0.4288 ns/op	       0 B/op	       0 allocs/op
PASS
ok  	pruebas/p03	6.086s
[codigo de salida: 0]
```

> El numero honesto va en pantalla tal cual salga. Si la diferencia es pequena, se dice:
> la razon para usar puntero es la MUTACION y la consistencia, no la velocidad.

---

## 4. La trampa silenciosa: copiar un mutex

**La pregunta:** Este codigo compila y corre. Que tiene de malo?

`p04/main.go`

```go
package main

import (
	"fmt"
	"sync"
)

type Cache struct {
	mu     sync.Mutex
	datos  map[string]string
}

// El error: recibe el Cache POR VALOR, y eso copia el mutex.
func (c Cache) Leer(k string) string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.datos[k]
}

func main() {
	c := Cache{datos: map[string]string{"a": "1"}}
	fmt.Println(c.Leer("a"))
}
```

```console
$ go vet ./p04   # ⚠️ falla A PROPÓSITO
p04/main.go:14:9: Leer passes lock by value: pruebas/p04.Cache contains sync.Mutex
[codigo de salida: 1]
```

```console
$ go run ./p04
1
[codigo de salida: 0]
```

> Compila y funciona: por eso es peligroso. Lo caza `go vet`, no el compilador.
> Cada llamada bloquea una copia distinta del mutex, o sea no bloquea nada.

---

## 5. Copiar el struct no copia lo que hay dentro

**La pregunta:** Si paso el struct por valor, esta a salvo todo lo de dentro?

`p05/main.go`

```go
package main

import "fmt"

type Informe struct {
	Titulo string
	Filas  []int
}

// Recibe el struct POR VALOR. Se copia entero... o eso parece.
func tocar(inf Informe) {
	inf.Titulo = "cambiado"  // esto no sale de aqui
	inf.Filas[0] = 999       // esto SI sale
}

func main() {
	inf := Informe{Titulo: "original", Filas: []int{1, 2, 3}}
	tocar(inf)
	fmt.Println("Titulo:", inf.Titulo, " <- la copia protegio el campo")
	fmt.Println("Filas: ", inf.Filas, "  <- pero NO el contenido del slice")
	fmt.Println()
	fmt.Println("En Go TODO se pasa por valor. Lo que se copio fue la cabecera")
	fmt.Println("del slice, y esa cabecera apunta al mismo array.")
}
```

```console
$ go run ./p05
Titulo: original  <- la copia protegio el campo
Filas:  [999 2 3]   <- pero NO el contenido del slice

En Go TODO se pasa por valor. Lo que se copio fue la cabecera
del slice, y esa cabecera apunta al mismo array.
[codigo de salida: 0]
```

> En Go TODO se pasa por valor. Con un slice dentro, lo copiado es la cabecera, y la cabecera apunta al mismo array. Enlaza directo con el episodio 1.
