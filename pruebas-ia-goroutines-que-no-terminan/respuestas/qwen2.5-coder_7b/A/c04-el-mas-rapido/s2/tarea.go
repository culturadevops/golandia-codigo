package tarea

import (
	"sync"
)

func ElMasRapido(consultas []func() string) string {
	var wg sync.WaitGroup
	results := make(chan string, 1)

	for _, consulta := range consultas {
		wg.Add(1)
		go func(consulta func() string) {
			defer wg.Done()
			result := consulta()
			select {
			case results <- result:
			default:
			}
		}(consulta)
	}

	wg.Wait()
	close(results)
	return <-results
}
