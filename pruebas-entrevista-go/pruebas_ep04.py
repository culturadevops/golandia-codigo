# -*- coding: utf-8 -*-
import sys
from pathlib import Path
sys.path.insert(0, str(Path(__file__).resolve().parent))
from probar import episodio

P1 = '''package main

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
'''

P2 = '''package main

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
'''

BENCH = '''package p03

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
'''

TAM = '''package main

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
'''

VET = '''package main

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
'''

P5 = r'''package main

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
'''


episodio(
    "04-punteros-y-receptores",
    "Episodio 4/8 — receptor por valor o por puntero",
    [
        {"nombre": "1. El metodo que no cambia nada",
         "pregunta": "Llamo al metodo tres veces y el campo sigue a cero. Por que?",
         "archivos": {"p01/main.go": P1},
         "correr": [(["go", "run", "./p01"], False)]},
        {"nombre": "2. Donde duele de verdad: la interfaz que no se cumple",
         "pregunta": "El metodo existe. Por que el compilador dice que no implementa la interfaz?",
         "archivos": {"p02/main.go": P2},
         "correr": [(["go", "build", "-o", "/tmp/salida-p02", "./p02"], True)],
         "nota": "Este es EL error que hay que reconocer de memoria en una entrevista: el conjunto de\nmetodos de T y el de *T no son el mismo."},
        {"nombre": "3. Cuanto cuesta de verdad copiar (medido, no supuesto)",
         "pregunta": "El receptor por valor es mas lento?",
         "archivos": {"p03/bench_test.go": BENCH, "p03tam/main.go": TAM},
         "correr": [(["go", "run", "./p03tam"], False),
                    (["go", "test", "-bench=.", "-benchmem", "-count=3", "./p03"], False)],
         "nota": "El numero honesto va en pantalla tal cual salga. Si la diferencia es pequena, se dice:\nla razon para usar puntero es la MUTACION y la consistencia, no la velocidad."},
        {"nombre": "4. La trampa silenciosa: copiar un mutex",
         "pregunta": "Este codigo compila y corre. Que tiene de malo?",
         "archivos": {"p04/main.go": VET},
         "correr": [(["go", "vet", "./p04"], True), (["go", "run", "./p04"], False)],
         "nota": "Compila y funciona: por eso es peligroso. Lo caza `go vet`, no el compilador.\nCada llamada bloquea una copia distinta del mutex, o sea no bloquea nada."},
        {"nombre": "5. Copiar el struct no copia lo que hay dentro",
         "pregunta": "Si paso el struct por valor, esta a salvo todo lo de dentro?",
         "archivos": {"p05/main.go": P5},
         "correr": [(["go", "run", "./p05"], False)],
         "nota": "En Go TODO se pasa por valor. Con un slice dentro, lo copiado es la cabecera, y la cabecera apunta al mismo array. Enlaza directo con el episodio 1."},
    ])
