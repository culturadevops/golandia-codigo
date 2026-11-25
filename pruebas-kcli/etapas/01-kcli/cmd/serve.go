package cmd

import "github.com/spf13/cobra"

// "serve" no hace nada por si sola: solo agrupa a sus hijas, igual que
// "kubectl config" agrupa a "view", "set-context", etc.
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "arranca o detiene el servidor de demostracion",
}

func init() {
	serveCmd.AddCommand(serveStartCmd)
	serveCmd.AddCommand(serveStopCmd)
}
