package main

import "fmt"

func cuadrados(nums []int, salida chan<- int) {
	for _, n := range nums {
		salida <- n * n
	}
	close(salida)
}

func main() {
	c := make(chan int)
	go cuadrados([]int{1, 2, 3, 4}, c)
	for v := range c {
		fmt.Println("recibí", v)
	}

	b := make(chan string, 2)
	b <- "uno"
	b <- "dos"
	fmt.Println("con buffer:", <-b, <-b, "· largo", len(b), "capacidad", cap(b))
}
