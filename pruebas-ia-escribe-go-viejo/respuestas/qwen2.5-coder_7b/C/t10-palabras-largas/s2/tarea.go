package tarea

import (
	"strings"
	"sync"
)

func PalabrasLargas(texto string) int {
	words := strings.Fields(texto)
	count := atomic.Int64{}

	var wg sync.WaitGroup
	for _, word := range words {
		wg.Add(1)
		go func(w string) {
			defer wg.Done()
			if len(w) > 3 {
				count.Add(1)
			}
		}(word)
	}

	wg.Wait()
	return int(count.Load())
}
