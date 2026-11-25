package tarea

import (
	"sync"
)

func ElMasRapido(consultas []func() string) string {
	var wg sync.WaitGroup
	resultChan := make(chan string, 1)

	for _, consulta := range consultas {
		wg.Add(1)
		go func(f func() string) {
			defer wg.Done()
			resultChan <- f()
		}(consulta)
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	return <-resultChan
}
