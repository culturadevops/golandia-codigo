// Lección 6 · Trazas: el contexto que se pierde, dentro del programa y entre dos servicios.
package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"

	"curso-obs/comun"
)

var (
	exp    = tracetest.NewInMemoryExporter()
	tp     = sdktrace.NewTracerProvider(sdktrace.WithSyncer(exp))
	tracer = tp.Tracer("tienda")
)

func cobrar(ctx context.Context) {
	_, s := tracer.Start(ctx, "cobrar")
	s.End()
}

// pintar enseña los spans exportados: su traza (8 primeros caracteres) y de quién cuelgan.
func pintar() {
	trazas := map[trace.TraceID]bool{}
	for _, s := range exp.GetSpans() {
		trazas[s.SpanContext.TraceID()] = true
		padre := "(raíz)"
		if s.Parent.IsValid() {
			padre = "hijo de " + s.Parent.SpanID().String()[:8]
		}
		fmt.Printf("  traza %s · span %s · %-28s %s\n", s.SpanContext.TraceID().String()[:8],
			s.SpanContext.SpanID().String()[:8], s.Name, padre)
	}
	fmt.Printf("  -> trazas distintas: %d\n", len(trazas))
	exp.Reset()
}

func main() {
	comun.Titulo("6.1 · Dentro del programa: el ctx que se pasa")
	ctx, raiz := tracer.Start(context.Background(), "POST /pagar")
	cobrar(ctx)
	raiz.End()
	pintar()

	comun.Titulo("6.2 · La trampa: un context.Background() a mitad de camino")
	_, raiz = tracer.Start(context.Background(), "POST /pagar")
	cobrar(context.Background()) // la función no recibía ctx y alguien puso Background()
	raiz.End()
	pintar()
	fmt.Println("  en el visor salen dos peticiones distintas: el cobro ya no cuelga del pago")

	// El otro servicio: inventario, instrumentado con otelhttp.
	var cabecera string
	inventario := httptest.NewServer(otelhttp.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cabecera = r.Header.Get("traceparent")
		w.Write([]byte("ok"))
	}), "GET /stock", otelhttp.WithTracerProvider(tp)))
	defer inventario.Close()

	llamar := func(cli *http.Client) {
		cabecera = ""
		ctx, raiz := tracer.Start(context.Background(), "POST /pagar")
		req, _ := http.NewRequestWithContext(ctx, "GET", inventario.URL, nil)
		resp, err := cli.Do(req)
		if err == nil {
			resp.Body.Close()
		}
		raiz.End()
		if cabecera == "" {
			cabecera = "(ninguna)"
		}
		fmt.Println("  traceparent que llega al inventario: " + cabecera)
		pintar()
	}
	conOtel := &http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport, otelhttp.WithTracerProvider(tp))}

	comun.Titulo("6.3 · Entre servicios: con http.DefaultClient")
	llamar(http.DefaultClient)

	comun.Titulo("6.4 · La trampa: con otelhttp.NewTransport, pero sin propagador")
	llamar(conOtel)
	fmt.Println("  el propagador global por defecto no hace nada: otelhttp no tiene con qué escribir la cabecera")

	comun.Titulo("6.5 · Lo correcto: otel.SetTextMapPropagator(propagation.TraceContext{})")
	otel.SetTextMapPropagator(propagation.TraceContext{})
	llamar(conOtel)
	_ = strings.TrimSpace
}
