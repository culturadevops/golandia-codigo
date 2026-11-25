package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	inicio := time.Now()
	var wg sync.WaitGroup
	for i := 1; i <= 3; i++ {
		wg.Go(func() {
			time.Sleep(100 * time.Millisecond)
			fmt.Println("terminó la tarea", i)
		})
	}
	wg.Wait()
	fmt.Printf("main terminó en %d ms\n", time.Since(inicio).Milliseconds()/10*10)
}
