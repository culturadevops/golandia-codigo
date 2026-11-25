// huellas: calcula el sha256 de varios archivos con un grupo de workers.
//
//	huellas [-workers N] archivo...
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"os"
	"runtime"
	"sync"
	"time"
)

type resultado struct {
	indice int
	ruta   string
	huella string
	bytes  int64
	err    error
}

func huella(ruta string) (string, int64, error) {
	f, err := os.Open(ruta)
	if err != nil {
		return "", 0, err
	}
	defer f.Close()
	h := sha256.New()
	n, err := io.Copy(h, f)
	if err != nil {
		return "", n, err
	}
	return hex.EncodeToString(h.Sum(nil)), n, nil
}

func main() {
	workers := flag.Int("workers", runtime.NumCPU(), "cuántos archivos a la vez")
	flag.Parse()
	rutas := flag.Args()
	if len(rutas) == 0 {
		fmt.Fprintln(os.Stderr, "uso: huellas [-workers N] archivo...")
		os.Exit(2)
	}

	inicio := time.Now()
	trabajos := make(chan int)
	resultados := make([]resultado, len(rutas))
	var wg sync.WaitGroup
	for range *workers {
		wg.Go(func() {
			for i := range trabajos {
				h, n, err := huella(rutas[i])
				resultados[i] = resultado{i, rutas[i], h, n, err}
			}
		})
	}
	for i := range rutas {
		trabajos <- i
	}
	close(trabajos)
	wg.Wait()

	var total int64
	fallos := 0
	for _, r := range resultados {
		if r.err != nil {
			fmt.Fprintln(os.Stderr, "error:", r.err)
			fallos++
			continue
		}
		total += r.bytes
		fmt.Printf("%s…  %s\n", r.huella[:16], r.ruta)
	}
	fmt.Fprintf(os.Stderr, "%d archivos · %d MiB · %d workers · %d ms\n",
		len(rutas)-fallos, total>>20, *workers, time.Since(inicio).Milliseconds())
	if fallos > 0 {
		os.Exit(1)
	}
}
