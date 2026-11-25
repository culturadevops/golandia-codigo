// Package ingenuo es la versión que lee el entorno por su cuenta, para ver qué les pasa a
// sus tests en una máquina que no es la tuya.
package ingenuo

import (
	"os"
	"time"
)

// Timeout lee VIGIA_TIMEOUT directamente con os.Getenv.
func Timeout() time.Duration {
	if d, err := time.ParseDuration(os.Getenv("VIGIA_TIMEOUT")); err == nil {
		return d
	}
	return 10 * time.Second
}
