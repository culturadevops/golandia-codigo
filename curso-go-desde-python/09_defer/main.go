// Leccion 9 — defer: el `finally` y el `with` de Go. Y cuando se acumula.
package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func paso(n int) {
	fmt.Printf("  empieza %d\n", n)
	defer fmt.Printf("  defer de %d (al SALIR de la funcion)\n", n)
	fmt.Printf("  termina %d\n", n)
}

func main() {
	fmt.Println("== 9.1 · defer corre al salir de la FUNCION ==")
	paso(1)

	fmt.Println()
	fmt.Println("== 9.2 · Varios defer: al reves de como se escriben ==")
	func() {
		for i := 1; i <= 3; i++ {
			defer fmt.Printf("  defer %d\n", i)
		}
		fmt.Println("  (fin del bucle)")
	}()

	fmt.Println()
	fmt.Println("== 9.3 · LA TRAMPA: defer dentro de un bucle no cierra en cada vuelta ==")
	dir, _ := os.MkdirTemp("", "curso-go-")
	defer os.RemoveAll(dir)
	abiertos := 0
	func() {
		for i := 0; i < 5; i++ {
			f, err := os.Create(filepath.Join(dir, fmt.Sprintf("f%d.txt", i)))
			if err != nil {
				fmt.Println("  error:", err)
				return
			}
			defer f.Close() // ⛔ no se cierra hasta que acaba ESTA funcion
			abiertos++
		}
		fmt.Printf("  al acabar el bucle siguen abiertos: %d ficheros\n", abiertos)
	}()
	fmt.Println("  Con 5 da igual. Con 50.000 lineas de log, te quedas sin descriptores.")
	fmt.Println("  En Python, `with open(...)` cierra al salir del BLOQUE. Aqui, de la funcion.")
}
