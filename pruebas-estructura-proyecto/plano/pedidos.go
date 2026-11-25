package main

import "fmt"

type Pedido struct {
	Producto string
	Cantidad int
}

func NuevoPedido(producto string, cantidad int) Pedido { return Pedido{producto, cantidad} }

func (p Pedido) Resumen() string { return fmt.Sprintf("%d x %s", p.Cantidad, p.Producto) }
