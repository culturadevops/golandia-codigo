```go
package tarea

import "sort"

// Mediana ordena una copia de nums de menor a mayor y devuelve la mediana.
// Para un slice vacío devuelve 0.
func Mediana(nums []int) float64 {
	n := len(nums)
	if n == 0 {
		return 0
	}

	copia := make([]int, n)
	copy(copia, nums)
	sort.Ints(copia)

	mitad := n / 2
	if n%2 == 1 {
		return float64(copia[mitad])
	}
	return (float64(copia[mitad-1]) + float64(copia[mitad])) / 2
}
```