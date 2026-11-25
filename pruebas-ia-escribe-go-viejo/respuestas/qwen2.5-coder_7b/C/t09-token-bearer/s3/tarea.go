package tarea

import (
	"strings"
)

func TokenBearer(cabecera string) (string, bool) {
	if len(cabecera) < 7 || !strings.HasPrefix(cabecera, "Bearer ") {
		return "", false
	}
	return cabecera[7:], true
}
