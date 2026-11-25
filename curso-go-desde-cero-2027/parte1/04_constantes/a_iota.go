package main

import "fmt"

type Estado int

const (
	Pendiente Estado = iota
	Enviado
	Entregado
	Devuelto
)

const (
	_  = iota
	KB = 1 << (10 * iota)
	MB
	GB
)

func main() {
	fmt.Println(Pendiente, Enviado, Entregado, Devuelto)
	fmt.Println(KB, MB, GB)
}
