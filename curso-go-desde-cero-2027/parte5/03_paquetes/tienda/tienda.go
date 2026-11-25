package tienda

// Precio es exportado: empieza con mayúscula.
func Precio(producto string) int {
	return precioBase(producto) * 116 / 100
}

// precioBase no se exporta: empieza con minúscula.
func precioBase(producto string) int {
	switch producto {
	case "café":
		return 250
	case "arepa":
		return 400
	}
	return 0
}
