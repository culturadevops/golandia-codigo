// Leccion 1 — De script a binario: lo primero que cambia al venir de Python.
package main

import (
	"fmt"
	"os"
	"runtime"
)

func main() {
	fmt.Println("== 1.1 · Esto ya esta compilado ==")
	fmt.Printf("  go         %s\n", runtime.Version())
	fmt.Printf("  sistema    %s/%s\n", runtime.GOOS, runtime.GOARCH)
	exe, _ := os.Executable()
	fmt.Printf("  binario    %s\n", exe)
	fmt.Println()
	fmt.Println("  En Python, `python script.py` necesita un interprete instalado,")
	fmt.Println("  la version correcta y las dependencias. En Go, `go build` deja UN")
	fmt.Println("  fichero que se copia a la maquina y se ejecuta. Nada mas.")
	fmt.Println()
	fmt.Println("== 1.2 · Compilar para otra maquina sin salir de la tuya ==")
	fmt.Println("  GOOS=linux GOARCH=amd64 go build -o revisar .")
	fmt.Println("  El tamano y el arranque medidos estan en la leccion 12.")
}
