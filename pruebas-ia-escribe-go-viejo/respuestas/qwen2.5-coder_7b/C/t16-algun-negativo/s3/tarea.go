package tarea

import (
	"sync"
)

func AlgunNegativo(nums []int) bool {
	var wg sync.WaitGroup
	var found atomic.Int64

	for i := range nums {
		wg.Go(func(i int) {
			if nums[i] < 0 {
				found.Store(1)
			}
		}(i))
	}

	wg.Wait()
	return found.Load() == 1
}
