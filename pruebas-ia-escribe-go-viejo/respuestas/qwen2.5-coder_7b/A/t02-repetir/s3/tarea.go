package tarea

func Repetir(s string, n int) []string {
    result := make([]string, n)
    for i := 0; i < n; i++ {
        result[i] = s
    }
    return result
}
