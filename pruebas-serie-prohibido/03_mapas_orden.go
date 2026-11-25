// 03_mapas_orden.go — ¿cuánto de aleatorio es el recorrido de un mapa en Go?
//
// POR QUE EXISTE. La prueba corta (02_nil_mapas_slices.go) recorre un mapa de 5 claves tres
// veces y da por hecho que salen tres ordenes distintos. En la corrida del 2026-09-05 las
// vueltas 1 y 2 salieron IGUALES: con 5 claves hay 120 ordenes posibles y las coincidencias
// pasan. Si el video dice "siempre distinto", el titular es falso y la regla de la serie obliga
// a cambiarlo.
//
// Esto mide lo que de verdad se puede afirmar: cuantos ordenes distintos salen en N vueltas, y
// que ni siquiera el PRIMER elemento se repite.
//
// Correr:  go run 03_mapas_orden.go
package main

import (
	"fmt"
	"sort"
	"strings"
)

func recorrer(m map[string]int) string {
	partes := make([]string, 0, len(m))
	for k := range m {
		partes = append(partes, k)
	}
	return strings.Join(partes, " ")
}

func main() {
	m := map[string]int{"a": 1, "b": 2, "c": 3, "d": 4, "e": 5}

	fmt.Println("=== tres vueltas seguidas, mapa de 5 claves ===")
	for v := 1; v <= 3; v++ {
		fmt.Printf("  vuelta %d: %s\n", v, recorrer(m))
	}

	// Cuantos ordenes DISTINTOS salen en 1.000 vueltas.
	vistos := map[string]int{}
	primeros := map[string]int{}
	const N = 1000
	for i := 0; i < N; i++ {
		o := recorrer(m)
		vistos[o]++
		primeros[strings.Fields(o)[0]]++
	}
	fmt.Printf("\n=== %d vueltas sobre el MISMO mapa, sin tocarlo ===\n", N)
	fmt.Printf("  ordenes distintos vistos : %d  (de 120 posibles)\n", len(vistos))

	claves := make([]string, 0, len(primeros))
	for k := range primeros {
		claves = append(claves, k)
	}
	sort.Strings(claves)
	fmt.Println("\n  cuantas veces empezo por cada clave:")
	for _, k := range claves {
		fmt.Printf("    %s : %3d de %d\n", k, primeros[k], N)
	}

	// La prueba de que dos vueltas SEGUIDAS pueden coincidir: es aleatorio, no alterno.
	repes := 0
	ant := recorrer(m)
	for i := 0; i < N; i++ {
		o := recorrer(m)
		if o == ant {
			repes++
		}
		ant = o
	}
	fmt.Printf("\n  vueltas seguidas que salieron IGUALES: %d de %d\n", repes, N)
	fmt.Println("  -> es ALEATORIO, no 'siempre distinto'. El titular tiene que decir eso.")
}
