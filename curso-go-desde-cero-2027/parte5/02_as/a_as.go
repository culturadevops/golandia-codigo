package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
)

func leerConfig(ruta string) ([]byte, error) {
	datos, err := os.ReadFile(ruta)
	if err != nil {
		return nil, fmt.Errorf("leer configuración: %w", err)
	}
	return datos, nil
}

func main() {
	_, err := leerConfig("/etc/no-existe.toml")
	fmt.Println("error:", err)
	fmt.Println("¿es fs.ErrNotExist?", errors.Is(err, fs.ErrNotExist))
	var pe *fs.PathError
	if errors.As(err, &pe) {
		fmt.Printf("errors.As: operación=%q ruta=%q causa=%v\n", pe.Op, pe.Path, pe.Err)
	}
	datos, err := leerConfig("/etc/hostname")
	fmt.Println("con un archivo que sí existe:", err == nil, len(datos) > 0)
}
