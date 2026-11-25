// Lección 6: go tool trace. El mismo trabajo en paralelo, con el candado en tres sitios.
//
// El servicio procesa por la noche los registros del día: 16 archivos, 8 goroutines.
// Cada goroutine decodifica líneas y suma en un mapa compartido.
//
//	go run ./06_trace -modo ancho -trace ancho.out
//	go tool trace -pprof=sync ancho.out > sync.prof && go tool pprof -top sync.prof
//	go tool trace -pprof=sched ancho.out > sched.prof && go tool pprof -top sched.prof
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"runtime"
	"runtime/trace"
	"slices"
	"sync"
	"time"

	"curso-go-rendimiento/datos"
)

type entrada struct {
	Metodo string  `json:"metodo"`
	Ruta   string  `json:"ruta"`
	Estado int     `json:"estado"`
	Ms     float64 `json:"ms"`
}

type suma struct {
	total, errores int
	ms             float64
}

func decodificar(linea []byte) (string, entrada) {
	var e entrada
	if err := json.Unmarshal(linea, &e); err != nil {
		panic(err)
	}
	return e.Metodo + " " + e.Ruta, e
}

func sumar(m map[string]*suma, k string, e entrada) {
	s := m[k]
	if s == nil {
		s = &suma{}
		m[k] = s
	}
	s.total++
	s.ms += e.Ms
	if e.Estado >= 500 {
		s.errores++
	}
}

func lineas(archivo []byte, f func([]byte)) {
	for _, l := range bytes.Split(bytes.TrimSpace(archivo), []byte("\n")) {
		f(l)
	}
}

func procesar(modo string, archivos [][]byte) map[string]*suma {
	trabajo := make(chan []byte)
	global := map[string]*suma{}
	var mu sync.Mutex
	var wg sync.WaitGroup
	for range 8 {
		wg.Go(func() {
			local := map[string]*suma{}
			for a := range trabajo {
				switch modo {
				case "ancho": // el candado envuelve también el trabajo caro
					lineas(a, func(l []byte) {
						mu.Lock()
						k, e := decodificar(l)
						sumar(global, k, e)
						mu.Unlock()
					})
				case "estrecho": // el candado solo envuelve la suma
					lineas(a, func(l []byte) {
						k, e := decodificar(l)
						mu.Lock()
						sumar(global, k, e)
						mu.Unlock()
					})
				case "local": // cada goroutine suma en su mapa; se juntan al final
					lineas(a, func(l []byte) {
						k, e := decodificar(l)
						sumar(local, k, e)
					})
				}
			}
			mu.Lock()
			for k, s := range local {
				g := global[k]
				if g == nil {
					g = &suma{}
					global[k] = g
				}
				g.total += s.total
				g.errores += s.errores
				g.ms += s.ms
			}
			mu.Unlock()
		})
	}
	for _, a := range archivos {
		trabajo <- a
	}
	close(trabajo)
	wg.Wait()
	return global
}

func main() {
	modo := flag.String("modo", "ancho", "ancho, estrecho o local")
	salida := flag.String("trace", "", "archivo donde escribir la traza")
	repetir := flag.Int("repetir", 1, "cuántas veces medir (se da la mediana)")
	flag.Parse()

	archivos := make([][]byte, 16)
	for i := range archivos {
		archivos[i] = datos.GenerarConSemilla(5000, uint64(100+i))
	}
	if *salida != "" {
		f, err := os.Create(*salida)
		if err != nil {
			panic(err)
		}
		defer f.Close()
		if err := trace.Start(f); err != nil {
			panic(err)
		}
		defer trace.Stop()
	}
	var ms []float64
	var r map[string]*suma
	for range *repetir {
		t0 := time.Now()
		r = procesar(*modo, archivos)
		ms = append(ms, float64(time.Since(t0).Microseconds())/1000)
	}
	slices.Sort(ms)
	total := 0
	for _, s := range r {
		total += s.total
	}
	fmt.Printf("modo %-8s  %d líneas · %d claves · %d repeticiones: mediana %5.1f ms, mínimo %5.1f ms  (GOMAXPROCS=%d)\n",
		*modo, total, len(r), *repetir, ms[len(ms)/2], ms[0], runtime.GOMAXPROCS(0))
}
