package main

// Descuento aplica un porcentaje a un precio en centavos, redondeando hacia abajo.
func Descuento(precio, porcentaje int) int {
	if porcentaje > 100 {
		porcentaje = 100
	}
	return precio - precio*porcentaje/100
}

func main() {}
