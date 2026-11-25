// iniciar: arranca el workflow Pedido con el id dado y sale (no espera el resultado).
package main

import (
	"context"
	"fmt"
	"os"

	"go.temporal.io/sdk/client"

	"temporaldemo/pedido"
)

func main() {
	id := os.Args[1]
	c, err := client.Dial(client.Options{HostPort: os.Getenv("TEMPORAL_ADDRESS")})
	if err != nil {
		panic(err)
	}
	defer c.Close()
	run, err := c.ExecuteWorkflow(context.Background(), client.StartWorkflowOptions{ID: id, TaskQueue: pedido.Cola},
		pedido.Pedido, id)
	if err != nil {
		panic(err)
	}
	fmt.Printf("workflow %s iniciado (run %s)\n", run.GetID(), run.GetRunID())
}
