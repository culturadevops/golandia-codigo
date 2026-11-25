//go:build ignore

// 3 · Leer el estado de la maquina SIN una sola dependencia.
// En Linux el nucleo publica su estado como ficheros de texto en /proc.
// Un agente de monitorizacion en Go es, literalmente, os.ReadFile y a contar.
package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

func main() {
	// --- carga del sistema --------------------------------------------------
	carga, err := os.ReadFile("/proc/loadavg")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	campos := strings.Fields(string(carga))
	fmt.Printf("carga (1, 5 y 15 minutos): %s  %s  %s\n", campos[0], campos[1], campos[2])
	fmt.Printf("procesos ejecutandose:     %s\n", campos[3])

	// --- memoria ------------------------------------------------------------
	mem, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	valores := map[string]int{}
	for _, linea := range strings.Split(string(mem), "\n") {
		partes := strings.Fields(linea)
		if len(partes) >= 2 {
			clave := strings.TrimSuffix(partes[0], ":")
			if n, err := strconv.Atoi(partes[1]); err == nil {
				valores[clave] = n // viene en kibibytes
			}
		}
	}
	total := valores["MemTotal"]
	libre := valores["MemAvailable"]
	fmt.Printf("memoria total:             %.1f GiB\n", float64(total)/1024/1024)
	fmt.Printf("memoria disponible:        %.1f GiB (%.0f%%)\n",
		float64(libre)/1024/1024, 100*float64(libre)/float64(total))

	// --- cuantos nucleos ve el proceso --------------------------------------
	nucleos, _ := os.ReadFile("/proc/cpuinfo")
	fmt.Printf("nucleos que ve el proceso: %d\n",
		strings.Count(string(nucleos), "processor\t:"))
}
