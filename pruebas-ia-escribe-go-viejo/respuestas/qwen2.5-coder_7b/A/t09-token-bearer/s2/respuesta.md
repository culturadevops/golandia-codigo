```go
package tarea

import (
	"strings"
)

func TokenBearer(cabecera string) (string, bool) {
	if strings.HasPrefix(cabecera, "Bearer ") {
		return cabecera[7:], true
	}
	return "", false
}
```