package main

import (
	"fmt"
	"reflect"
)

type Saludo struct{}

func (Saludo) Hola(n string) string { return "hola, " + n }

func main() {
	m := reflect.ValueOf(Saludo{}).MethodByName("Hola")
	fmt.Println(m.Call([]reflect.Value{reflect.ValueOf("Ana")})[0])
}
