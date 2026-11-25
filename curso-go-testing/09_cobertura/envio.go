// Lección 9 · Cobertura: -coverprofile, -coverpkg, y por qué el 100 % no garantiza nada.
//
// La regla de negocio: envío GRATIS DESDE 500,00 (500,00 incluido). Esta versión tiene un
// fallo en la frontera y sus pruebas cubren el 100 % de las líneas.
package cobertura

type Monto int64

const (
	GratisDesde Monto = 50000
	Costo       Monto = 4900
)

func Envio(subtotal Monto) Monto {
	if subtotal > GratisDesde { // debería ser >=
		return 0
	}
	return Costo
}

// Etiqueta es lo que se pinta en el carrito.
func Etiqueta(subtotal Monto) string {
	if Envio(subtotal) == 0 {
		return "envío gratis"
	}
	return "envío 49,00"
}
