package tarea

import "slices"

// Mediana ordena una copia de nums de menor a mayor y devuelve la mediana.
// Para un slice vacío devuelve 0.
func Mediana(nums []int) float64 {
	n := len(nums)
	if n == 0 {
		return 0
	}
	ordenados := slices.Clone(nums)
	slices.Sort(ordenados)
	mitad := n / 2
	if n%2 == 1 {
		return float64(ordenados[mitad])
	}
	return (float64(ordenados[mitad-1]) + float64(ordenados[mitad])) / 2
}
