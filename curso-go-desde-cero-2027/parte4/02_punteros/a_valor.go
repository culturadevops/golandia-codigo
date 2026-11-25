package main

import "fmt"

type Cuenta struct {
	Titular string
	Saldo   int
}

func (c Cuenta) Depositar(monto int) {
	c.Saldo += monto
	fmt.Println("  dentro de Depositar, saldo =", c.Saldo)
}

func main() {
	c := Cuenta{Titular: "Ana", Saldo: 100}
	c.Depositar(50)
	fmt.Println("después de Depositar, saldo =", c.Saldo)
}
