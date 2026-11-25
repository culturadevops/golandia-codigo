
// e04 · LO CORRECTO: el trabajo en la gorrutina, el retoque de la interfaz dentro de fyne.Do.
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
	w := a.NewWindow("hilos · BIEN")
	etiqueta := widget.NewLabel("esperando...")
	w.SetContent(container.NewVBox(widget.NewLabel("Subiendo a la nube"), etiqueta))
	w.Resize(fyne.NewSize(640, 200))

	go func() {
		for i := 1; i <= 5; i++ {
			time.Sleep(300 * time.Millisecond) // el TRABAJO va aqui, fuera
			n := i
			fyne.Do(func() { // y SOLO el retoque de la interfaz va dentro
				etiqueta.SetText(fmt.Sprintf("subido %d de 5", n))
			})
		}
	}()

	w.ShowAndRun()
}
