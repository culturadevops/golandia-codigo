package tienda

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// Monto son centavos. El dinero nunca va en float64.
type Monto int64

func (m Monto) String() string { return FormatearMonto(m) }

// FormatearMonto escribe 1250 como "12,50" y -5 como "-0,05".
//
// Esta es la versión ARREGLADA. La original, con los dos fallos que encontró el fuzzing,
// está en 07_fuzzing/monto_original.go.
func FormatearMonto(m Monto) string {
	signo := ""
	u := uint64(m)
	if m < 0 {
		signo = "-"
		u = -u // en unsigned no se desborda ni con math.MinInt64
	}
	return fmt.Sprintf("%s%d,%02d", signo, u/100, u%100)
}

// ParsearMonto lee "12,50", "-0,05" o "7". Con decimales, siempre dos.
func ParsearMonto(s string) (Monto, error) {
	orig := s
	neg := strings.HasPrefix(s, "-")
	if neg {
		s = s[1:]
	}
	ent, dec, conDec := strings.Cut(s, ",")
	if !digitos(ent) || (conDec && (len(dec) != 2 || !digitos(dec))) {
		return 0, fmt.Errorf("monto inválido %q", orig)
	}
	e, err := strconv.ParseUint(ent, 10, 64)
	if err != nil || e > math.MaxUint64/100 {
		return 0, fmt.Errorf("monto inválido %q: fuera de rango", orig)
	}
	var d uint64
	if conDec {
		d, _ = strconv.ParseUint(dec, 10, 64)
	}
	t := e*100 + d
	limite := uint64(math.MaxInt64)
	if neg {
		limite++
	}
	if t < e*100 || t > limite {
		return 0, fmt.Errorf("monto inválido %q: fuera de rango", orig)
	}
	if neg {
		return Monto(int64(-t)), nil
	}
	return Monto(int64(t)), nil
}

func digitos(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
