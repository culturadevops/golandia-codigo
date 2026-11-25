package main

import "fmt"

func procesar(datos []int) {
	defer fmt.Println("defer: limpio aunque haya pánico")
	fmt.Println("el tercero es", datos[2])
}

func main() {
	procesar([]int{1, 2})
	fmt.Println("esto ya no sale")
}
