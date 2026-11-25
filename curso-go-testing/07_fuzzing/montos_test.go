package fuzzing

import "testing"

// Las pruebas de tabla de siempre. Todas en verde con la versión original.
func TestFormatear(t *testing.T) {
	for _, c := range []struct {
		m    Monto
		want string
	}{{1250, "12,50"}, {5, "0,05"}, {0, "0,00"}, {100000, "1000,00"}} {
		if got := Formatear(c.m); got != c.want {
			t.Errorf("Formatear(%d) = %q; quiero %q", c.m, got, c.want)
		}
	}
}

func TestParsear(t *testing.T) {
	for _, c := range []struct {
		s    string
		want Monto
	}{{"12,50", 1250}, {"0,05", 5}, {"-3,25", -325}, {"7", 700}} {
		if got, err := Parsear(c.s); err != nil || got != c.want {
			t.Errorf("Parsear(%q) = %d, %v; quiero %d", c.s, got, err, c.want)
		}
	}
	if _, err := Parsear("doce"); err == nil {
		t.Error(`Parsear("doce") no dio error`)
	}
}

// La propiedad, en vez de ejemplos: todo monto que formateo lo tengo que poder leer igual.
func FuzzIdaYVuelta(f *testing.F) {
	f.Add(int64(1250))
	f.Add(int64(0))
	f.Fuzz(func(t *testing.T, n int64) {
		s := Formatear(Monto(n))
		m, err := Parsear(s)
		if err != nil {
			t.Fatalf("Parsear(Formatear(%d)) = Parsear(%q): %v", n, s, err)
		}
		if m != Monto(n) {
			t.Fatalf("ida y vuelta: %d -> %q -> %d", n, s, m)
		}
	})
}
