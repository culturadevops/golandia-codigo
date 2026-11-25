// e01 · La primera ventana. Nada mas: para medir que cuesta.
package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/widget"
)

func main() {
	a := app.New()
	w := a.NewWindow("nubecita")
	w.SetContent(widget.NewLabel("Tu nube local, con ventana"))
	w.Resize(fyne.NewSize(640, 220))
	w.ShowAndRun()
}
