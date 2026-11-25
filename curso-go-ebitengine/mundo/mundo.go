// Package mundo: la lógica de «esquiva», sin una sola línea de Ebitengine. Un tick avanza el mundo
// un paso FIJO; nada depende de la hora ni de lo que tarde el dibujado. Por eso se puede probar sin
// ventana, y dos partidas con la misma semilla son idénticas.
package mundo

import "math/rand/v2"

const (
	Ancho, Alto = 320, 180
	Lado        = 12
)

type Bloque struct{ X, Y, Vel float64 }

type Mundo struct {
	JugadorX float64
	Bloques  []Bloque
	Ticks    int
	Vivo     bool
	rng      *rand.Rand
}

func Nuevo(semilla uint64) *Mundo {
	return &Mundo{JugadorX: Ancho / 2, Vivo: true, rng: rand.New(rand.NewPCG(semilla, semilla))}
}

// Paso avanza un tick. dir: -1 izquierda, 0 quieto, +1 derecha.
func (m *Mundo) Paso(dir int) {
	if !m.Vivo {
		return
	}
	m.Ticks++
	m.JugadorX = min(max(m.JugadorX+float64(dir)*3, 0), Ancho-Lado)
	if m.Ticks%18 == 0 { // un bloque nuevo cada 18 ticks
		m.Bloques = append(m.Bloques, Bloque{X: m.rng.Float64() * (Ancho - Lado), Vel: 1.5 + m.rng.Float64()*1.5})
	}
	vivos := m.Bloques[:0]
	for _, b := range m.Bloques {
		b.Y += b.Vel
		if b.Y > Alto-2*Lado && b.Y < Alto-Lado && b.X < m.JugadorX+Lado && b.X+Lado > m.JugadorX {
			m.Vivo = false
		}
		if b.Y < Alto {
			vivos = append(vivos, b)
		}
	}
	m.Bloques = vivos
}

// Piloto: un jugador automático sencillo. Se aparta del bloque más peligroso que tenga encima.
func (m *Mundo) Piloto() int {
	peor, dist := -1, 1e9
	for i, b := range m.Bloques {
		if b.X < m.JugadorX+Lado*2 && b.X+Lado > m.JugadorX-Lado && Alto-b.Y < dist {
			peor, dist = i, Alto-b.Y
		}
	}
	if peor < 0 {
		return 0
	}
	if m.Bloques[peor].X > m.JugadorX {
		return -1
	}
	return 1
}
