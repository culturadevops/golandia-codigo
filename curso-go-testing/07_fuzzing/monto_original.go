// Lección 7 · Fuzzing que encuentra un fallo real.
//
// Esta es la versión ORIGINAL de FormatearMonto y ParsearMonto, la que tenía la tienda antes
// del fuzzing. Pasaba sus pruebas de tabla (montos_test.go). La arreglada está en tienda/monto.go.
//
//	go test ./07_fuzzing                                    -> las pruebas de tabla: verde
//	go test -fuzz=FuzzIdaYVuelta -fuzztime=60s ./07_fuzzing  -> el fuzzing encuentra el fallo 1
//	go test -tags paso2 -fuzz=FuzzIdaYVuelta ./07_fuzzing    -> arreglado el 1, encuentra el 2
package fuzzing

import (
	"fmt"
	"strconv"
	"strings"
)

type Monto int64

func Parsear(s string) (Monto, error) {
	ent, dec, _ := strings.Cut(s, ",")
	e, err := strconv.ParseInt(ent, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("monto inválido %q", s)
	}
	var d int64
	if dec != "" {
		if d, err = strconv.ParseInt(dec, 10, 64); err != nil {
			return 0, fmt.Errorf("monto inválido %q", s)
		}
	}
	if e < 0 {
		return Monto(e*100 - d), nil
	}
	return Monto(e*100 + d), nil // fallo 2: «-0,50» tiene e == 0, y el signo se pierde
}
