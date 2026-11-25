package tarea

// SumaCuadrados devuelve la suma de i*i para i desde 0 hasta n-1.
func SumaCuadrados(n int) int {
	suma := 0
	for i := 0; i < n; i++ {
		suma += i * i
	}
	return suma
}
