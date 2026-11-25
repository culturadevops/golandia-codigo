package tarea

import (
	"sync"
)

func ElMasRapido(consultas []func() string) string {
	var wg sync.WaitGroup
	resultChan := make(chan string, 1)

	for _, consulta := range consultas {
		wg.Add(1)
		go func(c func() string) {
			defer wg.Done()
			resultChan <- c()
		}(consulta)
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	return <-resultChan
}
