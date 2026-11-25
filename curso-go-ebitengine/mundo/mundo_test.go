package mundo

import (
	"testing"
	"time"
)

// Lección 6 · Probar la lógica sin ventana. 10.000 ticks son casi tres minutos de partida a 60 por
// segundo; aquí tardan milisegundos, y la misma semilla da la misma partida.
func TestPartidaDeterminista(t *testing.T) {
	jugar := func(semilla uint64) (int, time.Duration) {
		m := Nuevo(semilla)
		t0 := time.Now()
		for i := 0; i < 10_000 && m.Vivo; i++ {
			m.Paso(m.Piloto())
		}
		return m.Ticks, time.Since(t0)
	}
	a, d := jugar(42)
	b, _ := jugar(42)
	c, _ := jugar(7)
	t.Logf("semilla 42: sobrevive %d ticks (%.1f s de juego) · calculado en %v", a, float64(a)/60, d.Round(time.Microsecond))
	t.Logf("semilla 42 otra vez: %d ticks · semilla 7: %d ticks", b, c)
	if a != b {
		t.Fatalf("la misma semilla dio partidas distintas: %d y %d", a, b)
	}
}

func TestChocar(t *testing.T) {
	m := Nuevo(1)
	m.Bloques = []Bloque{{X: m.JugadorX, Y: Alto - 2*Lado - 1, Vel: 2}}
	m.Paso(0)
	if m.Vivo {
		t.Fatal("un bloque encima del jugador tenía que acabar la partida")
	}
}
