// Lección 8 · Todo junto: un servicio con logs, métricas y trazas que se encuentran entre sí.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"

	"curso-obs/comun"
)

// conTraza es un slog.Handler que añade trace_id y span_id a cada registro que lleva contexto.
type conTraza struct{ slog.Handler }

func (h conTraza) Handle(ctx context.Context, r slog.Record) error {
	if sc := trace.SpanContextFromContext(ctx); sc.IsValid() {
		r.AddAttrs(slog.String("trace_id", sc.TraceID().String()), slog.String("span_id", sc.SpanID().String()))
	}
	return h.Handler.Handle(ctx, r)
}

func (h conTraza) WithAttrs(a []slog.Attr) slog.Handler { return conTraza{h.Handler.WithAttrs(a)} }
func (h conTraza) WithGroup(n string) slog.Handler      { return conTraza{h.Handler.WithGroup(n)} }

type respuesta struct {
	http.ResponseWriter
	codigo int
}

func (r *respuesta) WriteHeader(c int) { r.codigo = c; r.ResponseWriter.WriteHeader(c) }

func main() {
	exp := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exp))
	tracer := tp.Tracer("tienda")

	var registros bytes.Buffer
	log := slog.New(conTraza{slog.NewJSONHandler(&registros, &slog.HandlerOptions{
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				return slog.Attr{}
			}
			return a
		},
	})})

	reg := prometheus.NewRegistry()
	duracion := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "tienda_http_duracion_seconds", Help: "Duración de las peticiones.",
		Buckets: prometheus.ExponentialBuckets(0.0005, 2, 16),
	}, []string{"ruta", "codigo"})
	reg.MustRegister(duracion)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /pedidos/{id}", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		_, s := tracer.Start(ctx, "consultar inventario")
		time.Sleep(2 * time.Millisecond)
		s.End()
		log.InfoContext(ctx, "pedido consultado", "pedido", r.PathValue("id"))
		w.Write([]byte("ok"))
	})
	// El middleware de métricas: la etiqueta es r.Pattern (la plantilla), nunca r.URL.Path.
	metricas := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t0 := time.Now()
		rw := &respuesta{ResponseWriter: w, codigo: 200}
		mux.ServeHTTP(rw, r)
		duracion.WithLabelValues(r.Pattern, fmt.Sprint(rw.codigo)).Observe(time.Since(t0).Seconds())
	})
	srv := httptest.NewServer(otelhttp.NewHandler(metricas, "tienda", otelhttp.WithTracerProvider(tp)))
	defer srv.Close()

	comun.Titulo("8.1 · Una petición: GET /pedidos/42")
	resp, _ := http.Get(srv.URL + "/pedidos/42")
	resp.Body.Close()
	linea := strings.TrimSpace(registros.String())
	fmt.Println("  el registro:")
	fmt.Println("    " + linea)
	fmt.Println("  la traza:")
	for _, s := range exp.GetSpans() {
		fmt.Printf("    %s · %s\n", s.SpanContext.TraceID().String(), s.Name)
	}
	var campos map[string]any
	json.Unmarshal([]byte(linea), &campos)
	fmt.Printf("  ¿el trace_id del registro es el de la traza?: %v\n",
		campos["trace_id"] == exp.GetSpans()[0].SpanContext.TraceID().String())

	comun.Titulo("8.2 · 50 pedidos distintos más: las series de la métrica")
	for id := 1; id <= 50; id++ {
		r, _ := http.Get(fmt.Sprintf("%s/pedidos/%d", srv.URL, id))
		r.Body.Close()
	}
	r, _ := http.Get(srv.URL + "/no-existe")
	r.Body.Close()
	familias, _ := reg.Gather()
	for _, f := range familias {
		for _, m := range f.GetMetric() {
			et := []string{}
			for _, e := range m.GetLabel() {
				et = append(et, fmt.Sprintf("%s=%q", e.GetName(), e.GetValue()))
			}
			n := m.GetHistogram().GetSampleCount()
			pal := "peticiones"
			if n == 1 {
				pal = "petición"
			}
			fmt.Printf("  %s{%s} · %d %s\n", f.GetName(), strings.Join(et, ","), n, pal)
		}
	}
	fmt.Println("  51 pedidos, una serie: la etiqueta es la plantilla que ya conoce el router (r.Pattern, Go 1.23)")
}
