package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "inspecciona la configuracion resuelta",
}

var configViewCmd = &cobra.Command{
	Use:   "view",
	Short: "imprime toda la configuracion ya resuelta (archivo + env + flags)",
	RunE: func(cmd *cobra.Command, args []string) error {
		for k, v := range viper.AllSettings() {
			fmt.Printf("%s = %v\n", k, v)
		}
		return nil
	},
}

func init() {
	configCmd.AddCommand(configViewCmd)
}
