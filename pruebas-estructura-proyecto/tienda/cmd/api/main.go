// El main solo arma las piezas y arranca. La lógica vive en internal/.
package main

import (
	"fmt"

	"github.com/golandia/tienda/internal/almacen"
	"github.com/golandia/tienda/internal/pedidos"
)

func main() {
	s := pedidos.Servicio{A: &almacen.Memoria{}}
	p, err := s.Crear("teclado", 3)
	fmt.Println(p, err)
	_, err = s.Crear("teclado", 0)
	fmt.Println(err)
}
