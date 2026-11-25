// vigia revisa endpoints HTTP. Es la CLI que se construye a lo largo del curso.
//
//	vigia revisar [-timeout 3s] [-concurrencia 8] [-json] [-solo-caidos] [-informe f] URL... | -
//	vigia config      la configuración efectiva y de dónde sale cada valor
//	vigia version     versión, commit y con qué se compiló
//
// Códigos de salida: 0 todo arriba · 1 hay endpoints caídos · 2 uso o configuración mal ·
// 130 interrumpido (Ctrl+C o SIGTERM).
package main

import (
	"context"
	"io"
	"os"
	"os/signal"
	"syscall"
)

// Códigos de salida. Son parte de la interfaz de vigia, igual que sus flags.
const (
	exitOK           = 0
	exitCaidos       = 1
	exitUso          = 2
	exitInterrumpido = 130
)

// Entorno es todo lo que vigia toma del mundo. En los tests se rellena a mano.
type Entorno struct {
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
	Getenv func(string) string
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	go func() {
		<-ctx.Done()
		stop() // a partir de aquí, un segundo Ctrl+C mata el proceso sin esperar
	}()
	codigo := run(ctx, os.Args[1:], Entorno{os.Stdin, os.Stdout, os.Stderr, os.Getenv})
	stop()
	os.Exit(codigo) // el único os.Exit del programa: los defer de run ya se ejecutaron
}
