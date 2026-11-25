// Package datos fabrica un registro de accesos de mentira, siempre el mismo.
//
// Cada línea es un objeto JSON, como lo dejaría un balanceador o un servicio que registra en
// JSON. La semilla es fija: dos ejecuciones dan exactamente los mismos bytes, así que los
// benchmarks de una versión y de otra procesan lo mismo.
package datos

import (
	"fmt"
	"math"
	"math/rand/v2"
	"strings"
)

var rutas = []string{
	"/api/v1/pedidos/%d",
	"/api/v1/pedidos/%d/items/%d",
	"/api/v1/clientes/%d",
	"/api/v1/clientes/%d/facturas",
	"/api/v1/productos",
	"/api/v1/productos/%d",
	"/api/v1/buscar",
	"/salud",
}

var metodos = []string{"GET", "GET", "GET", "GET", "POST", "PUT", "DELETE"}

var agentes = []string{
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/141.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 15_6) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/19.0 Safari/605.1.15",
	"curl/8.16.0",
	"Go-http-client/1.1",
}

// Generar devuelve n líneas de registro separadas por saltos de línea. Siempre las mismas.
func Generar(n int) []byte { return GenerarConSemilla(n, 686) }

// GenerarConSemilla da otro registro, igual de realista, para cada semilla.
func GenerarConSemilla(n int, semilla uint64) []byte {
	r := rand.New(rand.NewPCG(semilla, 2026))
	var b strings.Builder
	b.Grow(n * 230)
	for i := range n {
		plantilla := rutas[r.IntN(len(rutas))]
		var ruta string
		switch strings.Count(plantilla, "%d") {
		case 0:
			ruta = plantilla
		case 1:
			ruta = fmt.Sprintf(plantilla, 1000+r.IntN(90000))
		default:
			ruta = fmt.Sprintf(plantilla, 1000+r.IntN(90000), 1+r.IntN(40))
		}
		estado := 200
		switch x := r.IntN(100); {
		case x < 3:
			estado = 500
		case x < 5:
			estado = 503
		case x < 12:
			estado = 404
		}
		// Latencias con cola larga: casi todas rápidas, unas pocas muy lentas.
		ms := math.Round(math.Exp(2.3+0.8*r.NormFloat64())*10) / 10
		fmt.Fprintf(&b, `{"ts":"2026-10-28T10:%02d:%02d.%03dZ","metodo":%q,"ruta":%q,"estado":%d,"ms":%.1f,"ip":"10.0.%d.%d","ua":%q}`,
			(i/60)%60, i%60, r.IntN(1000), metodos[r.IntN(len(metodos))], ruta, estado, ms,
			r.IntN(8), 1+r.IntN(250), agentes[r.IntN(len(agentes))])
		b.WriteByte('\n')
	}
	return []byte(b.String())
}
