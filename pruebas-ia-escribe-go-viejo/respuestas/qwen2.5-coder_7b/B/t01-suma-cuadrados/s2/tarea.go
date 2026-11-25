package tarea

func SumaCuadrados(n int) int {
    sum := 0
    for i := 0; i < n; i++ {
        sum += i * i
    }
    return sum
}
