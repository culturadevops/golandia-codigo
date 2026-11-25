// e02 · Contenedores: quien decide el tamano NO es el widget, es el contenedor.
package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func main() {
	a := app.New()
	w := a.NewWindow("nubecita · contenedores")

	arriba := widget.NewLabel("  Buckets")
	abajo := widget.NewLabel("  4566 · conectado")
	izquierda := container.NewVBox(
		widget.NewButton("S3", func() {}),
		widget.NewButton("Colas", func() {}),
		widget.NewButton("Funciones", func() {}),
	)
	centro := container.NewGridWithColumns(2,
		widget.NewLabel("fotos"), widget.NewLabel("2 objetos"),
		widget.NewLabel("copias"), widget.NewLabel("0 objetos"),
		widget.NewLabel("informes"), widget.NewLabel("17 objetos"),
	)

	w.SetContent(container.NewBorder(arriba, abajo, izquierda, nil, centro))
	w.Resize(fyne.NewSize(640, 300))
	w.ShowAndRun()
}
