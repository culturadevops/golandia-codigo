// worker: se conecta al servidor de Temporal (TEMPORAL_ADDRESS) y ejecuta workflows y actividades de la cola «pedidos».
package main

import (
	"fmt"
	"os"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"temporaldemo/pedido"
)

func main() {
	if v := os.Getenv("VERSION"); v != "" {
		pedido.Version = v
	}
	c, err := client.Dial(client.Options{HostPort: os.Getenv("TEMPORAL_ADDRESS"), Identity: os.Getenv("WORKER")})
	if err != nil {
		panic(err)
	}
	defer c.Close()
	w := worker.New(c, pedido.Cola, worker.Options{})
	w.RegisterWorkflow(pedido.Pedido)
	w.RegisterActivity(pedido.Reservar)
	w.RegisterActivity(pedido.Validar)
	w.RegisterActivity(pedido.Cobrar)
	w.RegisterActivity(pedido.Facturar)
	w.RegisterActivity(pedido.Enviar)
	fmt.Printf("[%s] worker listo, versión %s del workflow\n", os.Getenv("WORKER"), pedido.Version)
	if err := w.Run(worker.InterruptCh()); err != nil {
		panic(err)
	}
}
