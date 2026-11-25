```go
package tarea

func AlReves(xs []string) []string {
    n := len(xs)
    reversed := make([]string, n)
    for i := 0; i < n; i++ {
        reversed[i] = xs[n-1-i]
    }
    return reversed
}
```