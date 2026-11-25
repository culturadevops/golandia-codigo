package main

import "fmt"

type MiError struct{ Codigo int }

func (e *MiError) Error() string { return fmt.Sprintf("fallo %d", e.Codigo) }

func main() {
	// No es cosa de `error`: le pasa a CUALQUIER interfaz, incluida `any`.
	var p *MiError
	var cualquiera any = p
	fmt.Println("any con un puntero nil dentro == nil?", cualquiera == nil)

	// Y tambien con un slice o un mapa nil dentro.
	var s []int
	var m map[string]int
	var a1 any = s
	var a2 any = m
	fmt.Println("any con un slice nil dentro == nil? ", a1 == nil)
	fmt.Println("any con un mapa  nil dentro == nil? ", a2 == nil)
	fmt.Println()
	fmt.Println("Pero el slice de dentro SI es nil:", s == nil)
}
