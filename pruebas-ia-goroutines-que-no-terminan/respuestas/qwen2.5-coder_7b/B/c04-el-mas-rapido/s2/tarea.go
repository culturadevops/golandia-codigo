package tarea

import (
	"sync"
)

func ElMasRapido(consultas []func() string) string {
	var wg sync.WaitGroup
	resultChan := make(chan string, 1)

	for _, consulta := range consultas {
		wg.Add(1)
		go func(consulta func() string) {
			defer wg.Done()
			select {
			case resultChan <- consulta():
			case <-wg.Done():
			}
		}(consulta)
	}

	wg.Wait()
	return <-resultChan
}
