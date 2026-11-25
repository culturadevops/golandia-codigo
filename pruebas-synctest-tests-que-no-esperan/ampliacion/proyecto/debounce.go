package reintentos

import (
	"sync"
	"time"
)

// Debounce devuelve una función que, llamada muchas veces seguidas, solo ejecuta f una vez:
// cuando pasan `espera` sin ninguna llamada nueva. Como el buscador que espera a que dejes de teclear.
func Debounce(espera time.Duration, f func()) func() {
	var mu sync.Mutex
	var t *time.Timer
	return func() {
		mu.Lock()
		defer mu.Unlock()
		if t != nil {
			t.Stop()
		}
		t = time.AfterFunc(espera, f)
	}
}
