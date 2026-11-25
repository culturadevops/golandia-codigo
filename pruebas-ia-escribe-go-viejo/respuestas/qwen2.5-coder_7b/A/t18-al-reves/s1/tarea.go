package tarea

func AlReves(xs []string) []string {
    n := len(xs)
    reversed := make([]string, n)
    for i := range xs {
        reversed[n-1-i] = xs[i]
    }
    return reversed
}
