package main

import (
	"fmt"

	"tiendita/tienda"
)

func main() {
	fmt.Println("precio sin impuesto:", tienda.precioBase("café"))
}
