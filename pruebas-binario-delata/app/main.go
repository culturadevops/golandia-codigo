// Una herramienta normal y corriente: lee un YAML de configuración, se conecta
// a una API y saca una tabla. Lo que cualquiera reparte sin pensarlo dos veces.
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"miempresa.interno/herramienta/interno"
)

type conf struct {
	Endpoint string `yaml:"endpoint"`
	Clave    string `yaml:"clave"`
}

func main() {
	raiz := &cobra.Command{
		Use:   "herramienta",
		Short: "la herramienta interna de la empresa",
		RunE: func(cmd *cobra.Command, args []string) error {
			var c conf
			datos, err := os.ReadFile("config.yaml")
			if err != nil {
				return err
			}
			if err := yaml.Unmarshal(datos, &c); err != nil {
				return err
			}
			fmt.Println(interno.Saludo(c.Endpoint))
			return nil
		},
	}
	if err := raiz.Execute(); err != nil {
		os.Exit(1)
	}
}
