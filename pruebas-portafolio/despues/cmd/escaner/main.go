// El binario: solo lee las opciones y llama al paquete. Toda la lógica está en escaner/, donde se prueba.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"time"

	"escaner/escaner"
)

func main() {
	host := flag.String("host", "127.0.0.1", "máquina a escanear (tuya o con permiso)")
	desde := flag.Int("desde", 1, "primer puerto")
	hasta := flag.Int("hasta", 1024, "último puerto")
	conc := flag.Int("conc", 500, "conexiones a la vez")
	espera := flag.Duration("timeout", 500*time.Millisecond, "espera por puerto")
	flag.Parse()

	ctx, parar := signal.NotifyContext(context.Background(), os.Interrupt)
	defer parar()
	res, err := escaner.Escanear(ctx, *host, *desde, *hasta, *conc, *espera)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	fmt.Printf("%d abiertos en %v: %v\nerrores: %v\n", len(res.Abiertos), res.Duracion.Round(time.Millisecond), res.Abiertos, res.Errores)
}
