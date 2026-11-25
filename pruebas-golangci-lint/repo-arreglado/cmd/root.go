package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "kcli",
	Short: "kcli es una CLI de demostracion construida con Cobra y Viper",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// flag persistente: la heredan TODOS los subcomandos, sin declararla de nuevo
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "archivo de configuracion (default ./kcli.yaml)")

	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(configCmd)
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath(".")
		viper.SetConfigName("kcli")
		viper.SetConfigType("yaml")
	}

	// sin esto, una variable de entorno anidada como KCLI_SERVER_HOST
	// nunca llega a pisar la clave "server.host" del archivo: viper la ignora en silencio.
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.SetEnvPrefix("KCLI")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "usando archivo de configuracion:", viper.ConfigFileUsed())
	} else {
		fmt.Fprintln(os.Stderr, "sin archivo de configuracion:", err)
	}
}
