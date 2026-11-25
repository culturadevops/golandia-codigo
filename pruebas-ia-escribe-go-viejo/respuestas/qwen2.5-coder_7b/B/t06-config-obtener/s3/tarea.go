package tarea

import (
	"errors"
)

func Obtener(cfg map[string]interface{}, clave string, porDefecto interface{}) (interface{}, error) {
	valor, existe := cfg[clave]
	if existe {
		return valor, nil
	}
	return porDefecto, errors.New("clave no encontrada")
}
