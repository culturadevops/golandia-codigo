package main

import (
	"fmt"
	"time"
)

func consultar(demora time.Duration) <-chan string {
	c := make(chan string, 1)
	go func() {
		time.Sleep(demora)
		c <- fmt.Sprintf("respuesta tras %v", demora)
	}()
	return c
}

func pedir(demora time.Duration) {
	select {
	case r := <-consultar(demora):
		fmt.Println("ok:", r)
	case <-time.After(500 * time.Millisecond):
		fmt.Println("timeout: me rindo a los 500ms")
	}
}

func main() {
	pedir(100 * time.Millisecond)
	pedir(2 * time.Second)
}
