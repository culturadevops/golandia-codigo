// certexp: exporter de Prometheus que, en cada scrape, sondea una lista de hosts TLS y expone
// cuándo vence su certificado. Patrón Collector: no guarda estado entre scrapes.
package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"certexp/sonda"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	descVence = prometheus.NewDesc("certexp_not_after_timestamp_seconds",
		"Fin de validez del certificado (segundos Unix).", []string{"objetivo"}, nil)
	descOK = prometheus.NewDesc("certexp_probe_success",
		"1 si la sonda TLS terminó bien, 0 si no.", []string{"objetivo"}, nil)
	descDur = prometheus.NewDesc("certexp_probe_duration_seconds",
		"Lo que tardó la sonda.", []string{"objetivo"}, nil)
	// ⚠️ Solo para la demo de cardinalidad: el texto del error como etiqueta.
	descErr = prometheus.NewDesc("certexp_probe_error",
		"Error de la sonda (MALA IDEA: el texto va en una etiqueta).", []string{"objetivo", "error"}, nil)
)

type certCollector struct {
	objetivos  []string
	timeout    time.Duration
	secuencial bool
	conError   bool
}

func (c *certCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- descVence
	ch <- descOK
	ch <- descDur
	if c.conError {
		ch <- descErr
	}
}

// Collect se llama en CADA scrape: aquí se hace el trabajo.
func (c *certCollector) Collect(ch chan<- prometheus.Metric) {
	var wg sync.WaitGroup
	for _, obj := range c.objetivos {
		sondear := func() {
			r := sonda.Sondear(context.Background(), obj, c.timeout)
			ch <- prometheus.MustNewConstMetric(descDur, prometheus.GaugeValue, r.Duracion.Seconds(), obj)
			if r.Err != nil {
				ch <- prometheus.MustNewConstMetric(descOK, prometheus.GaugeValue, 0, obj)
				if c.conError {
					ch <- prometheus.MustNewConstMetric(descErr, prometheus.GaugeValue, 1, obj, r.Err.Error())
				}
				return
			}
			ch <- prometheus.MustNewConstMetric(descOK, prometheus.GaugeValue, 1, obj)
			ch <- prometheus.MustNewConstMetric(descVence, prometheus.GaugeValue, float64(r.Vence.Unix()), obj)
		}
		if c.secuencial {
			sondear()
		} else {
			wg.Go(sondear)
		}
	}
	wg.Wait()
}

func main() {
	escuchar := flag.String("escuchar", ":9679", "dirección del /metrics")
	objetivos := flag.String("objetivos", "", "lista host:puerto separada por comas")
	timeout := flag.Duration("timeout", 2*time.Second, "límite por host (0 = sin límite)")
	secuencial := flag.Bool("secuencial", false, "sondear uno detrás de otro (para comparar)")
	conError := flag.Bool("etiqueta-error", false, "poner el texto del error en una etiqueta (para la demo de cardinalidad)")
	flag.Parse()

	c := &certCollector{objetivos: strings.Split(*objetivos, ","), timeout: *timeout, secuencial: *secuencial, conError: *conError}
	reg := prometheus.NewRegistry()
	reg.MustRegister(c)
	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	log.Printf("certexp en %s · %d objetivos · timeout %s · secuencial=%v", *escuchar, len(c.objetivos), *timeout, *secuencial)
	log.Fatal(http.ListenAndServe(*escuchar, nil))
}
