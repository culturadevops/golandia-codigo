```go
package tarea

import (
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
)

func Obtener(cfg map[string]any, clave string, porDefecto any) any {
	if valor, existe := cfg[clave]; existe {
		return valor
	}
	return porDefecto
}
```