```go
package tarea

import (
	"sync"
)

func Mezclar(dst, src map[string]int) {
	wg := sync.WaitGroup{}
	for key, value := range src {
		wg.Go(func() {
			dst[key] = value
		})
	}
	wg.Wait()
}
```