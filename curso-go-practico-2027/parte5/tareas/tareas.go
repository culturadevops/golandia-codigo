package main

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"
)

// Tarea es lo que guarda y devuelve la API.
type Tarea struct {
	ID     int64  `json:"id"`
	Titulo string `json:"titulo"`
	Hecha  bool   `json:"hecha"`
}

const maxTitulo = 120

// ErrNoEncontrada la devuelve el almacén cuando el id no existe.
var ErrNoEncontrada = errors.New("tarea no encontrada")

// validarTitulo quita los espacios de los extremos y comprueba el largo.
func validarTitulo(t string) (string, error) {
	t = strings.TrimSpace(t)
	n := utf8.RuneCountInString(t)
	switch {
	case n == 0:
		return "", errors.New("el título no puede estar vacío")
	case n > maxTitulo:
		return "", fmt.Errorf("el título tiene %d caracteres; el máximo es %d", n, maxTitulo)
	}
	return t, nil
}
