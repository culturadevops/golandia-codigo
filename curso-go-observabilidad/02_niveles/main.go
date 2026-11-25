// Lección 2 · Niveles: un Debug que no imprime nada sigue costando lo que cuestan sus argumentos.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"curso-obs/comun"
)

// El estado del carrito: algo que cuesta serializar.
type Carrito struct {
	Lineas []Linea
}

type Linea struct {
	Sku      string
	Cantidad int
	Precio   int64
}

var serializados int

func (c Carrito) JSON() string {
	serializados++
	b, _ := json.Marshal(c)
	return string(b)
}

// carritoPerezoso solo se serializa si el registro se va a escribir de verdad.
type carritoPerezoso struct{ c Carrito }

func (p carritoPerezoso) LogValue() slog.Value { return slog.StringValue(p.c.JSON()) }

const veces = 10_000

// contador cuenta las líneas que el handler escribe de verdad.
type contador struct{ lineas int }

func (c *contador) Write(p []byte) (int, error) { c.lineas++; return len(p), nil }

func main() {
	carrito := Carrito{}
	for i := 0; i < 200; i++ {
		carrito.Lineas = append(carrito.Lineas, Linea{Sku: fmt.Sprintf("SKU-%04d", i), Cantidad: i % 5, Precio: 1990})
	}
	salida := &contador{}
	log := slog.New(slog.NewJSONHandler(salida, &slog.HandlerOptions{Level: slog.LevelInfo}))
	ctx := context.Background()

	medir := func(titulo string, f func()) {
		serializados, salida.lineas = 0, 0
		t0 := time.Now()
		for i := 0; i < veces; i++ {
			f()
		}
		fmt.Printf("  %-28s %6s · carrito serializado %5d veces · líneas escritas: %d\n", titulo, comun.Ms(time.Since(t0)), serializados, salida.lineas)
	}

	comun.Titulo("2.1 · La trampa: 10.000 Debug con el nivel en Info")
	medir("Debug con c.JSON():", func() { log.Debug("carrito", "estado", carrito.JSON()) })
	fmt.Println("  no se escribió nada, y el carrito se serializó igual: los argumentos se evalúan antes de llamar")

	comun.Titulo("2.2 · Lo correcto (1): preguntar antes con Enabled")
	medir("if log.Enabled(Debug):", func() {
		if log.Enabled(ctx, slog.LevelDebug) {
			log.Debug("carrito", "estado", carrito.JSON())
		}
	})

	comun.Titulo("2.3 · Lo correcto (2): un LogValuer, que se evalúa solo si se escribe")
	medir("Debug con LogValuer:", func() { log.Debug("carrito", "estado", carritoPerezoso{carrito}) })
}
