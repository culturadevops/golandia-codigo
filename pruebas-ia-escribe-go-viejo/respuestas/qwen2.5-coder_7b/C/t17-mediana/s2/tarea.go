package tarea

import (
	"sort"
)

func Mediana(nums []int) float64 {
	if len(nums) == 0 {
		return 0
	}

	copied := make([]int, len(nums))
	copy(copied, nums)
	sort.Ints(copied)

	middle := len(copied) / 2
	if len(copied)%2 == 0 {
		return float64(copied[middle-1]+copied[middle]) / 2.0
	}
	return float64(copied[middle])
}
