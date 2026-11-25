// e03 · Binding: el dato manda, y la interfaz se entera sola.
package main

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
)

func main() {
	a := app.New()
	w := a.NewWindow("nubecita · binding")

	objetos := binding.NewInt()
	objetos.Set(2)

	// La etiqueta NO se actualiza a mano en ningun sitio: escucha al dato.
	texto := binding.IntToStringWithFormat(objetos, "el bucket tiene %d objetos")
	etiqueta := widget.NewLabelWithData(texto)

	subir := widget.NewButton("Subir un objeto", func() {
		n, _ := objetos.Get()
		objetos.Set(n + 1) // solo se toca el DATO
	})
	borrar := widget.NewButton("Borrar todo", func() { objetos.Set(0) })

	w.SetContent(container.NewVBox(
		widget.NewLabel(fmt.Sprintf("Bucket: fotos")),
		etiqueta, subir, borrar,
	))
	w.Resize(fyne.NewSize(640, 260))
	w.ShowAndRun()
}
