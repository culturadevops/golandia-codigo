// gauge-global: la versión «de tutorial» del mismo exporter, para comparar. Un GaugeVec global que
// una goroutine actualiza cada -cada. Es el patrón para instrumentar TU app, no para un exporter.
package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"strings"
	"time"

	"certexp/sonda"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var vence = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: "certexp_not_after_timestamp_seconds",
	Help: "Fin de validez del certificado (segundos Unix).",
}, []string{"objetivo"})

func main() {
	escuchar := flag.String("escuchar", ":9679", "dirección del /metrics")
	objetivos := flag.String("objetivos", "", "lista host:puerto separada por comas")
	cada := flag.Duration("cada", 60*time.Second, "cada cuánto se actualiza")
	flag.Parse()

	prometheus.MustRegister(vence)
	go func() {
		for {
			for _, obj := range strings.Split(*objetivos, ",") {
				r := sonda.Sondear(context.Background(), obj, 2*time.Second)
				if r.Err != nil {
					log.Printf("%s: %v", obj, r.Err) // y el valor viejo se queda ahí
					continue
				}
				vence.WithLabelValues(obj).Set(float64(r.Vence.Unix()))
			}
			time.Sleep(*cada)
		}
	}()
	http.Handle("/metrics", promhttp.Handler())
	log.Printf("gauge-global en %s · cada %s", *escuchar, *cada)
	log.Fatal(http.ListenAndServe(*escuchar, nil))
}
