```go
package tarea

import "go/ast"

func Clonar(m map[string]int) map[string]int {
    clone := make(map[string]int)
    for k, v := range m {
        clone[k] = v
    }
    return clone
}
```