package tarea

import (
	"strings"
	"sync"
)

func PalabrasLargas(texto string) int {
	words := strings.Fields(texto)
	count := 0

	var wg sync.WaitGroup
	for _, word := range words {
		wg.Add(1)
		go func(word string) {
			defer wg.Done()
			if len(word) > 3 {
				atomic.AddInt64(&count, 1)
			}
		}(word)
	}

	wg.Wait()
	return int(count)
}
