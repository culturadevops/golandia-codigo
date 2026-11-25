package tarea

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
)

func Obtener(cfg map[string]any, clave string, porDefecto any) any {
	if val, ok := cfg[clave]; ok {
		return val
	}
	return porDefecto
}
