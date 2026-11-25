// Lección 3 · Tiempo fijo: la velocidad, por segundo y no por tick. Si la lógica suma «1 px por
// Update» y alguien cambia el ritmo (SetTPS), el juego entero va a otra velocidad. Si suma
// «60 px por segundo / TPS», da igual el ritmo.
package main

import (
	"fmt"
	"image/color"
	"log"
	"os"
	"strconv"
	"time"

	"curso-ebiten/comun"

	"github.com/hajimehoshi/ebiten/v2"
)

const velocidad = 60.0 // píxeles por segundo

type Juego struct {
	c                   *comun.Contador
	porTick, porSegundo float64
}

func (j *Juego) Update() error {
	j.porTick += 1                                     // ⛔ supone 60 ticks por segundo
	j.porSegundo += velocidad / float64(ebiten.TPS()) // ✅ vale para cualquier ritmo
	return j.c.Tick()
}

func (j *Juego) Draw(p *ebiten.Image)          { j.c.Draws++; p.Fill(color.Black) }
func (j *Juego) Layout(w, h int) (int, int) { return 320, 180 }

func main() {
	tps, _ := strconv.Atoi(os.Getenv("TPS"))
	ebiten.SetTPS(tps)
	ebiten.SetWindowSize(960, 540)
	j := &Juego{c: comun.NuevoContador(3 * time.Second)}
	if err := ebiten.RunGame(j); err != nil {
		log.Fatal(err)
	}
	j.c.Resumen(fmt.Sprintf("SetTPS(%d)", tps))
	fmt.Printf("    «1 px por tick»: %4.0f px  ·  «%0.f px por segundo»: %4.0f px  (3 s: deberían ser 180)\n",
		j.porTick, velocidad, j.porSegundo)
}
