```go
package tarea

import (
	"context"
	"sync"
)

func Unir(ctx context.Context, canales ...<-chan int) <-chan int {
	result := make(chan int)
	closeChan := make(chan struct{})

	go func() {
		defer close(result)
		defer close(closeChan)

		wg := sync.WaitGroup{}
		wg.Add(len(canales))

		for _, canal := range canales {
			go func(c <-chan int) {
				defer wg.Done()
				for {
					select {
					case <-ctx.Done():
						return
					case value, ok := <-c:
						if !ok {
							return
						}
						select {
						case <-ctx.Done():
							return
						case result <- value:
						}
					}
				}
			}(canal)
		}

		wg.Wait()
	}()

	go func() {
		select {
		case <-ctx.Done():
		case <-closeChan:
		}
		close(result)
	}()

	return result
}
```