
// e04 · LO QUE NO SE DEBE HACER: tocar la interfaz desde una gorrutina, a pelo.
// Fyne 2.6 introdujo fyne.Do/DoAndWait. La 2.9 lo hace OBLIGATORIO.
package main

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func main() {
	a := app.New()
	w := a.NewWindow("hilos · MAL")
	etiqueta := widget.NewLabel("esperando...")
	w.SetContent(container.NewVBox(widget.NewLabel("Subiendo a la nube"), etiqueta))
	w.Resize(fyne.NewSize(640, 200))

	go func() {
		for i := 1; i <= 5; i++ {
			time.Sleep(300 * time.Millisecond)
			// ⛔ SIN fyne.Do: se toca el widget desde OTRA gorrutina.
			etiqueta.SetText(fmt.Sprintf("subido %d de 5", i))
		}
	}()

	w.ShowAndRun()
}
