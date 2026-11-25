package tarea

import (
	"strings"
)

func Dominio(email string) string {
	atIndex := strings.Index(email, "@")
	if atIndex == -1 {
		return ""
	}
	return email[atIndex+1:]
}
