package tarea

import (
	"fmt"
	"reflect"
)

func NombreTipo[T any]() string {
	return reflect.TypeOf((*T)(nil)).Elem().Name()
}
