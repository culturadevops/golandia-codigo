// Package comun: medir el bucle. Cuenta cuántas veces se llama a Update y a Draw, para cuánto
// tiempo, y termina el juego solo pasado ese tiempo (ebiten.Termination). Así cada lección se
// puede correr sin nadie delante y su salida se guarda literal en SALIDA.md.
package comun

import (
	"fmt"
	"os"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

type Contador struct {
	Updates, Draws int
	Inicio         time.Time
	Dura           time.Duration
	Captura        func() // se llama una vez, a mitad de partida, para fotografiar la ventana
	capturado      bool
}

func NuevoContador(d time.Duration) *Contador { return &Contador{Inicio: time.Now(), Dura: d} }

// Tick va al principio de Update: cuenta y dice si hay que terminar.
func (c *Contador) Tick() error {
	c.Updates++
	if !c.capturado && c.Captura != nil && time.Since(c.Inicio) > c.Dura/2 {
		c.capturado = true
		c.Captura()
	}
	if time.Since(c.Inicio) > c.Dura {
		return ebiten.Termination
	}
	return nil
}

func (c *Contador) Resumen(etiqueta string) {
	s := time.Since(c.Inicio).Seconds()
	fmt.Printf("  %-34s %5.1f s · Update: %5d (%5.1f/s) · Draw: %5d (%5.1f/s)\n",
		etiqueta, s, c.Updates, float64(c.Updates)/s, c.Draws, float64(c.Draws)/s)
}

// Foto pide a la pantalla virtual una captura (lo hace correr.sh con ImageMagick).
func Foto(nombre string) func() {
	return func() {
		if f, err := os.Create("/tmp/foto-" + nombre); err == nil {
			f.Close()
		}
	}
}
