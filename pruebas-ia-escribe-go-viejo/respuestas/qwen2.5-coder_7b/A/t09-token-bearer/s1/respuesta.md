```go
package tarea

import (
	"strings"
)

func TokenBearer(cabecera string) (string, bool) {
	if strings.HasPrefix(cabecera, "Bearer ") {
		return strings.TrimPrefix(cabecera, "Bearer "), true
	}
	return "", false
}
```