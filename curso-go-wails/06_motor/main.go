// Lección 6 · El motor es del sistema. Wails no lleva navegador: usa el del sistema operativo.
// Por eso el binario es pequeño, y por eso el mismo HTML no se ve igual en todas partes.
// Esta ventana le pregunta a SU motor quién es y qué sabe hacer.
package main

import (
	"embed"

	"curso-wails/comun"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var web embed.FS

type App struct{ comun.Chivato }

func main() {
	wails.Run(&options.App{
		Title: "Lección 6", Width: 900, Height: 640,
		AssetServer: &assetserver.Options{Assets: web},
		Bind:        []interface{}{&App{}},
	})
}
