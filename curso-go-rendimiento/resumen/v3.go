package resumen

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
)

// entrada son solo los campos que usamos. Los demás (ts, ip, ua) el decodificador los salta
// sin crear nada, al revés que con map[string]any, que los guarda todos.
type entrada struct {
	Metodo string  `json:"metodo"`
	Ruta   string  `json:"ruta"`
	Estado int     `json:"estado"`
	Ms     float64 `json:"ms"`
}

type acumulado struct {
	metodo, ruta string
	ms           []float64
	errores      int
}

// ResumenV3: lo que señalaron el perfil de memoria y el análisis de escape (lección 4).
func ResumenV3(cuerpo []byte) ([]byte, error) {
	filas, err := agregar(cuerpo)
	if err != nil {
		return nil, err
	}
	return json.Marshal(filas)
}

// agregar recorre el cuerpo sin copiarlo a un string ni partirlo en un []string.
func agregar(cuerpo []byte) ([]Fila, error) {
	acc := make(map[string]*acumulado, 64)
	var buf [128]byte // la clave se arma aquí: no escapa, vive en la pila
	// e se declara FUERA del bucle: &e acaba en una interfaz (any), así que escapa al heap.
	// Dentro del bucle sería una asignación por línea; aquí es una por llamada (lección 4).
	var e entrada
	for len(cuerpo) > 0 {
		linea := cuerpo
		if i := bytes.IndexByte(cuerpo, '\n'); i >= 0 {
			linea, cuerpo = cuerpo[:i], cuerpo[i+1:]
		} else {
			cuerpo = nil
		}
		if len(bytes.TrimSpace(linea)) == 0 {
			continue
		}
		e = entrada{}
		if err := json.Unmarshal(linea, &e); err != nil {
			return nil, fmt.Errorf("línea mal formada: %w", err)
		}
		clave := append(buf[:0], e.Metodo...)
		clave = append(clave, ' ')
		clave = normalizar(clave, e.Ruta)

		a := acc[string(clave)] // buscar con string(bytes) no reserva memoria
		if a == nil {
			a = &acumulado{metodo: e.Metodo, ruta: string(clave[len(e.Metodo)+1:])}
			acc[string(clave)] = a // solo la primera vez que aparece la ruta
		}
		a.ms = append(a.ms, e.Ms)
		if e.Estado >= 500 {
			a.errores++
		}
	}

	filas := make([]Fila, 0, len(acc))
	for _, a := range acc {
		sort.Float64s(a.ms) // los datos son nuestros: se ordenan en su sitio, sin copia
		filas = append(filas, Fila{
			Metodo: a.metodo, Ruta: a.ruta, Total: len(a.ms), Errores: a.errores,
			P50: percentil(a.ms, 50), P95: percentil(a.ms, 95),
		})
	}
	ordenar(filas)
	return filas, nil
}

// normalizar hace lo mismo que la expresión regular `/[0-9]+` -> "/:id", byte a byte.
func normalizar(dst []byte, ruta string) []byte {
	for i := 0; i < len(ruta); i++ {
		if ruta[i] == '/' && i+1 < len(ruta) && ruta[i+1] >= '0' && ruta[i+1] <= '9' {
			dst = append(dst, "/:id"...)
			for i+1 < len(ruta) && ruta[i+1] >= '0' && ruta[i+1] <= '9' {
				i++
			}
			continue
		}
		dst = append(dst, ruta[i])
	}
	return dst
}
