// Lección 4: análisis de escape. Mira lo que decide el compilador con
//
//	go test -gcflags=-m ./04_escape 2>&1 | grep escape_test.go
package escape

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"

	"curso-go-rendimiento/datos"
)

type entrada struct {
	Metodo string  `json:"metodo"`
	Ruta   string  `json:"ruta"`
	Estado int     `json:"estado"`
	Ms     float64 `json:"ms"`
}

var lineas = bytes.Split(bytes.TrimSpace(datos.Generar(1000)), []byte("\n"))

// e dentro del bucle: &e va a parar a un parámetro any, el compilador no puede probar que
// no sobreviva a la llamada, y la mueve al heap. Una asignación por línea.
func decodificarDentro(lineas [][]byte) (n int) {
	for _, l := range lineas {
		var e entrada
		if json.Unmarshal(l, &e) == nil {
			n += e.Estado
		}
	}
	return n
}

// e fuera del bucle: también escapa, pero una sola vez.
func decodificarFuera(lineas [][]byte) (n int) {
	var e entrada
	for _, l := range lineas {
		e = entrada{}
		if json.Unmarshal(l, &e) == nil {
			n += e.Estado
		}
	}
	return n
}

func BenchmarkDecodificar(b *testing.B) {
	b.Run("e=dentro", func(b *testing.B) {
		for b.Loop() {
			decodificarDentro(lineas)
		}
	})
	b.Run("e=fuera", func(b *testing.B) {
		for b.Loop() {
			decodificarFuera(lineas)
		}
	})
}

// Tres formas de armar la clave "GET /api/v1/pedidos/:id" y buscarla en un mapa.
var conteo = map[string]int{"GET /api/v1/pedidos/:id": 0}

func claveSprintf(metodo, ruta string) int { return conteo[fmt.Sprintf("%s %s", metodo, ruta)] }

func claveConcat(metodo, ruta string) int { return conteo[metodo+" "+ruta] }

func claveBuffer(metodo, ruta string) int {
	var buf [128]byte
	k := append(append(append(buf[:0], metodo...), ' '), ruta...)
	return conteo[string(k)] // conteo[string(bytes)] no crea el string
}

var metodo, ruta = "GET", "/api/v1/pedidos/:id"

func BenchmarkClave(b *testing.B) {
	b.Run("forma=Sprintf", func(b *testing.B) {
		for b.Loop() {
			claveSprintf(metodo, ruta)
		}
	})
	b.Run("forma=concatenar", func(b *testing.B) {
		for b.Loop() {
			claveConcat(metodo, ruta)
		}
	})
	b.Run("forma=bufer_en_pila", func(b *testing.B) {
		for b.Loop() {
			claveBuffer(metodo, ruta)
		}
	})
}
