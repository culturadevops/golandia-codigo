// Lección 2: subcomandos con la biblioteca estándar. Un flag.FlagSet por subcomando.
//
//	vigia02 revisar -timeout 2s URL...
//	vigia02 version
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"curso-cli/comun"
)

const ayuda = `uso: vigia02 <subcomando> [flags]

subcomandos:
  revisar   revisa uno o varios endpoints HTTP
  version   muestra la versión

"vigia02 <subcomando> -h" da la ayuda de cada uno.
`

func main() {
	if len(os.Args) < 2 {
		fmt.Fprint(os.Stderr, ayuda)
		os.Exit(2)
	}
	switch os.Args[1] {
	case "revisar":
		revisar(os.Args[2:])
	case "version":
		fmt.Println("vigia02 dev")
	case "-h", "-help", "--help", "ayuda":
		fmt.Print(ayuda)
	default:
		fmt.Fprintf(os.Stderr, "vigia02: subcomando desconocido %q\n\n%s", os.Args[1], ayuda)
		os.Exit(2)
	}
}

func revisar(args []string) {
	fs := flag.NewFlagSet("revisar", flag.ExitOnError)
	timeout := fs.Duration("timeout", 10*time.Second, "tiempo máximo por endpoint")
	fs.Parse(args) // con ExitOnError, un flag mal sale con código 2
	if fs.NArg() == 0 {
		fmt.Fprintln(os.Stderr, "vigia02 revisar: falta al menos una URL")
		os.Exit(2)
	}
	for _, u := range fs.Args() {
		fmt.Println(comun.Linea(comun.Revisar(context.Background(), u, *timeout)))
	}
}
