package main

import (
	"fmt"
	"plugin"
)

func main() {
	_, err := plugin.Open("x.so")
	fmt.Println(err)
}
