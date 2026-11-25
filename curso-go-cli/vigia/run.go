package main

import (
	"context"
	"fmt"
)

const ayuda = `vigia revisa endpoints HTTP.

uso:
  vigia revisar [flags] URL...     (o "-" para leer las URLs de stdin)
  vigia config  [flags]            la configuración efectiva, y de dónde sale
  vigia version                    versión, commit y con qué se compiló

configuración (gana la primera que exista):
  1. flags          -timeout 3s  -concurrencia 8  -config archivo.json
  2. entorno        VIGIA_TIMEOUT  VIGIA_CONCURRENCIA  VIGIA_CONFIG
  3. archivo        vigia.json en el directorio actual, o el de -config / VIGIA_CONFIG
  4. por defecto    timeout 10s, concurrencia 4

códigos de salida:
  0 todo arriba · 1 hay endpoints caídos · 2 uso o configuración mal · 130 interrumpido
`

// run es main sin os.Exit: recibe los argumentos y el entorno, y devuelve el código.
// Así los tests llaman a run directamente, sin compilar ni lanzar un proceso.
func run(ctx context.Context, args []string, e Entorno) int {
	if len(args) == 0 {
		fmt.Fprint(e.Stderr, ayuda)
		return exitUso
	}
	switch args[0] {
	case "revisar":
		return revisar(ctx, args[1:], e)
	case "config":
		return mostrarConfig(args[1:], e)
	case "version", "-version", "--version":
		return mostrarVersion(e)
	case "ayuda", "-h", "-help", "--help":
		fmt.Fprint(e.Stdout, ayuda)
		return exitOK
	default:
		fmt.Fprintf(e.Stderr, "vigia: subcomando desconocido %q (prueba «vigia ayuda»)\n", args[0])
		return exitUso
	}
}
