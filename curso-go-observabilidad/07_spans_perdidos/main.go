// Lección 7 · Los spans que nunca llegan: el proceso que sale sin Shutdown y el span sin End.
//
// Cada caso corre en un proceso hijo, porque lo que se mide es lo que pasa AL SALIR.
package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"curso-obs/comun"
)

// aFichero es un exportador que apunta en un fichero cada span que le llega.
type aFichero struct{ f *os.File }

func (e aFichero) ExportSpans(_ context.Context, spans []sdktrace.ReadOnlySpan) error {
	for _, s := range spans {
		fmt.Fprintln(e.f, s.Name())
	}
	return nil
}

func (e aFichero) Shutdown(context.Context) error { return e.f.Close() }

func hijo(modo, fichero string) {
	f, _ := os.Create(fichero)
	// El procesador por lotes de siempre, con sus valores por defecto (espera hasta 5 s para enviar).
	tp := sdktrace.NewTracerProvider(sdktrace.WithBatcher(aFichero{f}))
	tracer := tp.Tracer("tienda")
	for i := 0; i < 100; i++ {
		_, s := tracer.Start(context.Background(), fmt.Sprintf("pedido-%03d", i))
		if modo == "sin_end" && i%10 == 0 {
			continue // un return temprano antes del End: 1 de cada 10
		}
		s.End()
	}
	if modo != "sin_shutdown" {
		t0 := time.Now()
		tp.Shutdown(context.Background())
		fmt.Printf("Shutdown tardó %s", comun.Ms(time.Since(t0)))
	}
}

func main() {
	if len(os.Args) == 3 {
		hijo(os.Args[1], os.Args[2])
		return
	}
	caso := func(titulo, modo string) {
		comun.Titulo(titulo)
		fichero := "/tmp/spans_" + modo + ".txt"
		out, _ := exec.Command(os.Args[0], modo, fichero).CombinedOutput()
		b, _ := os.ReadFile(fichero)
		n := len(strings.Fields(string(b)))
		fmt.Printf("  spans creados: 100 · spans que llegaron al exportador: %d", n)
		if s := strings.TrimSpace(string(out)); s != "" {
			fmt.Printf(" · %s", s)
		}
		fmt.Println()
	}
	caso("7.1 · La trampa: el programa termina sin tp.Shutdown()", "sin_shutdown")
	fmt.Println("  estaban en la cola del lote, esperando su turno, y el proceso se fue con ellos")
	caso("7.2 · Lo correcto: defer tp.Shutdown(ctx)", "con_shutdown")
	caso("7.3 · La otra trampa: un span sin End() (1 de cada 10)", "sin_end")
	fmt.Println("  un span sin End no se exporta nunca: ni con Shutdown. Por eso siempre defer span.End()")
}
