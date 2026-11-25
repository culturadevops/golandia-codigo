package tarea

import (
	"sync"
)

func ElMasRapido(consultas []func() string) string {
	resultChan := make(chan string, 1)
	var wg sync.WaitGroup

	for _, consulta := range consultas {
		wg.Add(1)
		go func(q func() string) {
			defer wg.Done()
			select {
			case resultChan <- q():
			case <-wg.Done():
			}
		}(consulta)
	}

	wg.Wait()
	return <-resultChan
}
