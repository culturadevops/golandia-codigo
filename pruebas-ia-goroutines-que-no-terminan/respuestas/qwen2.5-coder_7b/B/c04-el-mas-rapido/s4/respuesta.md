```go
package tarea

import (
	"sync"
)

func ElMasRapido(consultas []func() string) string {
	result := make(chan string, 1)
	var wg sync.WaitGroup

	for _, consulta := range consultas {
		wg.Add(1)
		go func(q func() string) {
			defer wg.Done()
			select {
			case result <- q():
			default:
				// Si el resultado ya está disponible, no enviamos el nuevo
			}
		}(consulta)
	}

	go func() {
		wg.Wait()
		close(result)
	}()

	return <-result
}
```