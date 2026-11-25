```go
package tarea

import (
	"context"
	"sync"
)

func Unir(ctx context.Context, canales ...<-chan int) <-chan int {
	out := make(chan int)
	var wg sync.WaitGroup
	wg.Add(len(canales))

	for _, c := range canales {
		go func(c <-chan int) {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case val, ok := <-c:
					if !ok {
						return
					}
					select {
					case <-ctx.Done():
						return
					case out <- val:
					}
				}
			}
		}(c)
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}
```