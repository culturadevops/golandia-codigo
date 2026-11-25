```go
package tarea

import (
	"strings"
)

func EsAdmin(roles []string) bool {
	return slices.Contains(roles, "admin")
}
```