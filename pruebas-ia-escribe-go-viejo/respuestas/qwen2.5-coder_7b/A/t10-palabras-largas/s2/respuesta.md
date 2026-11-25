```go
package tarea

func PalabrasLargas(texto string) int {
    palabras := 0
    for _, palabra := range strings.Fields(texto) {
        if len(palabra) > 3 {
            palabras++
        }
    }
    return palabras
}
```