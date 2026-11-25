// 06_bucle_for.go — el bug del bucle `for` que Go arrastro diez anios y arreglo en 1.22.
//
// ⚠️ ESTE FICHERO HAY QUE CORRERLO DOS VECES, con golang:1.21 y con golang:1.22, y ensenar LAS
// DOS salidas. Con una sola no hay video: la gracia es que el MISMO codigo da resultados
// distintos segun la version, que es un lenguaje admitiendo que se equivoco.
//
//	docker run --rm -v "$(pwd -W):/app" -w /app golang:1.21-alpine go run 06_bucle_for.go
//	docker run --rm -v "$(pwd -W):/app" -w /app golang:1.22-alpine go run 06_bucle_for.go
//
// (En Git Bash hace falta MSYS_NO_PATHCONV=1 o rompe la ruta del volumen.)
//
// NOTA: no lleva `go.mod` propio a proposito — se ejecuta con `go run` de un fichero suelto para
// que la version del contenedor sea la unica que manda.
package main

import (
	"fmt"
	"sort"
	"sync"
)

func main() {
	fmt.Println("=== 1 · goroutines que capturan la variable del bucle ===")
	var wg sync.WaitGroup
	var mu sync.Mutex
	vistos := []int{}
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mu.Lock()
			vistos = append(vistos, i)
			mu.Unlock()
		}()
	}
	wg.Wait()
	sort.Ints(vistos)
	fmt.Println("  lo que vieron las 5 goroutines:", vistos)
	fmt.Println("  esperado (1.22+): [0 1 2 3 4]")
	fmt.Println("  en 1.21 y antes : algo como [5 5 5 5 5] — todas comparten LA MISMA variable")

	fmt.Println()
	fmt.Println("=== 2 · lo mismo sin concurrencia: punteros a la variable del bucle ===")
	punteros := []*int{}
	for j := 0; j < 3; j++ {
		punteros = append(punteros, &j)
	}
	fmt.Print("  los tres punteros valen: ")
	for _, p := range punteros {
		fmt.Print(*p, " ")
	}
	fmt.Println()
	fmt.Println("  esperado (1.22+): 0 1 2")
	fmt.Println("  en 1.21 y antes : 3 3 3 — los tres apuntan al MISMO sitio")

	fmt.Println()
	fmt.Println("=== 3 · y las direcciones lo demuestran ===")
	for k, p := range punteros {
		fmt.Printf("  puntero %d -> %p\n", k, p)
	}
	fmt.Println("  si las tres direcciones son iguales, es UNA sola variable compartida.")
}
