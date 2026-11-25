// Lección 1: por qué main.go solo llama a run().
//
//	go run ./lecciones/01_estructura -modo fatal   log.Fatal: los defer no corren
//	go run ./lecciones/01_estructura -modo run     run() devuelve el error: los defer sí
package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
)

func main() {
	modo := flag.String("modo", "run", "fatal o run")
	flag.Parse()
	log.SetFlags(0)
	if *modo == "fatal" {
		conFatal()
		return
	}
	if err := run(); err != nil {
		fmt.Println("main: error:", err)
		os.Exit(1)
	}
}

func abrirRecursos() func() {
	fmt.Println("  abro el pool de la base de datos")
	return func() { fmt.Println("  defer: cierro el pool y vacío los logs") }
}

func conFatal() {
	cerrar := abrirRecursos()
	defer cerrar()
	log.Fatal("  no puedo escuchar en :8080") // os.Exit(1) por dentro: el defer nunca llega
}

func run() error {
	cerrar := abrirRecursos()
	defer cerrar()
	return errors.New("no puedo escuchar en :8080")
}
