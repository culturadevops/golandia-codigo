package tienda

import (
	"math"
	"testing"
)

func TestFormatearMonto(t *testing.T) {
	t.Parallel()
	casos := []struct {
		m    Monto
		want string
	}{
		{1250, "12,50"}, {5, "0,05"}, {-5, "-0,05"}, {-150, "-1,50"}, {0, "0,00"},
		{math.MinInt64, "-92233720368547758,08"},
	}
	for _, tc := range casos {
		t.Run(tc.want, func(t *testing.T) {
			t.Parallel()
			if got := FormatearMonto(tc.m); got != tc.want {
				t.Errorf("FormatearMonto(%d) = %q; quiero %q", tc.m, got, tc.want)
			}
		})
	}
}

// La versión arreglada aguanta el mismo fuzzing que rompió la original (lección 7).
func FuzzMontoIdaYVuelta(f *testing.F) {
	for _, n := range []int64{0, 1250, -5, math.MaxInt64, math.MinInt64} {
		f.Add(n)
	}
	f.Fuzz(func(t *testing.T, n int64) {
		s := FormatearMonto(Monto(n))
		m, err := ParsearMonto(s)
		if err != nil {
			t.Fatalf("ParsearMonto(%q): %v", s, err)
		}
		if m != Monto(n) {
			t.Fatalf("ida y vuelta: %d -> %q -> %d", n, s, m)
		}
	})
}
