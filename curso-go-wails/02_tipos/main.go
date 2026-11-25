// Lección 2 · Los tipos, por el puente. Lo que devuelve un método de Go viaja como JSON hasta el
// JavaScript de la ventana. Nada da error. Pero no todo llega igual.
package main

import (
	"embed"
	"fmt"
	"time"

	"curso-wails/comun"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var web embed.FS

type Pedido struct {
	ID       int64            // un identificador de 64 bits, como los de una base de datos
	Lineas   []string         // un pedido recién creado: todavía sin líneas
	Creado   time.Time
	PorPais  map[int]string   // claves enteras
	interno  string           // minúscula: no se exporta
}

type App struct{ comun.Chivato }

func (App) Pedido() Pedido {
	return Pedido{
		ID:      9007199254740993,                                 // 2^53 + 1
		Lineas:  nil,                                              // var lineas []string
		Creado:  time.Date(2026, 10, 5, 12, 0, 0, 0, time.FixedZone("CDMX", -6*3600)),
		PorPais: map[int]string{52: "México", 54: "Argentina"},
		interno: "secreto",
	}
}

// PedidoBien: el mismo pedido, preparado para cruzar el puente.
type PedidoBien struct {
	ID     int64    `json:"ID,string"` // viaja como texto: JavaScript no lo redondea
	Lineas []string `json:"Lineas"`    // nunca nil: se inicializa vacía
	Creado time.Time
}

func (a App) PedidoBien() PedidoBien {
	p := a.Pedido()
	lineas := p.Lineas
	if lineas == nil {
		lineas = []string{}
	}
	return PedidoBien{ID: p.ID, Lineas: lineas, Creado: p.Creado}
}

// ComoLoVeGo: lo mismo, contado desde Go, para compararlo en la ventana.
func (a App) ComoLoVeGo() []string {
	p := a.Pedido()
	return []string{
		fmt.Sprintf("ID: %d", p.ID),
		fmt.Sprintf("Lineas: %#v (len %d)", p.Lineas, len(p.Lineas)),
		fmt.Sprintf("Creado: %s", p.Creado),
		fmt.Sprintf("PorPais: %v", p.PorPais),
		fmt.Sprintf("interno: %q", p.interno),
	}
}

func main() {
	wails.Run(&options.App{
		Title: "Lección 2", Width: 900, Height: 640,
		AssetServer: &assetserver.Options{Assets: web},
		Bind:        []interface{}{&App{}},
	})
}
