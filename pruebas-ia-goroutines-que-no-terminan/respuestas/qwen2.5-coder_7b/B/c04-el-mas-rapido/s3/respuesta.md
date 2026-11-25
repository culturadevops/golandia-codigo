```go
package tarea

import (
	"sync"
)

func ElMasRapido(consultas []func() string) string {
	type result struct {
		value string
		err   error
	}

	ch := make(chan result, len(consultas))
	var wg sync.WaitGroup

	for _, consulta := range consultas {
		wg.Add(1)
		go func(consulta func() string) {
			defer wg.Done()
			res := consulta()
			ch <- result{value: res, err: nil}
		}(consulta)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	select {
	case res := <-ch:
		return res.value
	}
}
```