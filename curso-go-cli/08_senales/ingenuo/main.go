// Lección 8, versión ingenua: escribe el informe sobre la marcha y no piensa en Ctrl+C.
//
//	ingenuo08 -informe informe.txt URL...            sin escuchar señales
//	ingenuo08 -escucha -informe informe.txt URL...   escucha, pero las peticiones no llevan el ctx
package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	informe := flag.String("informe", "informe.txt", "archivo del informe")
	escucha := flag.Bool("escucha", false, "atiende Ctrl+C con signal.NotifyContext")
	flag.Parse()

	ctx := context.Background()
	if *escucha {
		var stop context.CancelFunc
		ctx, stop = signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
		defer stop()
	}

	f, err := os.Create(*informe) // ⛔ trunca el informe de ayer ANTES de tener el de hoy
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	defer f.Close()

	for _, u := range flag.Args() {
		if ctx.Err() != nil {
			fmt.Fprintln(os.Stderr, "interrumpido")
			os.Exit(130)
		}
		ini := time.Now()
		resp, err := http.Get(u) // ⛔ sin ctx: Ctrl+C no corta esta petición
		if err != nil {
			fmt.Fprintf(f, "CAIDO   %s  (%v)\n", u, err)
			continue
		}
		resp.Body.Close()
		fmt.Fprintf(f, "%d  %s  %s\n", resp.StatusCode, time.Since(ini).Round(time.Millisecond), u)
	}
	if ctx.Err() != nil {
		fmt.Fprintln(os.Stderr, "interrumpido")
		os.Exit(130)
	}
}
