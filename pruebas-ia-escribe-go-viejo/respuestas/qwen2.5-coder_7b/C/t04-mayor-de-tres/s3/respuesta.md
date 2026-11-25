```go
package tarea

func MayorDeTres(a, b, c int) int {
    return max(a, max(b, c))
}

func max(a, b int) int {
    if a > b {
        return a
    }
    return b
}
```