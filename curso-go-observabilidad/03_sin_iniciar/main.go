// Lección 3 · La métrica que no existe hasta el primer error, y los primeros errores que no cuentan.
//
// Corre junto a un Prometheus de verdad que la raspa cada segundo (ejecutar_todo.py).
package main

import (
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"curso-obs/comun"
)

func main() {
	reg := prometheus.NewRegistry()
	mal := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "tienda_pagos_total", Help: "Pagos, por resultado.",
	}, []string{"resultado"})
	bien := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "tienda_pagos_iniciado_total", Help: "Pagos, por resultado (con las series creadas al arrancar).",
	}, []string{"resultado"})
	reg.MustRegister(mal, bien)
	// Lo correcto: crear las series al arrancar, aunque valgan 0.
	bien.WithLabelValues("ok")
	bien.WithLabelValues("error")
	comun.Servir(reg)

	for i := 0; i < 120; i++ {
		mal.WithLabelValues("ok").Inc()
		bien.WithLabelValues("ok").Inc()
	}

	comun.Titulo("3.1 · 120 pagos bien, ninguno mal: lo que publica /metrics")
	for _, l := range comun.Muestra(reg, "tienda_pagos_total") {
		fmt.Println("  " + l)
	}
	fmt.Println("  la serie resultado=\"error\" no existe: CounterVec la crea en el primer WithLabelValues")

	comun.EsperarRaspado()
	comun.Titulo("3.2 · Lo que ve la alerta, en el Prometheus de verdad")
	q := `sum(rate(tienda_pagos_total{resultado="error"}[1m])) / sum(rate(tienda_pagos_total[1m]))`
	fmt.Println("  tasa de errores = " + q)
	for _, l := range comun.Consulta(q) {
		fmt.Println("    -> " + l)
	}
	fmt.Println("  no es 0: es «sin datos». El panel dice «No data» y una alerta «> 0.05» no tiene con qué compararse")

	time.Sleep(2 * time.Second)
	comun.Titulo("3.3 · La trampa: llegan los 5 primeros errores")
	for i := 0; i < 5; i++ {
		mal.WithLabelValues("error").Inc()
		bien.WithLabelValues("error").Inc()
	}
	comun.EsperarRaspado()
	time.Sleep(2 * time.Second)
	comun.EsperarRaspado()
	for _, l := range comun.Muestra(reg, "tienda_pagos_total") {
		fmt.Println("  " + l)
	}
	for _, l := range comun.Consulta(`increase(tienda_pagos_total{resultado="error"}[1m])`) {
		fmt.Println("  increase(errores[1m]) -> " + l)
	}
	fmt.Println("  la serie nace valiendo 5 y para Prometheus ese es su punto de partida: esos 5 no se ven")

	comun.Titulo("3.4 · Lo correcto: las series creadas al arrancar, a 0")
	for _, l := range comun.Consulta(`increase(tienda_pagos_iniciado_total{resultado="error"}[1m])`) {
		fmt.Println("  increase(errores[1m]) -> " + l)
	}
	fmt.Println("  increase extrapola a los bordes de la ventana: sale cerca de 5, no exacto")
}
