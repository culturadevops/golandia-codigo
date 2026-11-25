package main

import "fmt"

func suma(nums [3]int) int {
	return nums[0] + nums[1] + nums[2]
}

func main() {
	cuatro := [4]int{1, 2, 3, 4}
	fmt.Println(suma(cuatro))
}
