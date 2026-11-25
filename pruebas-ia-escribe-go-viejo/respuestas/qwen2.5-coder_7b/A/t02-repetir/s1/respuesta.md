```go
package tarea

func Repetir(s string, n int) []string {
    resultado := make([]string, n)
    for i := 0; i < n; i++ {
        resultado[i] = s
    }
    return resultado
}
```