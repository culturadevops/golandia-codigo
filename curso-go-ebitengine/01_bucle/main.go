// Lección 1 · El bucle de fotogramas. Un juego en Ebitengine son tres métodos: Update (la lógica),
// Draw (el dibujo) y Layout (el tamaño). Update se llama a un ritmo FIJO (60 veces por segundo por
// defecto). Draw, las veces que pueda la pantalla. No son lo mismo, y aquí se mide.
package main

import (
	"fmt"
	"image/color"
	"log"
	"os"
	"time"

	"curso-ebiten/comun"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type Juego struct{ c *comun.Contador }

func (j *Juego) Update() error { return j.c.Tick() }

func (j *Juego) Draw(p *ebiten.Image) {
	j.c.Draws++
	p.Fill(color.RGBA{0x0d, 0x11, 0x17, 0xff})
	ebitenutil.DebugPrint(p, fmt.Sprintf("Update: %d  ·  Draw: %d\nTPS: %.0f  ·  FPS: %.0f",
		j.c.Updates, j.c.Draws, ebiten.ActualTPS(), ebiten.ActualFPS()))
}

func (j *Juego) Layout(w, h int) (int, int) { return 320, 180 }

func main() {
	vsync := os.Getenv("VSYNC") != "no"
	ebiten.SetVsyncEnabled(vsync)
	ebiten.SetWindowSize(960, 540)
	ebiten.SetWindowTitle("Lección 1")
	j := &Juego{c: comun.NuevoContador(5 * time.Second)}
	j.c.Captura = comun.Foto("01_bucle")
	if err := ebiten.RunGame(j); err != nil {
		log.Fatal(err)
	}
	j.c.Resumen(fmt.Sprintf("vsync %v", map[bool]string{true: "sí", false: "no"}[vsync]))
}
