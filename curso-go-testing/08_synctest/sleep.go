// Lección 8 · testing/synctest para código con tiempo.
//
// Se prueba el cliente de cobros de verdad (tienda/cobros): reintenta los 5xx esperando
// 1 s, 2 s, 4 s. Y aquí, la PRIMERA versión de ese bucle, con time.Sleep.
package relojes

import (
	"context"
	"time"
)

// ReintentarConSleep es como se escribió primero. Funciona... salvo que time.Sleep no se
// entera de que el context se canceló: espera hasta el final de la espera en curso.
func ReintentarConSleep(ctx context.Context, intentos int, espera time.Duration, f func() error) error {
	var err error
	for i := range intentos {
		if err = f(); err == nil {
			return nil
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if i+1 < intentos {
			time.Sleep(espera << i)
		}
	}
	return err
}
