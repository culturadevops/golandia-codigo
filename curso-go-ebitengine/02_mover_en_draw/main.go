// Lección 2 · La trampa: mover el personaje en Draw. Funciona en tu máquina. En la de un jugador
// con un monitor más rápido, Draw se llama más veces por segundo... y el personaje corre más.
// Aquí se simulan dos monitores haciendo que cada Draw tarde lo que tarda un fotograma en cada uno.
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
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type Juego struct {
	c                *comun.Contador
	enDraw, enUpdate float64 // la posición de los dos personajes, en píxeles
	fotograma        time.Duration
}

func (j *Juego) Update() error {
	j.enUpdate += 1 // ✅ 1 px por tick: 60 px por segundo, siempre
	return j.c.Tick()
}

func (j *Juego) Draw(p *ebiten.Image) {
	j.c.Draws++
	j.enDraw += 1 // ⛔ 1 px por fotograma: depende del monitor
	time.Sleep(j.fotograma)
	p.Fill(color.RGBA{0x0d, 0x11, 0x17, 0xff})
	vector.DrawFilledRect(p, float32(int(j.enDraw)%320), 50, 16, 16, color.RGBA{0xff, 0x7b, 0x72, 0xff}, false)
	vector.DrawFilledRect(p, float32(int(j.enUpdate)%320), 110, 16, 16, color.RGBA{0x7e, 0xe7, 0x87, 0xff}, false)
}

func (j *Juego) Layout(w, h int) (int, int) { return 320, 180 }

func main() {
	hz, _ := strconv.Atoi(os.Getenv("HZ"))
	ebiten.SetWindowSize(960, 540)
	ebiten.SetWindowTitle("Lección 2")
	j := &Juego{c: comun.NuevoContador(3 * time.Second), fotograma: time.Second / time.Duration(hz)}
	j.c.Captura = comun.Foto(fmt.Sprintf("02_mover-%dhz", hz))
	if err := ebiten.RunGame(j); err != nil {
		log.Fatal(err)
	}
	j.c.Resumen(fmt.Sprintf("monitor simulado de %d Hz", hz))
	fmt.Printf("    movido en Draw:   %4.0f px  · movido en Update: %4.0f px\n", j.enDraw, j.enUpdate)
}
