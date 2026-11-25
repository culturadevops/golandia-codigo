package main

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

func main() {
	viper.SetConfigFile("config.yaml")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.SetEnvPrefix("KCLI")
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}
	fmt.Println("CON replacer, server.host =", viper.GetString("server.host"))
}
