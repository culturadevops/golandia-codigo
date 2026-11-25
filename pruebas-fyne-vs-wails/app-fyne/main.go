// La MISMA app que app-wails: una ventana con un contador, un botón que suma,
// y una lista de cinco elementos. Ni una función de más en ninguno de los dos.
package main

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func main() {
	a := app.New()
	w := a.NewWindow("contador")

	n := 0
	eti := widget.NewLabel("0")
	bot := widget.NewButton("sumar", func() {
		n++
		eti.SetText(fmt.Sprintf("%d", n))
	})

	items := []string{"uno", "dos", "tres", "cuatro", "cinco"}
	lista := widget.NewList(
		func() int { return len(items) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(items[i])
		},
	)

	w.SetContent(container.NewBorder(container.NewVBox(eti, bot), nil, nil, nil, lista))
	w.ShowAndRun()
}
