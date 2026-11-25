// 04_mapas_tamano.go — ¿el recorrido de un mapa es una BARAJA o solo un punto de partida?
//
// POR QUE EXISTE. 03_mapas_orden.go midio que un mapa de 5 claves solo produce 5 ordenes
// distintos en 1.000 vueltas, de 120 posibles, y que una clave sale primera el 51% de las
// veces. Eso NO es una permutacion al azar. Antes de decir en un video POR QUE, hay que medir
// como se comporta al crecer el mapa.
//
// Lo que se mide, sin suponer nada de la implementacion:
//   1. cuantos ordenes distintos salen en 2.000 vueltas, por tamano de mapa
//   2. si la secuencia que sale es una ROTACION de una secuencia fija (mismo ciclo, otro
//      arranque) o un orden nuevo cada vez
//
// El punto 2 es la prueba que decide el guion: si casi todas las vueltas son rotaciones de la
// misma secuencia, "aleatorio" significa "arranca por otro sitio", no "se baraja".
//
// Correr:  go run 04_mapas_tamano.go
package main

import (
	"fmt"
	"strings"
)

func recorrer(m map[int]int) []int {
	o := make([]int, 0, len(m))
	for k := range m {
		o = append(o, k)
	}
	return o
}

// ¿`b` es una rotacion de `a`? (misma secuencia circular, distinto arranque)
func esRotacion(a, b []int) bool {
	if len(a) != len(b) || len(a) == 0 {
		return false
	}
	doble := append(append([]int{}, a...), a...)
	for i := 0; i < len(a); i++ {
		igual := true
		for j := range b {
			if doble[i+j] != b[j] {
				igual = false
				break
			}
		}
		if igual {
			return true
		}
	}
	return false
}

func clave(o []int) string {
	p := make([]string, len(o))
	for i, v := range o {
		p[i] = fmt.Sprint(v)
	}
	return strings.Join(p, ",")
}

func main() {
	const N = 2000
	fmt.Printf("%-8s %-12s %-14s %s\n", "claves", "ordenes", "posibles", "% que son rotacion del 1º")
	for _, n := range []int{5, 8, 20, 100} {
		m := map[int]int{}
		for i := 0; i < n; i++ {
			m[i] = i
		}
		vistos := map[string]bool{}
		base := recorrer(m)
		rot := 0
		for i := 0; i < N; i++ {
			o := recorrer(m)
			vistos[clave(o)] = true
			if esRotacion(base, o) {
				rot++
			}
		}
		posibles := "muchos"
		if n <= 8 {
			f := 1
			for i := 2; i <= n; i++ {
				f *= i
			}
			posibles = fmt.Sprint(f)
		}
		fmt.Printf("%-8d %-12d %-14s %d%%\n", n, len(vistos), posibles, rot*100/N)
	}
	fmt.Println()
	fmt.Println("Si el % de rotaciones es alto, el recorrido NO se baraja:")
	fmt.Println("es la misma vuelta empezando por otro sitio.")
}
