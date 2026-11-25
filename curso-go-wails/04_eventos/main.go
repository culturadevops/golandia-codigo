// Lección 4 · Eventos de Go a la ventana. Un trabajo largo avisa de su progreso con
// runtime.EventsEmit. La trampa: empezar a emitir en OnStartup, antes de que la ventana haya
// cargado y se haya suscrito. No da error: los eventos se pierden.
package main

import (
	"context"
	"embed"
	"os"
	"time"

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

func (a *App) trabajo(etiqueta string) {
	for i := 1; i <= 10; i++ {
		runtime.EventsEmit(a.ctx, "progreso", etiqueta, i)
		time.Sleep(100 * time.Millisecond)
	}
	runtime.EventsEmit(a.ctx, "fin", etiqueta)
}

func (a *App) arrancar(ctx context.Context) {
	a.ctx = ctx
	if os.Getenv("MODO") != "bueno" {
		go a.trabajo("en OnStartup") // ⛔ la ventana aún no escucha
	}
}

// Listo lo llama la ventana cuando ya se ha suscrito: entonces, y solo entonces, se empieza.
func (a *App) Listo() {
	if os.Getenv("MODO") == "bueno" {
		go a.trabajo("tras Listo()")
	}
}

func main() {
	a := &App{}
	wails.Run(&options.App{
		Title: "Lección 4", Width: 900, Height: 640,
		AssetServer: &assetserver.Options{Assets: web},
		OnStartup:   a.arrancar,
		Bind:        []interface{}{a},
	})
}
