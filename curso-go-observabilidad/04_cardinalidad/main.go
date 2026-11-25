// Lección 4 · Cardinalidad: un identificador en una etiqueta es una serie nueva por cada valor.
//
// Corre junto a un Prometheus de verdad que la raspa cada segundo (ejecutar_todo.py).
package main

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"

	"curso-obs/comun"
)

func main() {
	reg := prometheus.NewRegistry()
	mal := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "tienda_peticiones_ruta_cruda_total", Help: "Peticiones, con la ruta tal cual llega.",
	}, []string{"ruta"})
	bien := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "tienda_peticiones_total", Help: "Peticiones, con la plantilla de la ruta.",
	}, []string{"ruta"})
	reg.MustRegister(mal, bien)
	comun.Servir(reg)

	for id := 1; id <= 1000; id++ {
		ruta := fmt.Sprintf("/pedidos/%d", id)
		mal.WithLabelValues(ruta).Inc()             // la trampa: la ruta con el id dentro
		bien.WithLabelValues("/pedidos/{id}").Inc() // lo correcto: la plantilla, como la registró el router
	}

	comun.Titulo("4.1 · 1.000 peticiones a /pedidos/{id}, con mil ids distintos")
	fmt.Println("  con la ruta tal cual: " + strings.Join(comun.Muestra(reg, "tienda_peticiones_ruta_cruda_total")[:2], " · ") + " · ...")
	fmt.Println("  con la plantilla:     " + comun.Muestra(reg, "tienda_peticiones_total")[0])

	comun.EsperarRaspado()
	comun.Titulo("4.2 · Las series que Prometheus tiene que guardar")
	fmt.Println("  count(tienda_peticiones_ruta_cruda_total) -> " + comun.Consulta("count(tienda_peticiones_ruta_cruda_total)")[0])
	fmt.Println("  count(tienda_peticiones_total)            -> " + comun.Consulta("count(tienda_peticiones_total)")[0])
	fmt.Println("  ningún error: cada pedido nuevo es una serie más, para siempre")

	comun.Titulo("4.3 · Y lo que pesa /metrics en cada raspado")
	familias, _ := reg.Gather()
	for _, f := range familias {
		var b bytes.Buffer
		expfmt.MetricFamilyToText(&b, f)
		fmt.Printf("  %-36s %6d bytes\n", f.GetName(), b.Len())
	}
}
