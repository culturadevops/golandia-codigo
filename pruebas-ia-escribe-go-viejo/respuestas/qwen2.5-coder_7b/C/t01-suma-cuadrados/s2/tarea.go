package tarea

func SumaCuadrados(n int) int {
    sum := 0
    for i := range n {
        sum += i * i
    }
    return sum
}
