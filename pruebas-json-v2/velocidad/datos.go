package velocidad

import "fmt"

type Item struct {
	ID       int               `json:"id"`
	Nombre   string            `json:"nombre"`
	Precio   float64           `json:"precio"`
	Etiquetas []string         `json:"etiquetas"`
	Atributos map[string]string `json:"atributos"`
	Activo   bool              `json:"activo"`
}

// Datos genera siempre el mismo JSON: 10.000 items.
func Datos() []byte {
	b := []byte("[")
	for i := 0; i < 10000; i++ {
		if i > 0 {
			b = append(b, ',')
		}
		b = fmt.Appendf(b, `{"id":%d,"nombre":"producto número %d","precio":%d.%02d,"etiquetas":["a%d","b%d","c"],"atributos":{"color":"rojo","talla":"M%d"},"activo":%t}`,
			i, i, i%500, i%100, i%7, i%11, i%3, i%2 == 0)
	}
	return append(b, ']')
}

// SinMapa es Item sin el campo map: para ver si la diferencia de Marshal viene de los mapas.
type SinMapa struct {
	ID        int      `json:"id"`
	Nombre    string   `json:"nombre"`
	Precio    float64  `json:"precio"`
	Etiquetas []string `json:"etiquetas"`
	Activo    bool     `json:"activo"`
}

func Quitar(v []Item) []SinMapa {
	r := make([]SinMapa, len(v))
	for i, x := range v {
		r[i] = SinMapa{x.ID, x.Nombre, x.Precio, x.Etiquetas, x.Activo}
	}
	return r
}
