```go
package tarea

import (
	"sync"
)

func AlgunNegativo(nums []int) bool {
	var wg sync.WaitGroup
	var found atomic.Int64

	for _, num := range nums {
		wg.Go(func(num int) func() {
			if num < 0 {
				found.Store(1)
			}
			return func() {}
		}(num))
	}

	wg.Wait()
	return found.Load() == 1
}
```