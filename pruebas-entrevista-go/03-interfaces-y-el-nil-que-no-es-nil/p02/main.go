package main

import "fmt"

type MiError struct{ Codigo int }

func (e *MiError) Error() string { return fmt.Sprintf("fallo %d", e.Codigo) }

func main() {
	var puntero *MiError // nil
	var iface error = puntero

	fmt.Println("el puntero es nil?          ", puntero == nil)
	fmt.Println("la interfaz es nil?         ", iface == nil, "  <-- LA TRAMPA")
	fmt.Println()
	fmt.Println("una interfaz son DOS palabras: (tipo, valor)")
	fmt.Printf("  tipo  = %T\n", iface)
	fmt.Printf("  valor = %v\n", iface)
	fmt.Println()

	var vacia error // nunca se le asigno nada
	fmt.Printf("interfaz de verdad vacia: tipo=%T valor=%v  nil? %v\n", vacia, vacia, vacia == nil)
}
