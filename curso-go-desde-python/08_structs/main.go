// Leccion 8 — No hay clases: structs y metodos. Y el metodo que "no hace nada".
package main

import "fmt"

type Servidor struct {
	Nombre  string
	Replicas int
}

// Receptor por VALOR: trabaja sobre una COPIA.
func (s Servidor) EscalarMal(n int) { s.Replicas = n }

// Receptor por PUNTERO: trabaja sobre el original.
func (s *Servidor) Escalar(n int) { s.Replicas = n }

func main() {
	fmt.Println("== 8.1 · Un struct es lo mas parecido a tu clase de Python ==")
	s := Servidor{Nombre: "api", Replicas: 1}
	fmt.Printf("  %+v\n", s)

	fmt.Println()
	fmt.Println("== 8.2 · LA TRAMPA: el metodo que no modifica ==")
	s.EscalarMal(5)
	fmt.Printf("  tras s.EscalarMal(5)  -> Replicas = %d\n", s.Replicas)
	fmt.Println("  Compila, no da error, y no cambio nada: el metodo recibio una copia.")
	fmt.Println("  En Python `self` es siempre el objeto de verdad. Aqui, depende del *.")

	s.Escalar(5)
	fmt.Printf("  tras s.Escalar(5)     -> Replicas = %d\n", s.Replicas)
	fmt.Println("  La unica diferencia es el * en (s *Servidor).")

	fmt.Println()
	fmt.Println("== 8.3 · Asignar un struct SI copia (al reves que una lista de Python) ==")
	a := Servidor{Nombre: "web", Replicas: 2}
	b := a
	b.Replicas = 9
	fmt.Printf("  a.Replicas=%d  b.Replicas=%d\n", a.Replicas, b.Replicas)
	fmt.Println("  En Python, b = a con un objeto seria otro nombre para lo mismo.")
}
