//go:build ignore

// 3 · El valor cero: una variable que no inicializas YA VALE ALGO, y ese algo es valido.
// El problema real no es la variable suelta: es no poder distinguir "no vino" de "vino un 0".
package main

import (
	"encoding/json"
	"fmt"
)

type Pedido struct {
	Cliente  string `json:"cliente"`
	Descuento int   `json:"descuento"`
}

type PedidoBien struct {
	Cliente   string `json:"cliente"`
	Descuento *int   `json:"descuento"` // el puntero SI distingue
}

func main() {
	var intentos int
	var nombre string
	var lista []int
	fmt.Printf("sin inicializar -> intentos=%d nombre=%q lista=%v (nil? %v)\n",
		intentos, nombre, lista, lista == nil)

	sinCampo := []byte(`{"cliente":"Ada"}`)
	conCero := []byte(`{"cliente":"Ada","descuento":0}`)

	var a, b Pedido
	_ = json.Unmarshal(sinCampo, &a)
	_ = json.Unmarshal(conCero, &b)
	fmt.Printf("sin el campo -> descuento=%d\n", a.Descuento)
	fmt.Printf("con cero     -> descuento=%d\n", b.Descuento)
	fmt.Printf("son iguales? %v   <- y son dos cosas DISTINTAS\n", a.Descuento == b.Descuento)

	var c, d PedidoBien
	_ = json.Unmarshal(sinCampo, &c)
	_ = json.Unmarshal(conCero, &d)
	fmt.Printf("con puntero: sin el campo -> %v (nil? %v)\n", c.Descuento, c.Descuento == nil)
	fmt.Printf("con puntero: con cero     -> %d (nil? %v)\n", *d.Descuento, d.Descuento == nil)
}
