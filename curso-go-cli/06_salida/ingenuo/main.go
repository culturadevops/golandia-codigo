// Lección 6, versión ingenua: os.Exit y log.Fatal donde «toca».
package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"curso-cli/comun"
)

func main() {
	timeout := flag.String("timeout", "10s", "tiempo máximo por endpoint")
	flag.Parse()
	d, err := time.ParseDuration(*timeout)
	if err != nil {
		log.Fatal(err) // ⛔ código 1: el mismo que «hay un endpoint caído»
	}

	w := bufio.NewWriter(os.Stdout)
	defer w.Flush() // ⛔ nunca se ejecuta si sales con os.Exit

	caidos := 0
	for _, u := range flag.Args() {
		r := comun.Revisar(context.Background(), u, d)
		if !r.OK() {
			caidos++
		}
		fmt.Fprintln(w, comun.Linea(r))
	}
	if caidos > 0 {
		os.Exit(1) // ⛔ se salta el defer: lo que había en el búfer se pierde
	}
}
