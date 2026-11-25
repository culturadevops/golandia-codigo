// La MISMA app que app-fyne: una ventana con un contador, un botón que suma,
// y una lista de cinco elementos. Aquí la ventana es un navegador del sistema.
package main

import (
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	wails.Run(&options.App{
		Title:  "contador",
		Width:  400,
		Height: 300,
		Assets: assets,
		Bind:   []interface{}{&App{}},
	})
}
