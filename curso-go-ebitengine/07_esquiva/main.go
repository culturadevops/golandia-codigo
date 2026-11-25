// Lección 7 · El proyecto: «esquiva». La lógica es el paquete mundo (lección 6); aquí solo se lee
// la entrada en Update y se dibuja en Draw. Con DEMO=1 juega el piloto automático.
package main

import (
	"fmt"
	"image/color"
	"log"
	"os"
	"time"

	"curso-ebiten/comun"
	"curso-ebiten/mundo"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type Juego struct {
	c    *comun.Contador
	m    *mundo.Mundo
	demo bool
}

func (j *Juego) Update() error {
	dir := 0
	if j.demo {
		dir = j.m.Piloto()
	} else if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
		dir = -1
	} else if ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
		dir = 1
	}
	j.m.Paso(dir) // la lógica: un paso fijo por tick
	return j.c.Tick()
}

func (j *Juego) Draw(p *ebiten.Image) {
	j.c.Draws++
	p.Fill(color.RGBA{0x0d, 0x11, 0x17, 0xff})
	for _, b := range j.m.Bloques {
		vector.DrawFilledRect(p, float32(b.X), float32(b.Y), mundo.Lado, mundo.Lado, color.RGBA{0xff, 0x7b, 0x72, 0xff}, false)
	}
	vector.DrawFilledRect(p, float32(j.m.JugadorX), mundo.Alto-mundo.Lado, mundo.Lado, mundo.Lado, color.RGBA{0x00, 0xad, 0xd8, 0xff}, false)
	ebitenutil.DebugPrint(p, fmt.Sprintf("esquiva · %.1f s · %d bloques", float64(j.m.Ticks)/60, len(j.m.Bloques)))
}

func (j *Juego) Layout(w, h int) (int, int) { return mundo.Ancho, mundo.Alto }

func main() {
	ebiten.SetWindowSize(960, 540)
	ebiten.SetWindowTitle("esquiva")
	j := &Juego{c: comun.NuevoContador(8 * time.Second), m: mundo.Nuevo(42), demo: os.Getenv("DEMO") == "1"}
	j.c.Captura = comun.Foto("07_esquiva")
	if err := ebiten.RunGame(j); err != nil {
		log.Fatal(err)
	}
	j.c.Resumen("esquiva, piloto automático, semilla 42")
	fmt.Printf("    ticks de juego: %d · vivo: %v · bloques en pantalla: %d\n", j.m.Ticks, j.m.Vivo, len(j.m.Bloques))
}
