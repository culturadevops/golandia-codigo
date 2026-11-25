// Lección 1: vigia con el paquete flag de la biblioteca estándar.
//
//	vigia01 -timeout 2s -v https://ejemplo.com/salud https://ejemplo.com/api
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"curso-cli/comun"
)

func main() {
	timeout := flag.Duration("timeout", 10*time.Second, "tiempo máximo por endpoint")
	verbose := flag.Bool("v", false, "cuenta lo que entendió de los argumentos")
	version := flag.Bool("version", false, "muestra la versión y sale")
	flag.Parse()

	if *version {
		fmt.Println("vigia01 dev")
		return
	}
	if *verbose {
		fmt.Fprintf(os.Stderr, "# entendí: timeout=%s v=%t urls=%q\n", *timeout, *verbose, flag.Args())
	}
	for _, u := range flag.Args() {
		fmt.Println(comun.Linea(comun.Revisar(context.Background(), u, *timeout)))
	}
}
