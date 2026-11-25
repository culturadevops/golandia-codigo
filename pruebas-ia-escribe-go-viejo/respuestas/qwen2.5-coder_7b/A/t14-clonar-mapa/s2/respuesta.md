```go
package tarea

import "go/types"

func Clonar(m map[string]int) map[string]int {
    copia := make(map[string]int)
    for clave, valor := range m {
        copia[clave] = valor
    }
    return copia
}
```