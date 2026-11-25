// medir lanza un programa muchas veces y da la mediana y el p90 del tiempo que tarda.
//
//	go run ./medir -n 300 ./bin/vigia -version
//
// Sirve para medir el ARRANQUE de una CLI: el comando tiene que hacer casi nada.
package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"slices"
	"time"
)

func main() {
	n := flag.Int("n", 300, "cuántas veces se lanza")
	flag.Parse()
	if flag.NArg() == 0 {
		fmt.Fprintln(os.Stderr, "uso: medir [-n 300] programa [args...]")
		os.Exit(2)
	}
	args := flag.Args()
	for i := 0; i < 10; i++ { // calentar la caché del sistema de archivos
		_ = exec.Command(args[0], args[1:]...).Run()
	}
	tiempos := make([]time.Duration, 0, *n)
	for i := 0; i < *n; i++ {
		cmd := exec.Command(args[0], args[1:]...)
		ini := time.Now()
		if err := cmd.Run(); err != nil {
			fmt.Fprintln(os.Stderr, "falló:", err)
			os.Exit(1)
		}
		tiempos = append(tiempos, time.Since(ini))
	}
	slices.Sort(tiempos)
	med := tiempos[len(tiempos)/2]
	p90 := tiempos[len(tiempos)*9/10]
	fmt.Printf("mediana %.2f ms   p90 %.2f ms   (%d arranques)\n",
		float64(med.Microseconds())/1000, float64(p90.Microseconds())/1000, *n)
}
