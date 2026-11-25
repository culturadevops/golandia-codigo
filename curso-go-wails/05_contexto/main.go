// Lección 5 · El contexto de Wails. Las funciones de runtime (eventos, diálogos, la ventana)
// necesitan el contexto que Wails pasa a OnStartup. ¿Qué pasa si se usa otro?
package main

import (
	"context"
	"embed"
	"fmt"

	"curso-wails/comun"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed all:frontend/dist
var web embed.FS

type App struct {
	comun.Chivato
	ctx context.Context
}

func (a *App) Titulo(bueno bool) (r string) {
	defer func() {
		if p := recover(); p != nil {
			r = fmt.Sprint("pánico: ", p)
		}
	}()
	ctx := context.Background() // ⛔ un contexto cualquiera
	if bueno {
		ctx = a.ctx // ✅ el que dio Wails en OnStartup
	}
	runtime.WindowSetTitle(ctx, "título cambiado desde Go")
	return "cambiado"
}

func main() {
	a := &App{}
	wails.Run(&options.App{
		Title: "Lección 5", Width: 900, Height: 640,
		AssetServer: &assetserver.Options{Assets: web},
		OnStartup:   func(ctx context.Context) { a.ctx = ctx },
		Bind:        []interface{}{a},
	})
}
