// ¿Qué VE el desarrollador cuando la variable de entorno anidada no funciona?
//
// Las dos librerías fallan igual: sin la transformación de KCLI_SERVER_HOST -> server.host,
// ninguna pisa el valor del archivo. La pregunta del vídeo es otra: cuando pasa, ¿tienes
// forma de darte cuenta? Esto imprime lo que cada una expone al depurar.
package main

import (
	"fmt"
	"sort"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"github.com/spf13/viper"
)

func main() {
	fmt.Println("== VIPER, sin SetEnvKeyReplacer ==")
	viper.SetConfigFile("kcli.yaml")
	viper.SetEnvPrefix("KCLI")
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}
	fmt.Println("  server.host =", viper.GetString("server.host"))
	fmt.Println("  lo que expone AllSettings():")
	vs := viper.AllSettings()
	claves := make([]string, 0, len(vs))
	for c := range vs {
		claves = append(claves, c)
	}
	sort.Strings(claves)
	for _, c := range claves {
		fmt.Printf("    %s = %v\n", c, vs[c])
	}

	fmt.Println()
	fmt.Println("== KOANF, sin funcion de transformacion ==")
	k := koanf.New(".")
	if err := k.Load(file.Provider("kcli.yaml"), yaml.Parser()); err != nil {
		panic(err)
	}
	k.Load(env.Provider("KCLI_", ".", nil), nil)
	fmt.Println("  server.host =", k.String("server.host"))
	fmt.Println("  lo que expone All():")
	todo := k.All()
	ks := make([]string, 0, len(todo))
	for c := range todo {
		ks = append(ks, c)
	}
	sort.Strings(ks)
	for _, c := range ks {
		fmt.Printf("    %s = %v\n", c, todo[c])
	}
}
