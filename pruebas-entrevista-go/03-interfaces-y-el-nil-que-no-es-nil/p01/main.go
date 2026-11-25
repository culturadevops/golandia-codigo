package main

import "fmt"

type MiError struct{ Codigo int }

func (e *MiError) Error() string { return fmt.Sprintf("fallo %d", e.Codigo) }

// El error esta AQUI, y parece correcto.
func hacerAlgo(fallar bool) error {
	var err *MiError // nil
	if fallar {
		err = &MiError{Codigo: 500}
	}
	return err // devuelve la variable, no nil literal
}

func main() {
	err := hacerAlgo(false) // le decimos que NO falle
	fmt.Println("err vale:", err)
	if err != nil {
		fmt.Println(">> ENTRO EN EL IF. No fallo nada y estamos tratando un error.")
	} else {
		fmt.Println(">> todo bien")
	}
}
