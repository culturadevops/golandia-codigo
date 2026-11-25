```go
package tarea

import (
	"strings"
	"sync"
)

func PalabrasLargas(texto string) int {
	count := atomic.Int64{}
	var wg sync.WaitGroup

	for _, palabra := range strings.Fields(texto) {
		wg.Go(func() {
			if len(palabra) > 3 {
				count.Add(1)
			}
		})
	}

	wg.Wait()
	return int(count.Load())
}
```