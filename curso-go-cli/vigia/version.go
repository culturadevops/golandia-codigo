package main

import (
	"fmt"
	"runtime/debug"
)

// Las rellena el enlazador: go build -ldflags "-X main.version=1.2.3 -X main.commit=abc1234".
// Tienen que ser variables string de nivel de paquete. Con una constante, o con un nombre
// mal escrito, -X no dice nada y aquí se queda "dev".
var (
	version = "dev"
	commit  = ""
	fecha   = ""
)

func mostrarVersion(e Entorno) int {
	fmt.Fprintf(e.Stdout, "vigia %s\n", version)
	if commit != "" {
		fmt.Fprintf(e.Stdout, "  %-13s %s (%s)\n", "commit", commit, fecha)
	}
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return exitOK
	}
	fmt.Fprintf(e.Stdout, "  %-13s %s\n", "go", info.GoVersion)
	fmt.Fprintf(e.Stdout, "  %-13s %s %s\n", "módulo", info.Main.Path, info.Main.Version)
	for _, s := range info.Settings {
		switch s.Key {
		case "vcs.revision", "vcs.time", "vcs.modified", "GOOS", "GOARCH", "-trimpath", "CGO_ENABLED":
			fmt.Fprintf(e.Stdout, "  %-13s %s\n", s.Key, s.Value)
		}
	}
	return exitOK
}
