```go
package tarea

import "fmt"

func Obtener(cfg map[string]interface{}, clave string, porDefecto interface{}) interface{} {
    valor, existe := cfg[clave]
    if existe {
        return valor
    }
    return porDefecto
}
```