// Lección 3: la misma CLI de la lección 2, con Cobra. Para comparar, no para elegir a ciegas.
//
//	vigia03 revisar --timeout 2s URL...
//	vigia03 version
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"curso-cli/comun"
)

func main() {
	raiz := &cobra.Command{
		Use:   "vigia03",
		Short: "Revisa endpoints HTTP",
	}

	var timeout time.Duration
	revisar := &cobra.Command{
		Use:   "revisar URL...",
		Short: "revisa uno o varios endpoints HTTP",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			for _, u := range args {
				fmt.Println(comun.Linea(comun.Revisar(context.Background(), u, timeout)))
			}
		},
	}
	revisar.Flags().DurationVarP(&timeout, "timeout", "t", 10*time.Second, "tiempo máximo por endpoint")

	version := &cobra.Command{
		Use:   "version",
		Short: "muestra la versión",
		Run:   func(cmd *cobra.Command, args []string) { fmt.Println("vigia03 dev") },
	}

	raiz.AddCommand(revisar, version)
	if err := raiz.Execute(); err != nil {
		os.Exit(2)
	}
}
