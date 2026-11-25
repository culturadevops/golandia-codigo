// Mis tareas: la primera app de escritorio en Go con Fyne. Una ventana, una caja de texto, un botón y una lista.
package main

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type ventana struct {
	w       fyne.Window
	entrada *widget.Entry
	agregar *widget.Button
	conteo  *widget.Label
	tareas  []string
}

func construir(a fyne.App) *ventana {
	v := &ventana{w: a.NewWindow("Mis tareas")}
	lista := widget.NewList(
		func() int { return len(v.tareas) },
		func() fyne.CanvasObject { return widget.NewLabel("plantilla") },
		func(i widget.ListItemID, o fyne.CanvasObject) { o.(*widget.Label).SetText("• " + v.tareas[i]) },
	)
	v.entrada = widget.NewEntry()
	v.entrada.SetPlaceHolder("Escribe una tarea…")
	v.conteo = widget.NewLabel("0 tareas")
	v.agregar = widget.NewButton("Agregar", func() {
		if v.entrada.Text == "" {
			return
		}
		v.tareas = append(v.tareas, v.entrada.Text)
		v.entrada.SetText("")
		lista.Refresh()
		v.conteo.SetText(fmt.Sprintf("%d tareas", len(v.tareas)))
	})
	arriba := container.NewBorder(nil, nil, nil, v.agregar, v.entrada)
	v.w.SetContent(container.NewBorder(arriba, v.conteo, nil, nil, lista))
	v.w.Resize(fyne.NewSize(420, 320))
	return v
}

func main() {
	construir(app.New()).w.ShowAndRun()
}
