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

	// A proposito, un typo real: "host" en vez de "host". Lookup con un nombre que
	// no existe devuelve nil, y eso es justo lo que el error ignorado se comia en silencio.
	if err := viper.BindPFlag("server.host", serveStartCmd.Flags().Lookup("hosT")); err != nil {
		fmt.Println("BindPFlag fallo de verdad:", err)
	}
	viper.BindPFlag("server.port", serveStartCmd.Flags().Lookup("port"))
}
