package main

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
