// vigia-sdk: lo mismo que el sondeo de vigia (listar enfermos con la etiqueta y reiniciarlos), pero con el SDK oficial.
// Solo existe para comparar tamaño de binario y dependencias.
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/moby/moby/client"
)

func main() {
	ctx := context.Background()
	cli, err := client.New(client.FromEnv)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer cli.Close()
	res, err := cli.ContainerList(ctx, client.ContainerListOptions{
		Filters: make(client.Filters).Add("label", os.Args[1]).Add("health", "unhealthy"),
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	for _, c := range res.Items {
		fmt.Println("reinicio", c.Names[0])
		if _, err := cli.ContainerRestart(ctx, c.ID, client.ContainerRestartOptions{}); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}
}
