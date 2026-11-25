// Leccion 4 — Mapas: el orden que Python te acostumbro a tener, aqui no existe.
package main

import (
	"fmt"
	"sort"
	"strings"
)

func orden(m map[string]int) string {
	var ks []string
	for k := range m {
		ks = append(ks, k)
	}
	return strings.Join(ks, " ")
}

func main() {
	srv := map[string]int{"web-01": 1, "web-02": 2, "web-03": 3, "db-01": 4, "cache-01": 5}

	fmt.Println("== 4.1 · Las cinco primeras vueltas al MISMO mapa ==")
	for i := 1; i <= 5; i++ {
		fmt.Printf("  vuelta %d: %s\n", i, orden(srv))
	}
	fmt.Println("  A veces salen iguales varias veces seguidas. Esa es la trampa:")
	fmt.Println("  PARECE estable, y no lo es.")

	fmt.Println()
	fmt.Println("== 4.2 · Mil vueltas: cuantos ordenes distintos salen ==")
	vistos := map[string]int{}
	for i := 0; i < 1000; i++ {
		vistos[orden(srv)]++
	}
	fmt.Printf("  en 1000 vueltas salieron %d ordenes distintos\n", len(vistos))
	fmt.Println("  Go baraja el orden A PROPOSITO para que nadie dependa de el.")
	fmt.Println("  Desde Python 3.7, un dict mantiene el orden de insercion.")
	fmt.Println("  Por eso un test que 'siempre pasa' en tu maquina falla un dia en CI.")

	fmt.Println()
	fmt.Println("== 4.3 · Si quieres orden, lo pides ==")
	claves := make([]string, 0, len(srv))
	for k := range srv {
		claves = append(claves, k)
	}
	sort.Strings(claves)
	fmt.Printf("  ordenadas: %s\n", strings.Join(claves, " "))
}
