package main

import (
	"fmt"
	"os"
	"runtime/pprof"
	"strings"
)

// Un trabajo realista: normalizar y contar palabras de un texto grande.
// Tiene un punto caliente escondido a proposito.
func normalizar(palabras []string) map[string]int {
	cuenta := map[string]int{}
	for _, p := range palabras {
		// EL PUNTO CALIENTE: se construye la clave concatenando en cada vuelta.
		clave := ""
		for _, r := range strings.ToLower(p) {
			if r >= 'a' && r <= 'z' {
				clave += string(r)
			}
		}
		cuenta[clave]++
	}
	return cuenta
}

func main() {
	f, _ := os.Create("cpu.pprof")
	pprof.StartCPUProfile(f)
	defer func() { pprof.StopCPUProfile(); f.Close() }()

	base := strings.Fields(strings.Repeat("El Rapido Zorro Marron Salta Sobre El Perro Perezoso ", 4000))

	// ⚠️ El perfilador de CPU muestrea a 100 Hz. Con una sola pasada el programa duraba 5 ms
	// y el perfil salia con «Total samples = 0»: cero muestras, nada que mirar. Se repite el
	// trabajo hasta pasar del segundo, que es lo minimo para que el `top` diga algo.
	var c map[string]int
	for i := 0; i < 400; i++ {
		c = normalizar(base)
	}
	fmt.Println("palabras distintas:", len(c))
}
