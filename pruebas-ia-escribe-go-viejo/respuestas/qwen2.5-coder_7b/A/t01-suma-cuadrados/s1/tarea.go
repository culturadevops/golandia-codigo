package tarea

func SumaCuadrados(n int) int {
    suma := 0
    for i := 0; i < n; i++ {
        suma += i * i
    }
    return suma
}
