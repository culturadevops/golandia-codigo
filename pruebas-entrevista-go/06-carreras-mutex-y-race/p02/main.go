package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

func conMutex() int {
	var mu sync.Mutex
	c := 0
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100000; j++ {
				mu.Lock()
				c++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	return c
}

func conAtomic() int64 {
	var c atomic.Int64
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100000; j++ {
				c.Add(1)
			}
		}()
	}
	wg.Wait()
	return c.Load()
}

// La tercera via, la que Go prefiere cuando encaja: nadie comparte nada.
func conCanal() int {
	res := make(chan int, 2)
	for i := 0; i < 2; i++ {
		go func() {
			c := 0
			for j := 0; j < 100000; j++ {
				c++ // cada goroutine tiene SU contador
			}
			res <- c
		}()
	}
	return <-res + <-res
}

func main() {
	fmt.Println("con mutex: ", conMutex())
	fmt.Println("con atomic:", conAtomic())
	fmt.Println("con canal: ", conCanal())
}
