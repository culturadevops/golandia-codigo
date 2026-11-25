```go
package tarea

import (
	"strings"
)

func AlReves(xs []string) []string {
	var result []string
	for i := len(xs) - 1; i >= 0; i-- {
		result = append(result, xs[i])
	}
	return result
}
```