package main

import (
	"fmt"
	"slices"
)

func main() {
	nums := []int{5, 2, 8, 1, 9}
	slices.Sort(nums)
	fmt.Println("slices.Sort:", nums)
	fmt.Println("slices.Contains(8):", slices.Contains(nums, 8))
	fmt.Println("slices.Max:", slices.Max(nums))

	nombres := []string{"go", "generics", "go"}
	fmt.Println("slices.Compact (quita repetidos seguidos):", slices.Compact([]string{"go", "go", "generics"}))
	_ = nombres
}
