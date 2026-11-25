// Reproducción del caso Discord (blog del 4-feb-2020, «Why Discord is switching from Go to Rust»): un servicio con una caché
// ENORME en memoria y POCA basura nueva por petición. Así el recolector casi nunca arranca por crecimiento del montón… y Go lo
// fuerza cada 2 minutos, recorriendo toda la caché. Compila con Go 1.9.2 (la versión de Discord) y con Go 1.26: nada de módulos,
// genéricos ni rutas con método.
//   MODO=punteros → map[uint64]*Estado, cada Estado con 8 *Lectura (como una caché «normal» de Go)
//   MODO=plano    → map[uint64]EstadoPlano con un array de 8 Lectura: sin un solo puntero, el recolector no lo recorre
//   N             → usuarios en la caché (por defecto 3.000.000)
package main

import (
	"fmt"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"sync"
	"time"
)

type Lectura struct {
	Canal   uint64
	Mensaje uint64
}

type Estado struct {
	Usuario  uint64
	Lecturas []*Lectura
}

type EstadoPlano struct {
	Usuario  uint64
	Lecturas [8]Lectura
}

const fragmentos = 64

type fragPunteros struct {
	mu sync.Mutex
	m  map[uint64]*Estado
}

type fragPlano struct {
	mu sync.Mutex
	m  map[uint64]EstadoPlano
}

var (
	conPunteros [fragmentos]fragPunteros
	sinPunteros [fragmentos]fragPlano
	plano       bool
	ok          = []byte("ok\n")
)

func main() {
	n := 3000000
	if v, err := strconv.Atoi(os.Getenv("N")); err == nil {
		n = v
	}
	plano = os.Getenv("MODO") == "plano"
	t0 := time.Now()
	for i := 0; i < fragmentos; i++ {
		conPunteros[i].m = map[uint64]*Estado{}
		sinPunteros[i].m = map[uint64]EstadoPlano{}
	}
	for u := uint64(0); u < uint64(n); u++ {
		f := u % fragmentos
		if plano {
			e := EstadoPlano{Usuario: u}
			for j := range e.Lecturas {
				e.Lecturas[j] = Lectura{Canal: u*8 + uint64(j)}
			}
			sinPunteros[f].m[u] = e
		} else {
			e := &Estado{Usuario: u, Lecturas: make([]*Lectura, 8)}
			for j := range e.Lecturas {
				e.Lecturas[j] = &Lectura{Canal: u*8 + uint64(j)}
			}
			conPunteros[f].m[u] = e
		}
	}
	runtime.GC()
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	modo := "punteros"
	if plano {
		modo = "plano"
	}
	fmt.Fprintf(os.Stderr, "LISTO %s · modo=%s · usuarios=%d · montón vivo=%d MiB · llenar la caché tardó %.1f s\n",
		runtime.Version(), modo, n, ms.HeapAlloc>>20, time.Since(t0).Seconds())

	http.HandleFunc("/leer", func(w http.ResponseWriter, r *http.Request) {
		u, _ := strconv.ParseUint(r.URL.Query().Get("u"), 10, 64)
		u = u % uint64(n)
		f := u % fragmentos
		if plano {
			fr := &sinPunteros[f]
			fr.mu.Lock()
			e := fr.m[u]
			e.Lecturas[u%8].Mensaje++
			fr.m[u] = e
			fr.mu.Unlock()
		} else {
			fr := &conPunteros[f]
			fr.mu.Lock()
			fr.m[u].Lecturas[u%8].Mensaje++
			fr.mu.Unlock()
		}
		w.Write(ok)
	})
	http.ListenAndServe(":8080", nil)
}
