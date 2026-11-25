// Comando api: el servicio de tareas.
//
//	api           sirve HTTP (PUERTO, BD_URL, ...: ver internal/config)
//	api migrar    aplica las migraciones con BD_URL_MIGRAR (el usuario dueño del esquema)
//	api salud     pide /salud/listo a 127.0.0.1:PUERTO; para el HEALTHCHECK de Docker
package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	err := run(ctx, os.Args[1:], os.Getenv, os.Stdout)
	stop()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1) // os.Exit no corre los defer: por eso aquí, y no dentro de run
	}
}

// run es main sin os.Exit ni globales: los defer se cumplen y se puede probar.
func run(ctx context.Context, args []string, getenv func(string) string, out io.Writer) error {
	orden := "servir"
	if len(args) > 0 {
		orden = args[0]
	}
	switch orden {
	case "servir":
		return servir(ctx, getenv, out)
	case "migrar":
		return migrar(ctx, getenv, out)
	case "salud":
		return salud(ctx, getenv)
	}
	return fmt.Errorf("orden desconocida %q (servir, migrar, salud)", orden)
}
