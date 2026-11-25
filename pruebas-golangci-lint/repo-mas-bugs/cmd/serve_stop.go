package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var serveStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "detiene el servidor (demostracion)",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("servidor detenido")
		return nil
	},
}
