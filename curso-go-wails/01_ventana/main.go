// Lección 1 · La primera ventana. La ventana es el navegador del sistema (en Linux, WebKitGTK);
// el HTML va metido en el binario con go:embed, y Go le presta métodos con Bind.
package main

import (
	"embed"
	"runtime"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var web embed.FS

type App struct{}

// Saludo lo llama el JavaScript de la ventana, y le llega como una promesa.
func (App) Saludo(nombre string) string {
	return "Hola, " + nombre + ". Te escribe Go " + runtime.Version() + " desde " + runtime.GOOS
}

func main() {
	wails.Run(&options.App{
		Title: "Lección 1", Width: 900, Height: 640,
		AssetServer: &assetserver.Options{Assets: web},
		Bind:        []interface{}{&App{}},
	})
}
