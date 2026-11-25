// goleak contra CÓDIGO REAL DEL CANAL, no contra un ejemplo escrito para quedar bien.
//
// `chequeo.go` es tal cual el del vídeo «tu chequeo de salud te miente» (cultura-devops).
// Lo único que se tocó: el nombre del paquete y el de `main`, para poder importarlo en un
// test. La lógica no se ha cambiado ni una línea.
//
// ⚠️ La sospecha de la que sale esto: en aquel vídeo, al cerrar, apareció
//    «httptest.Server blocked in Close after 5 seconds, waiting for connections».
//    Es decir: el context.WithTimeout cancela TU espera, pero la petición al servicio
//    colgado sigue viva. Si eso es verdad, goleak tiene que verlo.
package chequeo

import (
	"testing"
	"time"

	"go.uber.org/goleak"
)

// La versión concurrente correcta, la que el vídeo presentó como buena.
func TestConcurrente(t *testing.T) {
	defer goleak.VerifyNone(t)

	s := servidor()
	defer s.Close()
	r := concurrente(s.URL)
	if len(r) != servicios {
		t.Fatalf("esperaba %d servicios, salieron %d", servicios, len(r))
	}
}

// La del límite de tiempo, que es donde estaba la sospecha.
func TestConTimeout(t *testing.T) {
	defer goleak.VerifyNone(t)

	s := servidor()
	defer s.Close()
	vivos, caidos := conTimeout(s.URL, 1*time.Second)
	if vivos != servicios || caidos != 1 {
		t.Fatalf("esperaba %d vivos y 1 caido, salieron %d y %d", servicios, vivos, caidos)
	}
}

// ⚠️ ¿Por qué el test de arriba pasó si la conexión se quedó viva 30 segundos?
// Hipótesis: los `defer` salen en orden inverso, así que `s.Close()` corre ANTES que
// goleak y se pasa 30 segundos esperando a la petición colgada. Cuando goleak por fin
// mira, la goroutine ya murió. El mismo test, con goleak mirando PRIMERO:
func TestConTimeout_GoleakPrimero(t *testing.T) {
	s := servidor()
	// ⚠️ Los defer salen en orden INVERSO: el ultimo declarado corre primero.
	// Asi que para que goleak mire ANTES del Close, goleak va declarado DESPUES.
	defer s.Close()            // declarado 1º => corre el ULTIMO
	defer goleak.VerifyNone(t) // declarado 2º => corre el PRIMERO

	vivos, caidos := conTimeout(s.URL, 1*time.Second)
	if vivos != servicios || caidos != 1 {
		t.Fatalf("esperaba %d vivos y 1 caido, salieron %d y %d", servicios, vivos, caidos)
	}
}
