// Leccion 7 — El nil que no es nil: la trampa que no existe en Python.
package main

import (
	"errors"
	"fmt"
)

type MiError struct{ Motivo string }

func (e *MiError) Error() string { return e.Motivo }

// Devuelve un *MiError nil... metido en un `error`.
func validar(ok bool) error {
	var e *MiError
	if !ok {
		e = &MiError{"puerto invalido"}
	}
	return e // ⛔ aqui esta el fallo
}

func validarBien(ok bool) error {
	if !ok {
		return &MiError{"puerto invalido"}
	}
	return nil // ✅ nil de verdad
}

func main() {
	fmt.Println("== 7.1 · Una validacion que va BIEN ==")
	err := validar(true)
	fmt.Printf("  err == nil  ->  %v\n", err == nil)
	fmt.Printf("  err         ->  %v\n", err)
	fmt.Println("  La validacion fue bien, y aun asi err NO es nil.")
	fmt.Println()
	fmt.Println("== 7.2 · Por que ==")
	fmt.Println("  Una interfaz (error lo es) guarda DOS cosas: el tipo y el valor.")
	fmt.Printf("  aqui guarda: tipo %T, valor nil\n", err)
	fmt.Println("  Solo es nil si las DOS son nil. El tipo no lo es.")
	fmt.Println("  En Python, None es None y ya: esto no tiene equivalente.")
	fmt.Println()
	fmt.Println("== 7.3 · El arreglo: devolver nil a secas ==")
	err = validarBien(true)
	fmt.Printf("  validarBien(true) == nil  ->  %v\n", err == nil)
	var me *MiError
	err = validarBien(false)
	fmt.Printf("  errors.As(err, &me)       ->  %v (%s)\n", errors.As(err, &me), me.Motivo)
}
