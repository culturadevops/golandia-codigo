package main

import (
	"fmt"

	"github.com/spf13/viper"
)

func main() {
	// dos "servicios" que asumen que su propio viper.ReadInConfig() los deja aislados
	viper.SetConfigFile("a.yaml")
	viper.ReadInConfig()
	fmt.Println("global, justo despues de leer a.yaml: service =", viper.GetString("service"))

	viper.SetConfigFile("b.yaml")
	viper.ReadInConfig()
	fmt.Println("global, justo despues de leer b.yaml: service =", viper.GetString("service"))

	// el primer llamador vuelve a preguntar, mucho despues, por el MISMO viper global
	fmt.Println("global, preguntado de nuevo mas tarde: service =", viper.GetString("service"))

	fmt.Println()

	// con una instancia por cada uno, las dos conviven a la vez, sin pisarse
	va := viper.New()
	va.SetConfigFile("a.yaml")
	va.ReadInConfig()

	vb := viper.New()
	vb.SetConfigFile("b.yaml")
	vb.ReadInConfig()

	fmt.Println("instancia A (viper.New):", va.GetString("service"))
	fmt.Println("instancia B (viper.New):", vb.GetString("service"))
}
