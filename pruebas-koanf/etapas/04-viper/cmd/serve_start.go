package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var serveStartCmd = &cobra.Command{
	Use:   "start",
	Short: "arranca el servidor (demostracion, no abre ningun puerto de verdad)",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("host=%s puerto=%d\n", viper.GetString("server.host"), viper.GetInt("server.port"))
		return nil
	},
}

func init() {
	serveStartCmd.Flags().String("host", "127.0.0.1", "host donde escucha")
	serveStartCmd.Flags().Int("port", 8080, "puerto donde escucha")

	// BindPFlag no ata el valor: ata la PRIORIDAD. Si el usuario no toco el flag
	// en esta llamada, viper sigue mirando env y config antes de caer al default del flag.
	viper.BindPFlag("server.host", serveStartCmd.Flags().Lookup("host"))
	viper.BindPFlag("server.port", serveStartCmd.Flags().Lookup("port"))
}
