// Package resumen es el servicio del curso: recibe un registro de accesos (una línea JSON por
// petición) y devuelve, por método y ruta, cuántas peticiones hubo, cuántas fallaron y la
// latencia p50 y p95.
//
// Hay una versión por paso del curso. Todas devuelven exactamente los mismos bytes
// (lo comprueba TestTodasIguales), así que lo único que cambia entre ellas es el coste.
//
//	V1  la versión lenta, la que «funciona»                    (lección 1)
//	V2  la expresión regular compilada una vez                 (lección 2)
//	V3  sin map[string]any, sin Sprintf, sin copiar el cuerpo  (lección 4)
//	V4  V3 + el búfer de salida en un sync.Pool                (lección 5, con su carrera: lección 8)
//	V5  V4 arreglada: escribe antes de devolver el búfer       (lección 8)
package resumen

import (
	"math"
	"sort"
)

// Fila es una línea del resumen.
type Fila struct {
	Metodo  string  `json:"metodo"`
	Ruta    string  `json:"ruta"`
	Total   int     `json:"total"`
	Errores int     `json:"errores"`
	P50     float64 `json:"p50_ms"`
	P95     float64 `json:"p95_ms"`
}

// percentil de una lista YA ordenada, por el método del rango más cercano.
func percentil(ordenados []float64, p float64) float64 {
	if len(ordenados) == 0 {
		return 0
	}
	i := int(math.Ceil(p/100*float64(len(ordenados)))) - 1
	return ordenados[max(i, 0)]
}

// ordenar deja primero las rutas con más tráfico; a igualdad, por método y ruta.
func ordenar(filas []Fila) {
	sort.Slice(filas, func(i, j int) bool {
		a, b := filas[i], filas[j]
		if a.Total != b.Total {
			return a.Total > b.Total
		}
		if a.Ruta != b.Ruta {
			return a.Ruta < b.Ruta
		}
		return a.Metodo < b.Metodo
	})
}
