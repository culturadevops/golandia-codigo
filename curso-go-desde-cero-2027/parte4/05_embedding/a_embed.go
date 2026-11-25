package main

import (
	"fmt"
	"time"
)

type Registro struct {
	Creado time.Time
}

func (r Registro) Antiguedad(ahora time.Time) time.Duration { return ahora.Sub(r.Creado) }

type Producto struct {
	Registro
	Nombre string
	Precio int
}

func (p Producto) String() string {
	return fmt.Sprintf("%s ($%d.%02d)", p.Nombre, p.Precio/100, p.Precio%100)
}

func main() {
	base := time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC)
	p := Producto{Registro: Registro{Creado: base}, Nombre: "café", Precio: 250}
	fmt.Println(p)
	fmt.Println("antigüedad:", p.Antiguedad(base.Add(36*time.Hour)))
	fmt.Println("creado:", p.Creado.Format("2006-01-02"))
	fmt.Printf("%v · %+v\n", p, p.Registro)
}
