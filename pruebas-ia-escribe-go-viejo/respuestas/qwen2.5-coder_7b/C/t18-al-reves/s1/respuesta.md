```go
package tarea

import (
	"strings"
)

func AlReves(xs []string) []string {
	n := len(xs)
	result := make([]string, n)
	for i := range xs {
		result[n-1-i] = xs[i]
	}
	return result
}
```