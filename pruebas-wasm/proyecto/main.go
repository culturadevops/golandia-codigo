package main

import "syscall/js"

func sumar(this js.Value, args []js.Value) interface{} {
	a := args[0].Int()
	b := args[1].Int()
	return a + b
}

func esPrimo(this js.Value, args []js.Value) interface{} {
	n := args[0].Int()
	if n < 2 {
		return false
	}
	for i := 2; i*i <= n; i++ {
		if n%i == 0 {
			return false
		}
	}
	return true
}

func main() {
	c := make(chan struct{})
	js.Global().Set("sumarEnGo", js.FuncOf(sumar))
	js.Global().Set("esPrimoEnGo", js.FuncOf(esPrimo))
	println("wasm de Go listo: sumarEnGo y esPrimoEnGo registradas")
	<-c
}
