//go:build linux

// Lección 7: el recolector, medido. GOGC y GOMEMLIMIT cambian dos cosas a la vez: cuánta
// memoria usa el proceso y cuánta CPU gasta recogiendo. Este programa lo mide.
//
// El servicio tiene unos 64 MB vivos (una caché que no se suelta) y atiende 150 peticiones que
// generan basura (ResumenV2: unos 6 MB por petición).
//
//	GOGC=100 go run ./07_gc
//	GOGC=off GOMEMLIMIT=150MiB go run ./07_gc
//	GODEBUG=gctrace=1 go run ./07_gc
//
// Solo Linux: lee la memoria pico de /proc/self/status (VmHWM) y la CPU de getrusage.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"runtime"
	"runtime/debug"
	"strings"
	"syscall"
	"time"

	"curso-go-rendimiento/datos"
	"curso-go-rendimiento/resumen"
)

// La caché: un millón de nodos pequeños con punteros, como un mapa de sesiones o de rutas.
// Con punteros importa: el recolector tiene que recorrerlos todos en cada ciclo.
type nodo struct {
	sig   *nodo
	clave string
	dato  [40]byte
}

var cache []*nodo

func picoRSS() string {
	f, err := os.Open("/proc/self/status")
	if err != nil {
		return "?"
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	for s.Scan() {
		if v, ok := strings.CutPrefix(s.Text(), "VmHWM:"); ok {
			var kb int
			fmt.Sscan(v, &kb)
			return fmt.Sprintf("%4d MB", kb/1024)
		}
	}
	return "?"
}

func cpu() time.Duration {
	var r syscall.Rusage
	syscall.Getrusage(syscall.RUSAGE_SELF, &r)
	return time.Duration(r.Utime.Nano() + r.Stime.Nano())
}

func main() {
	vivo := flag.Int("vivo", 64, "MB vivos todo el tiempo (la caché)")
	peticiones := flag.Int("peticiones", 150, "peticiones a atender")
	flag.Parse()

	cuerpo := datos.Generar(5000)
	cache = make([]*nodo, 0, *vivo*(1<<20)/64)
	for i := range cap(cache) { // ~64 bytes por nodo
		n := &nodo{clave: "s"}
		n.dato[0] = byte(i)
		if len(cache) > 0 {
			n.sig = cache[len(cache)-1]
		}
		cache = append(cache, n)
	}
	runtime.GC()
	// Pone a cero el pico de memoria (VmHWM): se mide solo mientras atiende peticiones.
	os.WriteFile("/proc/self/clear_refs", []byte("5"), 0)
	var antes runtime.MemStats
	runtime.ReadMemStats(&antes)
	c0, t0 := cpu(), time.Now()

	for range *peticiones {
		if _, err := resumen.ResumenV2(cuerpo); err != nil {
			panic(err)
		}
	}

	pared, gasto := time.Since(t0), cpu()-c0
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	gogc := os.Getenv("GOGC")
	if gogc == "" {
		gogc = "100"
	}
	limite := "sin límite"
	if l := debug.SetMemoryLimit(-1); l != 1<<63-1 {
		limite = fmt.Sprintf("%d MiB", l>>20)
	}
	fmt.Printf("GOGC=%-4s GOMEMLIMIT=%-10s  GCs %4d  memoria pico %s  CPU %5.2f s  (pared %4.2f s)\n",
		gogc, limite, m.NumGC-antes.NumGC, picoRSS(), gasto.Seconds(), pared.Seconds())
	runtime.KeepAlive(cache)
}
