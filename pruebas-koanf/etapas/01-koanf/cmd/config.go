package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "inspecciona la configuracion resuelta",
}

var configViewCmd = &cobra.Command{
	Use:   "view",
	Short: "imprime toda la configuracion ya resuelta (archivo + env + flags)",
	RunE: func(cmd *cobra.Command, args []string) error {
		// El equivalente a viper.AllSettings(). koanf devuelve el mapa aplanado,
		// con las claves ya en forma "server.host", que se lee mejor para depurar.
		for clave, v := range k.All() {
			fmt.Printf("%s = %v\n", clave, v)
		}
		return nil
	},
}

func init() {
	configCmd.AddCommand(configViewCmd)
}
