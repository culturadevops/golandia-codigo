package main

import "fmt"

func dividir(a, b int) (int, int) {
	return a / b, a % b
}

func estadisticas(nums ...int) (min, max int) {
	min, max = nums[0], nums[0]
	for _, n := range nums {
		if n < min {
			min = n
		}
		if n > max {
			max = n
		}
	}
	return
}

func aplicar(nums []int, f func(int) int) []int {
	out := []int{}
	for _, n := range nums {
		out = append(out, f(n))
	}
	return out
}

func main() {
	c, r := dividir(17, 5)
	fmt.Println("17 / 5 =", c, "resto", r)
	_, resto := dividir(10, 3)
	fmt.Println("solo el resto:", resto)
	fmt.Println(estadisticas(7, 2, 9, 4))
	doble := func(n int) int { return n * 2 }
	fmt.Println(aplicar([]int{1, 2, 3}, doble))
}
