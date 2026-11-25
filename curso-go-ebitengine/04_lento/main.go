// Lección 4 · Lo que tarda el dibujado no frena la lógica... pero lo que tarda la lógica, sí.
// Draw lento: menos fotogramas, la lógica sigue a 60 por segundo. Update lento: el juego entero
// va a cámara lenta, porque la lógica no da abasto.
package main

import (
	"fmt"
	"image/color"
	"log"
	"os"
	"time"

	"curso-ebiten/comun"

	"github.com/hajimehoshi/ebiten/v2"
)

type Juego struct {
	c          *comun.Contador
	drawLento  time.Duration
	updateLento time.Duration
}

func (j *Juego) Update() error {
	time.Sleep(j.updateLento)
	return j.c.Tick()
}

func (j *Juego) Draw(p *ebiten.Image) {
	j.c.Draws++
	time.Sleep(j.drawLento)
	p.Fill(color.Black)
}

func (j *Juego) Layout(w, h int) (int, int) { return 320, 180 }

func main() {
	d, _ := time.ParseDuration(os.Getenv("DRAW") + "ms")
	u, _ := time.ParseDuration(os.Getenv("UPDATE") + "ms")
	ebiten.SetWindowSize(960, 540)
	j := &Juego{c: comun.NuevoContador(3 * time.Second), drawLento: d, updateLento: u}
	if err := ebiten.RunGame(j); err != nil {
		log.Fatal(err)
	}
	j.c.Resumen(fmt.Sprintf("Draw +%v · Update +%v", d, u))
}
