```go
package tarea

import (
	"strings"
)

func Dominio(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) > 1 {
		return parts[1]
	}
	return ""
}
```