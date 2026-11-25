```go
package tarea

func Mezclar(dst, src map[string]int) {
    for key, value := range src {
        dst[key] = value
    }
}
```