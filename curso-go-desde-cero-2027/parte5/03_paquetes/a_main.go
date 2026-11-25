package main

import (
	"fmt"

	"tiendita/tienda"
)

func main() {
	fmt.Println("precio con impuesto:", tienda.Precio("café"))
}
