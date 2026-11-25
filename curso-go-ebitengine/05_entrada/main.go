// Lección 5 · La entrada: «está pulsada» no es «se acaba de pulsar». Un disparo con
// IsKeyPressed dispara en CADA tick mientras la tecla siga abajo. Una pulsación de medio segundo
// son treinta disparos. inpututil.IsKeyJustPressed cuenta la pulsación una vez.
// La tecla la pulsa xdotool desde fuera (correr.sh), como una persona.
package main

import (
	"fmt"
	"image/color"
	"log"
	"time"

	"curso-ebiten/comun"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type Juego struct {
	c                      *comun.Contador
	conPressed, conJust    int
}

func (j *Juego) Update() error {
	if ebiten.IsKeyPressed(ebiten.KeySpace) {
		j.conPressed++ // ⛔ un disparo por tick
	}
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		j.conJust++ // ✅ un disparo por pulsación
	}
	return j.c.Tick()
}

func (j *Juego) Draw(p *ebiten.Image)          { j.c.Draws++; p.Fill(color.Black) }
func (j *Juego) Layout(w, h int) (int, int) { return 320, 180 }

func main() {
	ebiten.SetWindowSize(960, 540)
	ebiten.SetWindowTitle("Lección 5")
	j := &Juego{c: comun.NuevoContador(5 * time.Second)}
	if err := ebiten.RunGame(j); err != nil {
		log.Fatal(err)
	}
	j.c.Resumen("dos pulsaciones de espacio")
	fmt.Printf("    disparos con IsKeyPressed: %d  ·  con inpututil.IsKeyJustPressed: %d\n", j.conPressed, j.conJust)
}
