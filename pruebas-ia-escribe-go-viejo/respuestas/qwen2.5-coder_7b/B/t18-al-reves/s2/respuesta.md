```go
package tarea

func AlReves(xs []string) []string {
    reversed := make([]string, len(xs))
    for i, j := 0, len(xs)-1; i <= j; i, j = i+1, j-1 {
        reversed[i], reversed[j] = xs[j], xs[i]
    }
    return reversed
}
```