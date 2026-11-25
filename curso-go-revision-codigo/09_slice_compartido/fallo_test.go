package pedidos

import "testing"

func TestDosCotizacionesDelMismoCarrito(t *testing.T) {
	// len 3, cap 4: lo mismo que queda si el cliente añade los productos de uno en uno.
	carrito := make([]string, 0, 4)
	carrito = append(carrito, "taza", "playera", "gorra")
	normal := ConEnvio(carrito, "envío normal")
	expres := ConEnvio(carrito, "envío exprés")
	if normal[3] != "envío normal" {
		t.Fatalf("cap(carrito) = %d; la cotización normal dice %q: la pisó la exprés", cap(carrito), normal[3])
	}
	t.Logf("normal: %q · exprés: %q", normal[3], expres[3])
}
