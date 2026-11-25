// e08 · Tema propio y recursos EMPOTRADOS.
// El icono viaja DENTRO del binario con go:embed: si estuviera en disco y copiaras solo el
// ejecutable, la aplicacion arrancaria sin icono y nadie entenderia por que.
package main

import (
	_ "embed"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

//go:embed icono.png
var iconoPNG []byte // <- el fichero entra en el BINARIO, no se lee de disco

// Un tema es una interfaz de cuatro metodos. Aqui solo se cambia un color y se delega el resto.
type temaGo struct{ fyne.Theme }

func (t temaGo) Color(n fyne.ThemeColorName, v fyne.ThemeVariant) color.Color {
	if n == theme.ColorNamePrimary {
		return color.NRGBA{R: 0x00, G: 0xAD, B: 0xD8, A: 255} // el cian de Go
	}
	return t.Theme.Color(n, v)
}

func main() {
	a := app.New()
	a.Settings().SetTheme(temaGo{Theme: theme.DefaultTheme()})

	icono := fyne.NewStaticResource("icono.png", iconoPNG)
	a.SetIcon(icono)

	w := a.NewWindow("nubecita · tema propio")
	w.SetIcon(icono)
	w.SetContent(container.NewVBox(
		widget.NewLabelWithStyle("Tu nube local", fyne.TextAlignLeading,
			fyne.TextStyle{Bold: true}),
		widget.NewIcon(icono),
		widget.NewButton("Crear bucket", func() {}),
		widget.NewCheck("Mostrar objetos ocultos", func(bool) {}),
		widget.NewProgressBar(),
	))
	w.Resize(fyne.NewSize(640, 300))
	w.ShowAndRun()
}
