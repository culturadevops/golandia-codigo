// Lección 3: perfil de memoria. alloc_space contra inuse_space.
//
// El servicio tiene una función más: guarda la última línea con error de cada cliente (por IP)
// para el panel de soporte. Atiende 300 peticiones distintas y escribe un perfil del heap.
//
//	go run ./03_memoria -perfil heap.prof              # la versión con la fuga
//	go run ./03_memoria -perfil heap.prof -arreglado   # con strings.Clone
//	go tool pprof -sample_index=alloc_space -top heap.prof
//	go tool pprof -sample_index=inuse_space -top heap.prof
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"strings"

	"curso-go-rendimiento/datos"
	"curso-go-rendimiento/resumen"
)

// ultimoError: IP -> la línea completa de su último error 5xx.
var ultimoError = map[string]string{}

var arreglado = flag.Bool("arreglado", false, "copiar la línea con strings.Clone antes de guardarla")

func guardarErrores(cuerpo []byte) {
	texto := string(cuerpo) // una copia del cuerpo entero: 1,1 MB
	for _, linea := range strings.Split(texto, "\n") {
		if !strings.Contains(linea, `"estado":5`) {
			continue
		}
		var e struct {
			IP string `json:"ip"`
		}
		if json.Unmarshal([]byte(linea), &e) != nil {
			continue
		}
		if *arreglado {
			linea = strings.Clone(linea) // 230 bytes propios
		}
		// Sin Clone, linea es un trozo de texto: apunta DENTRO de la copia de 1,1 MB.
		// Mientras el mapa guarde ese trozo, el recolector no puede soltar el resto.
		ultimoError[e.IP] = linea
	}
}

func main() {
	perfil := flag.String("perfil", "heap.prof", "dónde escribir el perfil del heap")
	flag.Parse()
	runtime.MemProfileRate = 4096 // más muestras que el valor por defecto (512 KB)

	for i := range 300 {
		cuerpo := datos.GenerarConSemilla(5000, uint64(i+1))
		if _, err := resumen.ResumenV2(cuerpo); err != nil {
			panic(err)
		}
		guardarErrores(cuerpo)
	}

	runtime.GC() // inuse_space describe el heap en el último GC: que sea ahora
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("300 peticiones · %d clientes con error guardado · heap en uso tras el GC: %.1f MB\n",
		len(ultimoError), float64(m.HeapInuse)/(1<<20))
	fmt.Printf("asignado en total durante la prueba: %.0f MB\n", float64(m.TotalAlloc)/(1<<20))

	f, err := os.Create(*perfil)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	if err := pprof.Lookup("allocs").WriteTo(f, 0); err != nil {
		panic(err)
	}
}
