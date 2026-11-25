```go
package tarea

import (
	"errors"
	"sort"
)

func Mediana(nums []int) float64 {
	if len(nums) == 0 {
		return 0
	}

	sortedNums := make([]int, len(nums))
	copy(sortedNums, nums)
	sort.Ints(sortedNums)

	mid := len(sortedNums) / 2
	if len(sortedNums)%2 == 0 {
		return float64(sortedNums[mid-1]+sortedNums[mid]) / 2.0
	}
	return float64(sortedNums[mid])
}
```