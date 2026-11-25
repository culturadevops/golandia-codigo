// e05 · widget.List NO es un VBox: solo construye las filas que se VEN.
// 50.000 objetos en un bucket, y la ventana abre al instante.
package main

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func main() {
	claves := make([]string, 50000)
	for i := range claves {
		claves[i] = fmt.Sprintf("2026/informes/%05d-informe.pdf", i)
	}

	construidas := 0 // cuantas filas ha creado Fyne de verdad

	a := app.New()
	w := a.NewWindow("nubecita · 50.000 objetos")
	cuenta := widget.NewLabel("")

	lista := widget.NewList(
		func() int { return len(claves) },
		func() fyne.CanvasObject { // se llama UNA vez por fila VISIBLE
			construidas++
			return widget.NewLabel("plantilla")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(claves[i])
		},
	)

	cuenta.SetText(fmt.Sprintf("%d claves en el bucket", len(claves)))
	w.SetContent(container.NewBorder(cuenta, nil, nil, nil, lista))
	w.Resize(fyne.NewSize(640, 360))

	go func() {
		for i := 0; i < 40; i++ {
			fyne.Do(func() {
				cuenta.SetText(fmt.Sprintf("%d claves · filas construidas: %d",
					len(claves), construidas))
			})
			if i == 6 {
				break
			}
		}
	}()

	w.ShowAndRun()
}
