//go:build trampa

// Las trampas de la lección 2.
//
//	go test -tags trampa -v -run 'TestDefer|TestContextCancelado|TestSetenvParalelo' ./02_ayudantes
package ayudantes

import (
	"os"
	"path/filepath"
	"testing"
)

// defer en el padre + subtests paralelos: el padre RETORNA antes de que corran los subtests
// (t.Parallel los pausa hasta entonces), así que el defer cierra el archivo antes de tiempo.
func TestDefer(t *testing.T) {
	f, err := os.Create(filepath.Join(t.TempDir(), "eventos.log"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	for _, ev := range []string{"creado", "cobrado", "enviado"} {
		t.Run(ev, func(t *testing.T) {
			t.Parallel()
			if _, err := f.WriteString(ev + "\n"); err != nil {
				t.Errorf("escribiendo %q: %v", ev, err)
			}
		})
	}
}

// Usar t.Context() DENTRO de un Cleanup: ya está cancelado.
func TestContextCancelado(t *testing.T) {
	t.Cleanup(func() {
		t.Logf("en el Cleanup: t.Context().Err() = %v", t.Context().Err())
		if t.Context().Err() != nil {
			t.Error("la limpieza con t.Context() no llega a hacerse")
		}
	})
}

// t.Setenv cambia el proceso entero: no se deja combinar con t.Parallel.
func TestSetenvParalelo(t *testing.T) {
	t.Parallel()
	t.Setenv("TIENDA_MONEDA", "USD")
}
