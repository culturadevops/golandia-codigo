// Lección 3 · Los errores, por el puente. Un método que devuelve (T, error) se convierte en una
// promesa: si el error no es nil, la promesa se rechaza. ¿Y qué le llega a JavaScript del error?
// ¿Y si el método entra en pánico?
package main

import (
	"embed"
	"errors"
	"fmt"
	"os"

	"curso-wails/comun"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var web embed.FS

var ErrSinStock = errors.New("sin stock")

type App struct{ comun.Chivato }

func (App) Comprar(producto string, cantidad int) (int, error) {
	if cantidad > 3 {
		return 0, fmt.Errorf("comprar %q: %w (quedan 3)", producto, ErrSinStock)
	}
	return 3 - cantidad, nil
}

func (App) Romper() string {
	var m map[string]int
	m["x"] = 1 // ⛔ pánico: asignar en un mapa nil
	return "no llega"
}

func main() {
	err := wails.Run(&options.App{
		Title: "Lección 3", Width: 900, Height: 640,
		AssetServer: &assetserver.Options{Assets: web},
		Bind:        []interface{}{&App{}},
	})
	fmt.Fprintln(os.Stdout, "wails.Run terminó:", err)
}
