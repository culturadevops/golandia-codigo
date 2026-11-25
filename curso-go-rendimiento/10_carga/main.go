// Lección 10: el antes y el después, con tráfico. Levanta el servicio en este mismo proceso
// (o ataca uno que ya corre, con -url) y lo carga con varios clientes a la vez durante unos
// segundos. Cuenta también las respuestas que no coinciden con la esperada.
//
//	go run ./10_carga -version V1
//	go run ./10_carga -version V5 -clientes 8 -duracion 5s
//	go run ./10_carga -url http://127.0.0.1:8080/resumen -duracion 8s
package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"slices"
	"sync"
	"time"

	"curso-go-rendimiento/datos"
	"curso-go-rendimiento/resumen"
)

func main() {
	version := flag.String("version", "V5", "V1, V3, V4 o V5")
	url := flag.String("url", "", "atacar este servicio en vez de levantar uno")
	clientes := flag.Int("clientes", 8, "clientes a la vez")
	duracion := flag.Duration("duracion", 5*time.Second, "cuánto dura la carga")
	flag.Parse()

	registro := datos.Generar(5000)
	esperado, _ := resumen.ResumenV1(registro)
	if *url == "" {
		h := map[string]http.HandlerFunc{
			"V1": resumen.HandlerV1, "V3": resumen.HandlerV3,
			"V4": resumen.HandlerV4, "V5": resumen.HandlerV5,
		}[*version]
		if h == nil {
			fmt.Fprintln(os.Stderr, "versión desconocida:", *version)
			os.Exit(2)
		}
		srv := httptest.NewServer(h)
		defer srv.Close()
		*url = srv.URL + "/resumen"
	} else {
		*version = *url
	}

	cliente := &http.Client{Transport: &http.Transport{MaxIdleConnsPerHost: *clientes}}
	var (
		mu       sync.Mutex
		tiempos  []time.Duration
		malas    int
		fallidas int
		wg       sync.WaitGroup
	)
	fin := time.Now().Add(*duracion)
	for range *clientes {
		wg.Go(func() {
			var mios []time.Duration
			m, f := 0, 0
			for time.Now().Before(fin) {
				t0 := time.Now()
				r, err := cliente.Post(*url, "application/x-ndjson", bytes.NewReader(registro))
				if err != nil {
					f++
					continue
				}
				cuerpo, _ := io.ReadAll(r.Body)
				r.Body.Close()
				mios = append(mios, time.Since(t0))
				if r.StatusCode != 200 || !bytes.Equal(cuerpo, esperado) {
					m++
				}
			}
			mu.Lock()
			tiempos = append(tiempos, mios...)
			malas += m
			fallidas += f
			mu.Unlock()
		})
	}
	wg.Wait()
	slices.Sort(tiempos)
	pct := func(p float64) float64 {
		if len(tiempos) == 0 {
			return 0
		}
		return float64(tiempos[int(p/100*float64(len(tiempos)-1))].Microseconds()) / 1000
	}
	fmt.Printf("%s  %d clientes · %v: %5d peticiones  %6.1f pet/s  p50 %6.1f ms  p99 %6.1f ms  respuestas mal %d  errores %d\n",
		*version, *clientes, *duracion, len(tiempos), float64(len(tiempos))/duracion.Seconds(),
		pct(50), pct(99), malas, fallidas)
}
