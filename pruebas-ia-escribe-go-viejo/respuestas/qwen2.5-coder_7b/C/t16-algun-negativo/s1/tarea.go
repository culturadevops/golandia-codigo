package tarea

import (
	"slices"
)

func AlgunNegativo(nums []int) bool {
	return slices.ContainsFunc(nums, func(n int) bool {
		return n < 0
	})
}
