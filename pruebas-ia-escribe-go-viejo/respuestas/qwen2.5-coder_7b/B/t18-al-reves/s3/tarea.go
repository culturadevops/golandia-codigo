package tarea

func AlReves(xs []string) []string {
    reversed := make([]string, len(xs))
    for i, j := 0, len(xs)-1; i < len(xs); i, j = i+1, j-1 {
        reversed[i] = xs[j]
    }
    return reversed
}
