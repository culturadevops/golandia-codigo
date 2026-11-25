```go
package tarea

import (
	"sync"
)

func ElMasRapido(consultas []func() string) string {
	type Result struct {
		value string
		err   error
	}

	results := make(chan Result, len(consultas))
	var wg sync.WaitGroup

	for _, consulta := range consultas {
		wg.Add(1)
		go func(f func() string) {
			defer wg.Done()
			result := f()
			results <- Result{value: result}
		}(consulta)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	select {
	case res := <-results:
		return res.value
	}
}
```