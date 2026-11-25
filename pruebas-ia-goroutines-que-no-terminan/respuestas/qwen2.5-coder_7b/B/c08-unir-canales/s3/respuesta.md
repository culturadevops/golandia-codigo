```go
package tarea

import (
	"context"
	"sync"
)

func Unir(ctx context.Context, canales ...<-chan int) <-chan int {
	out := make(chan int)
	var wg sync.WaitGroup
	closeChan := make(chan struct{})

	go func() {
		defer close(closeChan)
		for _, c := range canales {
			wg.Add(1)
			go func(c <-chan int) {
				defer wg.Done()
				for {
					select {
					case <-ctx.Done():
						return
					case v, ok := <-c:
						if !ok {
							return
						}
						select {
						case out <- v:
						case <-ctx.Done():
						case <-closeChan:
						}
					case <-ctx.Done():
						return
					case <-closeChan:
						return
					}
				}
			}(c)
		}
	}()

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}
```