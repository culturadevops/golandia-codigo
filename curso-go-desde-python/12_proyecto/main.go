// Leccion 12 — El revisar_servicios del curso de Python, en Go: mismas cuatro cosas.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"
)

const (
	salidaOK          = 0
	salidaHayProblema = 1
	salidaErrorDeUso  = 2
)

type Servicio struct {
	Nombre string  `json:"nombre"`
	Estado string  `json:"estado"`
	CPU    float64 `json:"cpu"`
}

func (s Servicio) Sano() bool { return s.Estado == "activo" }

var inventario = []Servicio{
	{"web-01", "activo", 0.91},
	{"web-02", "parado", 0.00},
	{"db-01", "activo", 0.42},
	{"cache-01", "activo", 0.88},
}

func main() { os.Exit(run()) }

func run() int {
	umbral := flag.Float64("umbral", 0.8, "cpu por encima de la cual se avisa")
	comoJSON := flag.Bool("json", false, "salida en json, para otra herramienta")
	verboso := flag.Bool("v", false, "mensajes de depuracion")
	flag.Parse()

	nivel := slog.LevelInfo
	if *verboso {
		nivel = slog.LevelDebug
	}
	// el registro va a STDERR: asi `revisar -json > estado.json` deja el json limpio
	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: nivel}))

	if *umbral < 0 || *umbral > 1 {
		log.Error("el umbral tiene que estar entre 0 y 1", "umbral", *umbral)
		return salidaErrorDeUso
	}

	var caidos, saturados []Servicio
	for _, s := range inventario {
		switch {
		case !s.Sano():
			caidos = append(caidos, s)
		case s.CPU > *umbral:
			saturados = append(saturados, s)
		}
	}

	if *comoJSON {
		out, _ := json.MarshalIndent(map[string]any{
			"umbral": *umbral, "total": len(inventario),
			"caidos": caidos, "saturados": saturados,
		}, "", "  ")
		fmt.Println(string(out))
	} else {
		fmt.Printf("Revisados %d servicios (umbral %.0f%%)\n", len(inventario), *umbral*100)
		for _, s := range caidos {
			fmt.Printf("  CAIDO      %s\n", s.Nombre)
		}
		for _, s := range saturados {
			fmt.Printf("  SATURADO   %s  cpu %.0f%%\n", s.Nombre, s.CPU*100)
		}
	}
	if len(caidos)+len(saturados) > 0 {
		log.Warn("hay problemas", "caidos", len(caidos), "saturados", len(saturados))
		return salidaHayProblema
	}
	return salidaOK
}
