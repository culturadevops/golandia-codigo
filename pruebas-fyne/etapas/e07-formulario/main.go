// e07 · Formulario con validacion: el boton se apaga solo mientras el dato este mal.
package main

import (
	"errors"
	"regexp"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// Los nombres de bucket de S3: minusculas, numeros y guiones, de 3 a 63.
var valido = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{1,61}[a-z0-9]$`)

func main() {
	a := app.New()
	w := a.NewWindow("nubecita · crear bucket")

	nombre := widget.NewEntry()
	nombre.SetPlaceHolder("mi-bucket")
	nombre.Validator = func(s string) error {
		if !valido.MatchString(s) {
			return errors.New("minusculas, numeros y guiones; de 3 a 63")
		}
		return nil
	}
	nombre.SetText("Fotos_2026") // a proposito: MAL

	form := widget.NewForm(widget.NewFormItem("Nombre", nombre))
	form.SubmitText = "Crear"
	form.OnSubmit = func() {}

	w.SetContent(container.NewVBox(widget.NewLabel("Crear un bucket"), form))
	w.Resize(fyne.NewSize(640, 260))
	w.ShowAndRun()
}
