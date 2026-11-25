// Cliente de carga para la reproducción del caso Discord: K trabajadores piden /leer sin parar durante D segundos
// y, cada segundo, imprimen peticiones, p50, p99 y la PEOR latencia de ese segundo. Así se ven los picos en el tiempo.
//   cliente <url base> <segundos> <trabajadores> [peticiones por segundo en total]
// Con el 4º argumento, cada trabajador va a ritmo fijo (como un servicio con tráfico normal, no saturado). 2026-10-19: sin ritmo,
// el cliente saturaba los 2 CPU del servidor y Docker lo frenaba en cada periodo de cuota: ~65 ms de máximo CADA segundo, hubiera
// o no recolección. Eso medía el límite de CPU, no el recolector.
package main

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"slices"
	"strconv"
	"sync"
	"time"
)

func main() {
	base := os.Args[1]
	seg, _ := strconv.Atoi(os.Args[2])
	k, _ := strconv.Atoi(os.Args[3])
	var cada time.Duration // intervalo entre peticiones de cada trabajador (0 = sin pausa)
	if len(os.Args) > 4 {
		rps, _ := strconv.Atoi(os.Args[4])
		cada = time.Duration(float64(time.Second) * float64(k) / float64(rps))
	}
	cli := &http.Client{Transport: &http.Transport{MaxIdleConnsPerHost: k}, Timeout: 30 * time.Second}

	var mu sync.Mutex
	porSeg := make([][]time.Duration, seg+1)
	inicio := time.Now()
	fin := inicio.Add(time.Duration(seg) * time.Second)
	var wg sync.WaitGroup
	for w := 0; w < k; w++ {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()
			r := rand.New(rand.NewSource(int64(w)))
			proxima := time.Now().Add(time.Duration(w) * cada / time.Duration(k))
			for time.Now().Before(fin) {
				if cada > 0 {
					time.Sleep(time.Until(proxima))
					proxima = proxima.Add(cada)
				}
				t0 := time.Now()
				resp, err := cli.Get(base + "/leer?u=" + strconv.FormatUint(r.Uint64()%10_000_000, 10))
				if err != nil {
					continue
				}
				io.Copy(io.Discard, resp.Body)
				resp.Body.Close()
				d := time.Since(t0)
				s := int(t0.Sub(inicio) / time.Second)
				mu.Lock()
				if s <= seg {
					porSeg[s] = append(porSeg[s], d)
				}
				mu.Unlock()
			}
		}(w)
	}
	wg.Wait()
	ms := func(d time.Duration) float64 { return float64(d.Microseconds()) / 1000 }
	for s, v := range porSeg[:seg] {
		if len(v) == 0 {
			fmt.Printf("t=%3ds sin respuestas\n", s)
			continue
		}
		slices.Sort(v)
		fmt.Printf("t=%3ds pet=%5d p50=%6.2fms p99=%7.2fms max=%7.2fms\n", s, len(v), ms(v[len(v)/2]), ms(v[len(v)*99/100]), ms(v[len(v)-1]))
	}
}
