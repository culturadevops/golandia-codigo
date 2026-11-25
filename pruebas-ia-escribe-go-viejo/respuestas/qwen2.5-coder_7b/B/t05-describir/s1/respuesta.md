```go
package tarea

import (
	"fmt"
)

func Describir(valores ...interface{}) []string {
	resultados := make([]string, len(valores))
	for i, valor := range valores {
		resultados[i] = fmt.Sprintf("%T", valor)
	}
	return resultados
}
```