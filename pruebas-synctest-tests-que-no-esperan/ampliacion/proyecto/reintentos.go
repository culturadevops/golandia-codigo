package reintentos

import (
	"context"
	"errors"
	"time"
)

var ErrRendido = errors.New("me rindo: 5 intentos")

// Hacer llama a f hasta 5 veces, esperando 1, 2, 4 y 8 segundos entre intentos.
func Hacer(ctx context.Context, f func() error) (int, error) {
	espera := time.Second
	for intento := 1; ; intento++ {
		if f() == nil {
			return intento, nil
		}
		if intento == 5 {
			return intento, ErrRendido
		}
		select {
		case <-time.After(espera):
		case <-ctx.Done():
			return intento, ctx.Err()
		}
		espera *= 2
	}
}
