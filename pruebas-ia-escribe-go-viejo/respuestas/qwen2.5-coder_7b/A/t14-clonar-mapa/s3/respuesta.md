```go
package tarea

func Clonar(m map[string]int) map[string]int {
    copia := make(map[string]int, len(m))
    for clave, valor := range m {
        copia[clave] = valor
    }
    return copia
}
```