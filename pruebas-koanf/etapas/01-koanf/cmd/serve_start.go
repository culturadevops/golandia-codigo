package cmd

import (
	"fmt"

	"github.com/knadh/koanf/providers/posflag"
	"github.com/spf13/cobra"
)

var serveStartCmd = &cobra.Command{
	Use:   "start",
	Short: "arranca el servidor (demostracion)",
	RunE: func(cmd *cobra.Command, args []string) error {
		// posflag se carga AQUI, al ejecutar, porque necesita las flags ya parseadas.
		// Al cargarse el ultimo, gana sobre entorno y archivo: ese es el orden.
		// ⚠️ Y a diferencia de viper.BindPFlag, este error SI se comprueba
		// (el linter del video 244 encontro justo ese fallo en la version con viper).
		if err := k.Load(posflag.Provider(cmd.Flags(), ".", k), nil); err != nil {
			return fmt.Errorf("cargando flags: %w", err)
		}
		fmt.Printf("host=%s puerto=%d\n", k.String("server.host"), k.Int("server.port"))
		return nil
	},
}

func init() {
	serveStartCmd.Flags().String("server.host", "127.0.0.1", "host donde escucha")
	serveStartCmd.Flags().Int("server.port", 8080, "puerto donde escucha")
}
