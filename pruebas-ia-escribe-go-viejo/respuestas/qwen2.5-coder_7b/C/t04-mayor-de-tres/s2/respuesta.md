```go
package tarea

func MayorDeTres(a, b, c int) int {
    return max(a, max(b, c))
}

func max(x, y int) int {
    if x > y {
        return x
    }
    return y
}
```