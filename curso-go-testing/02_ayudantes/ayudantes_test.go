// Lección 2 · t.Cleanup, t.TempDir, t.Setenv y t.Context.
package ayudantes

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// t.TempDir: un directorio nuevo por test, que se borra solo al terminar.
func TestTempDir(t *testing.T) {
	var dir string
	t.Run("usa", func(t *testing.T) {
		dir = t.TempDir()
		os.WriteFile(filepath.Join(dir, "pedidos.csv"), []byte("1,ana,12,50\n"), 0o600)
		_, err := os.Stat(dir)
		t.Logf("dentro del test:  existe=%v  %s", err == nil, dir)
	})
	_, err := os.Stat(dir)
	t.Logf("después del test: existe=%v", err == nil)
	if err == nil {
		t.Error("t.TempDir no se borró")
	}
}

// t.Setenv: pone la variable y la deja como estaba al terminar el test.
func TestSetenv(t *testing.T) {
	os.Setenv("TIENDA_MONEDA", "MXN")
	t.Run("cambia", func(t *testing.T) {
		t.Setenv("TIENDA_MONEDA", "USD")
		t.Logf("dentro:  TIENDA_MONEDA=%s", os.Getenv("TIENDA_MONEDA"))
	})
	t.Logf("después: TIENDA_MONEDA=%s", os.Getenv("TIENDA_MONEDA"))
	if os.Getenv("TIENDA_MONEDA") != "MXN" {
		t.Error("t.Setenv no restauró la variable")
	}
}

// t.Cleanup en un helper: quien llama no tiene que acordarse de cerrar nada.
func abrirLog(t *testing.T) *os.File {
	t.Helper()
	f, err := os.Create(filepath.Join(t.TempDir(), "eventos.log"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { f.Close() })
	return f
}

// Con t.Cleanup en el padre, el archivo se cierra cuando acaban TODOS los subtests,
// también los paralelos.
func TestCleanupConParalelos(t *testing.T) {
	f := abrirLog(t)
	for _, ev := range []string{"creado", "cobrado", "enviado"} {
		t.Run(ev, func(t *testing.T) {
			t.Parallel()
			if _, err := f.WriteString(ev + "\n"); err != nil {
				t.Errorf("escribiendo %q: %v", ev, err)
			}
		})
	}
}

// t.Context se cancela JUSTO ANTES de los Cleanup. Para limpiar con un context, usa uno
// que no se cancele con el test.
func TestContextEnCleanup(t *testing.T) {
	ctx := context.WithoutCancel(t.Context())
	t.Cleanup(func() {
		if err := ctx.Err(); err != nil {
			t.Errorf("limpiando: %v", err)
		}
	})
	t.Logf("durante el test: t.Context().Err() = %v", t.Context().Err())
}
