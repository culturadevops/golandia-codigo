package resumen

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// La expresión regular se compila UNA vez, al cargar el paquete. Es el único cambio frente a
// V1, y es el que señala el perfil de CPU de la lección 2.
var reID = regexp.MustCompile(`/[0-9]+`)

func ResumenV2(cuerpo []byte) ([]byte, error) {
	latencias := map[string][]float64{}
	errores := map[string]int{}
	for _, linea := range strings.Split(string(cuerpo), "\n") {
		if strings.TrimSpace(linea) == "" {
			continue
		}
		var m map[string]any
		if err := json.Unmarshal([]byte(linea), &m); err != nil {
			return nil, fmt.Errorf("línea mal formada: %w", err)
		}
		metodo, _ := m["metodo"].(string)
		ruta, _ := m["ruta"].(string)
		ms, _ := m["ms"].(float64)
		estado, _ := m["estado"].(float64)

		ruta = reID.ReplaceAllString(ruta, "/:id")

		clave := fmt.Sprintf("%s %s", metodo, ruta)
		latencias[clave] = append(latencias[clave], ms)
		if estado >= 500 {
			errores[clave]++
		}
	}

	var filas []Fila
	for clave, ms := range latencias {
		partes := strings.SplitN(clave, " ", 2)
		orden := append([]float64(nil), ms...)
		sort.Float64s(orden)
		filas = append(filas, Fila{
			Metodo: partes[0], Ruta: partes[1], Total: len(ms), Errores: errores[clave],
			P50: percentil(orden, 50), P95: percentil(orden, 95),
		})
	}
	ordenar(filas)
	return json.Marshal(filas)
}
