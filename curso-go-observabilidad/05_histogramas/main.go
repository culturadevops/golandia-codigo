// Lección 5 · Histogramas: con los buckets por defecto, el p99 que da Prometheus puede ser mentira.
//
// Corre junto a un Prometheus de verdad que la raspa cada segundo (ejecutar_todo.py).
// Las latencias se OBSERVAN directamente (h.Observe), no se esperan: lo que se mide aquí es lo que
// Prometheus calcula a partir de los buckets, no un servidor lento.
package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/prometheus/client_golang/prometheus"

	"curso-obs/comun"
)

func percentil(v []float64, p float64) float64 {
	s := append([]float64(nil), v...)
	sort.Float64s(s)
	return s[int(p*float64(len(s)))-1]
}

func coma(f float64) string {
	return strings.Replace(fmt.Sprintf("%g", f), ".", ",", 1)
}

func main() {
	reg := prometheus.NewRegistry()
	nuevo := func(nombre string, buckets []float64) prometheus.Histogram {
		h := prometheus.NewHistogram(prometheus.HistogramOpts{Name: nombre, Help: "Latencia.", Buckets: buckets})
		reg.MustRegister(h)
		return h
	}
	ajustados := prometheus.ExponentialBuckets(0.0005, 2, 16) // de 0,5 ms a 16,4 s

	// El pago: 900 peticiones de 200 ms y 100 de 15 s (un proveedor que se cuelga).
	var pago []float64
	for i := 0; i < 900; i++ {
		pago = append(pago, 0.2)
	}
	for i := 0; i < 100; i++ {
		pago = append(pago, 15)
	}
	// La caché: 1.000 peticiones de 0,8 ms.
	var cache []float64
	for i := 0; i < 1000; i++ {
		cache = append(cache, 0.0008)
	}
	for _, par := range []struct {
		h  prometheus.Histogram
		vs []float64
	}{
		{nuevo("pago_defecto_seconds", prometheus.DefBuckets), pago},
		{nuevo("pago_ajustado_seconds", ajustados), pago},
		{nuevo("cache_defecto_seconds", prometheus.DefBuckets), cache},
		{nuevo("cache_ajustado_seconds", ajustados), cache},
	} {
		for _, v := range par.vs {
			par.h.Observe(v)
		}
	}
	comun.Servir(reg)

	comun.Titulo("5.1 · Los buckets por defecto (prometheus.DefBuckets), en segundos")
	partes := []string{}
	for _, b := range prometheus.DefBuckets {
		partes = append(partes, coma(b))
	}
	fmt.Println("  " + strings.Join(partes, " · ") + " · +Inf")

	comun.EsperarRaspado()
	q := func(p, nombre string) string {
		return comun.Consulta(fmt.Sprintf("histogram_quantile(%s, %s_bucket)", p, nombre))[0]
	}
	comun.Titulo("5.2 · La trampa por arriba: 900 pagos de 200 ms y 100 de 15 s")
	fmt.Printf("  p99 de verdad: %s s · histogram_quantile(0.99) con DefBuckets: %s s\n", coma(percentil(pago, 0.99)), q("0.99", "pago_defecto_seconds"))
	fmt.Println("  todo lo que pasa de 10 s cae en +Inf, y Prometheus responde el último borde: 10")

	comun.Titulo("5.3 · La trampa por abajo: 1.000 respuestas de caché de 0,8 ms")
	fmt.Printf("  p50 de verdad: %s s · histogram_quantile(0.5) con DefBuckets: %s s\n", coma(percentil(cache, 0.5)), q("0.5", "cache_defecto_seconds"))
	fmt.Println("  todo cae en el primer bucket (0 a 5 ms) y Prometheus interpola por la mitad")

	comun.Titulo("5.4 · Lo correcto: buckets para TU rango (ExponentialBuckets(0.0005, 2, 16))")
	fmt.Printf("  pago, p99:  de verdad %s s · Prometheus %s s\n", coma(percentil(pago, 0.99)), q("0.99", "pago_ajustado_seconds"))
	fmt.Printf("  caché, p50: de verdad %s s · Prometheus %s s\n", coma(percentil(cache, 0.5)), q("0.5", "cache_ajustado_seconds"))
	fmt.Println("  un histograma no guarda valores, guarda cuántos cayeron en cada bucket: el resultado es una estimación")
}
