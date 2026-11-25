```go
package tarea

// Acotar devuelve x limitado al rango [lo, hi].
func Acotar(x, lo, hi int) int {
	if x < lo {
		return lo
	}
	if x > hi {
		return hi
	}
	return x
}
```