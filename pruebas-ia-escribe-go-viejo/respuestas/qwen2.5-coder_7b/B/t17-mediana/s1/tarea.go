package tarea

import (
	"sort"
)

func Mediana(nums []int) float64 {
	if len(nums) == 0 {
		return 0
	}

	sortedNums := make([]int, len(nums))
	copy(sortedNums, nums)
	sort.Ints(sortedNums)

	middle := len(sortedNums) / 2
	if len(sortedNums)%2 == 0 {
		return float64(sortedNums[middle-1]+sortedNums[middle]) / 2.0
	}
	return float64(sortedNums[middle])
}
