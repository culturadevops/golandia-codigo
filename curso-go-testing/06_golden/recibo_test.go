// Lección 6 · Golden files con -update.
//
//	go test ./06_golden             -> compara tienda.Recibo con testdata/*.golden
//	go test ./06_golden -update     -> REESCRIBE los .golden con lo que salga ahora
//
// -update no comprueba nada: acepta lo que haya. Después, git diff testdata/ y se revisa.
package golden

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"curso-testing/tienda"
)

var actualizar = flag.Bool("update", false, "reescribe los archivos .golden")

var casos = []struct {
	nombre string
	p      tienda.Pedido
}{
	{"primer_pedido", tienda.Pedido{ID: 1, Cliente: "ana", Total: 18400, Estado: "pagado", CobroID: "cob-1",
		Items: []tienda.Item{{SKU: "TAZA", Cantidad: 1, Precio: 15000}}}},
	{"envio_gratis", tienda.Pedido{ID: 2, Cliente: "beto", Total: 65000, Estado: "pagado", CobroID: "cob-2",
		Items: []tienda.Item{{SKU: "TAZA", Cantidad: 3, Precio: 15000}, {SKU: "FILTRO", Cantidad: 10, Precio: 2000}}}},
	{"rechazado", tienda.Pedido{ID: 3, Cliente: "moroso", Total: 4999, Estado: "rechazado",
		Items: []tienda.Item{{SKU: "CHICLE", Cantidad: 1, Precio: 99}}}},
}

func TestRecibo(t *testing.T) {
	fecha := time.Date(2026, 10, 28, 12, 0, 0, 0, time.UTC)
	for _, tc := range casos {
		t.Run(tc.nombre, func(t *testing.T) {
			t.Parallel()
			p := tc.p
			p.Creado = fecha
			got := []byte(tienda.Recibo(&p))
			ruta := filepath.Join("testdata", tc.nombre+".golden")
			if *actualizar {
				if err := os.WriteFile(ruta, got, 0o644); err != nil {
					t.Fatal(err)
				}
			}
			want, err := os.ReadFile(ruta)
			if err != nil {
				t.Fatalf("%v (¿falta correr con -update?)", err)
			}
			if !bytes.Equal(got, want) {
				t.Errorf("%s no coincide (%d bytes, el golden %d):\n%s", ruta, len(got), len(want),
					diferencia(string(want), string(got)))
			}
		})
	}
}

// diferencia enseña solo las líneas que cambian: «-» lo del golden, «+» lo de ahora.
func diferencia(want, got string) string {
	w, g := strings.Split(want, "\n"), strings.Split(got, "\n")
	var b strings.Builder
	for i := range max(len(w), len(g)) {
		var lw, lg string
		if i < len(w) {
			lw = w[i]
		}
		if i < len(g) {
			lg = g[i]
		}
		if lw != lg {
			fmt.Fprintf(&b, "  línea %d\n    - %q\n    + %q\n", i+1, lw, lg)
		}
	}
	return b.String()
}
