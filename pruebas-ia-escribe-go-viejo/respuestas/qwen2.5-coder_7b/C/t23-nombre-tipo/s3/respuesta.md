```go
package tarea

import (
	"fmt"
	"reflect"
)

func NombreTipo[T any]() string {
	t := reflect.TypeFor[T]()
	return t.Name()
}
```