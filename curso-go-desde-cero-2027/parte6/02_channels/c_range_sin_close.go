package main

import "fmt"

func main() {
	c := make(chan int)
	go func() {
		for i := 1; i <= 3; i++ {
			c <- i
		}
	}()
	for v := range c {
		fmt.Println("recibí", v)
	}
}
