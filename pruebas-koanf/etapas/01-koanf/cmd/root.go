package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"github.com/spf13/cobra"
)

var cfgFile string

// koanf NO tiene estado global como viper: se crea una instancia y se pasa.
// El "." es el separador de claves anidadas (server.host).
var k = koanf.New(".")

var rootCmd = &cobra.Command{
	Use:   "kcli",
	Short: "kcli con koanf en vez de viper",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "archivo de configuracion (default ./kcli.yaml)")
	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(configCmd)
}

func initConfig() {
	ruta := cfgFile
	if ruta == "" {
		ruta = "kcli.yaml"
	}

	// En koanf el parser es EXPLICITO: se elige por extension a mano.
	// viper lo adivinaba solo; aqui lo decides tu, que es mas codigo y menos magia.
	var parser koanf.Parser = yaml.Parser()
	if strings.HasSuffix(ruta, ".json") {
		parser = json.Parser()
	}

	if err := k.Load(file.Provider(ruta), parser); err != nil {
		fmt.Fprintln(os.Stderr, "sin archivo de configuracion:", err)
	} else {
		fmt.Fprintln(os.Stderr, "usando archivo de configuracion:", ruta)
	}

	// El equivalente al AutomaticEnv + SetEnvKeyReplacer de viper.
	// Aqui la transformacion de KCLI_SERVER_HOST -> server.host es OBLIGATORIA
	// y va escrita a la vista: no hay forma de olvidarla en silencio.
	k.Load(env.Provider("KCLI_", ".", func(s string) string {
		return strings.ReplaceAll(strings.ToLower(strings.TrimPrefix(s, "KCLI_")), "_", ".")
	}), nil)
}
