// Video 205 — Go no tiene herencia. Tiene EMPOTRADO, y NO es lo mismo.
//
// LA DEMOSTRACION QUE DECIDE EL VIDEO: en herencia de verdad, un metodo de la clase base que
// llama a otro metodo acaba en la version REDEFINIDA de la hija (despacho virtual). Con el
// empotrado de Go, NO: el tipo empotrado no sabe que esta empotrado, asi que llama a SU propio
// metodo. Eso es lo que separa "parece herencia" de "es herencia".
//
// Y el segundo bloque ensena la salida buena: pasar la dependencia como INTERFAZ.
//
//	go run .
package main

import "fmt"

// ─────────────────────────── 1 · el empotrado NO despacha a la hija ──

type Base struct{ nombre string }

func (b Base) Saludar() string { return "hola, soy " + b.Quien() }
func (b Base) Quien() string   { return "la base (" + b.nombre + ")" }

type Hija struct {
	Base // empotrado: parece que hereda
}

// La hija REDEFINE Quien(). En un lenguaje con herencia, Saludar() usaria esta.
func (h Hija) Quien() string { return "la HIJA (" + h.nombre + ")" }

// ─────────────────────────── 2 · la salida buena: composicion con interfaz ──

type Identidad interface{ Quien() string }

type Saludador struct{ id Identidad }

func (s Saludador) Saludar() string { return "hola, soy " + s.id.Quien() }

type Persona struct{ nombre string }

func (p Persona) Quien() string { return "la persona (" + p.nombre + ")" }

func main() {
	fmt.Println("=== 1 · el empotrado parece herencia, y no lo es ===")
	h := Hija{Base{"ana"}}

	fmt.Println("  h.Quien()   =", h.Quien())
	fmt.Println("  h.Saludar() =", h.Saludar())
	fmt.Println()
	fmt.Println("  Saludar() esta en Base y llama a Quien().")
	fmt.Println("  Con herencia de verdad diria 'la HIJA'. Aqui dice 'la base'.")
	fmt.Println("  -> Base no sabe que esta empotrada. Llama a SU propio Quien().")

	fmt.Println()
	fmt.Println("=== 2 · la composicion con interfaz SI hace lo que querias ===")
	s := Saludador{id: Persona{"ana"}}
	fmt.Println("  s.Saludar() =", s.Saludar())
	fmt.Println("  -> aqui no hay magia: le pasas QUIEN tiene que responder.")

	fmt.Println()
	fmt.Println("=== 3 · y el empotrado sirve para lo que SI sirve: reutilizar ===")
	fmt.Println("  h.nombre =", h.nombre, " <- el campo de Base, accesible sin escribir .Base")
	fmt.Println("  h.Base.Quien() =", h.Base.Quien(), " <- y el original sigue ahi")
}
