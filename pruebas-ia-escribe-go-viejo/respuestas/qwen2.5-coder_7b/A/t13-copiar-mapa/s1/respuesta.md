```go
package tarea

func Mezclar(dst, src map[string]int) {
	for k, v := range src {
		dst[k] = v
	}
}
```