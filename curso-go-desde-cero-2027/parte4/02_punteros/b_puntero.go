package main

import "fmt"

type Cuenta struct {
	Titular string
	Saldo   int
}

func (c *Cuenta) Depositar(monto int) {
	c.Saldo += monto
}

func duplicar(n *int) {
	*n = *n * 2
}

func main() {
	c := Cuenta{Titular: "Ana", Saldo: 100}
	c.Depositar(50)
	fmt.Println("con receptor puntero, saldo =", c.Saldo)

	x := 21
	p := &x
	duplicar(p)
	fmt.Println("x =", x, "· *p =", *p)

	var q *int
	fmt.Println("un puntero sin asignar es", q)
	fmt.Println(*q)
}
