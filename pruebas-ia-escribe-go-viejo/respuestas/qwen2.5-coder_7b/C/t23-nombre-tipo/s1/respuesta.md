```go
package tarea

import (
	"fmt"
	"reflect"
)

func NombreTipo[T any]() string {
	t := reflect.TypeOf(T{})
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.Name()
}
```