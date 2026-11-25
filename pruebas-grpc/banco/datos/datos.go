package datos

// Datos DETERMINISTAS: sin aleatoriedad, para que dos corridas den exactamente los mismos bytes.

import "fmt"

type Pedido struct {
	ID       int64  `json:"id"`
	Cliente  string `json:"cliente"`
	Articulo string `json:"articulo"`
	Unidades int32  `json:"unidades"`
	Centimos int64  `json:"centimos"`
	Estado   string `json:"estado"`
	Creado   string `json:"creado"`
}

var estados = []string{"pendiente", "enviado", "entregado", "devuelto"}

func Uno(i int) Pedido {
	return Pedido{
		ID:       int64(100000 + i),
		Cliente:  fmt.Sprintf("cliente-%05d", i%997),
		Articulo: fmt.Sprintf("articulo-%04d", i%1301),
		Unidades: int32(1 + i%7),
		Centimos: int64(199 + (i*37)%9800),
		Estado:   estados[i%len(estados)],
		Creado:   fmt.Sprintf("2026-08-%02dT%02d:%02d:00Z", 1+i%28, i%24, i%60),
	}
}

func Muchos(n int) []Pedido {
	xs := make([]Pedido, n)
	for i := range xs {
		xs[i] = Uno(i)
	}
	return xs
}
