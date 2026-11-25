package main

import (
	"fmt"
	"strconv"
)

func main() {
	stock, _ := strconv.Atoi("12 unidades")
	fmt.Println("stock:", stock)
	if stock == 0 {
		fmt.Println("¡sin stock! cancelo el pedido")
	}
}
