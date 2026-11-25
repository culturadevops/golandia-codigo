package ingenuo

import (
	"testing"
	"time"
)

// Pasa en tu máquina. En la del compañero que tiene VIGIA_TIMEOUT en su .bashrc, no.
func TestTimeoutPorDefecto(t *testing.T) {
	if got := Timeout(); got != 10*time.Second {
		t.Fatalf("timeout por defecto: %s, quería 10s", got)
	}
}
