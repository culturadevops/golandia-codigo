// 07_defer_tags_any_fugas.go — evidencia de cuatro piezas de la serie, en una corrida.
//
//	bloque 1 -> video 214  `defer` dentro de un bucle
//	bloque 2 -> video 213  los struct tags son cadenas magicas
//	bloque 3 -> video 212  `any` mata el tipado
//	bloque 4 -> video 209  goroutines que nunca mueren
//
// REGLA DE LA SERIE: ninguna se produce sin ejecutar antes lo que va a salir en pantalla. Esto
// es ese paso. La salida literal se pega en SALIDA.md.
//
//	go run 07_defer_tags_any_fugas.go
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

// ─────────────────────────────── 1 · defer dentro de un bucle (video 214) ──

func conDeferEnBucle(dir string, n int) int {
	abiertos := 0
	for i := 0; i < n; i++ {
		f, err := os.Open(filepath.Join(dir, fmt.Sprintf("f%d.txt", i)))
		if err != nil {
			continue
		}
		defer f.Close() // <- NO se ejecuta al final de la vuelta: al final de la FUNCION
		abiertos++
	}
	// aqui siguen los n ficheros abiertos, todos a la vez
	return abiertos
}

func sinDeferEnBucle(dir string, n int) int {
	max := 0
	for i := 0; i < n; i++ {
		func() { // el cuerpo del bucle, en su propia funcion
			f, err := os.Open(filepath.Join(dir, fmt.Sprintf("f%d.txt", i)))
			if err != nil {
				return
			}
			defer f.Close() // ahora SI cierra al acabar cada vuelta
			max = 1
		}()
	}
	return max
}

// ────────────────────────────────── 2 · struct tags magicos (video 213) ──

type BienEscrito struct {
	NombreCompleto string `json:"nombre_completo"`
	Edad           int    `json:"edad"`
}

type MalEscrito struct {
	NombreCompleto string `json:"nombre_completo`  // falta la comilla de cierre
	Edad           int    `json: "edad"`           // espacio despues de los dos puntos
}

// ⚠️ EL CASO QUE `go vet` NO PUEDE CAZAR, y es el que de verdad te muerde: la etiqueta es
// sintacticamente PERFECTA, solo que el nombre esta mal escrito. Ninguna herramienta sabe como
// querias llamar tu campo. Sale un JSON valido con la clave equivocada, y el consumidor recibe
// un campo vacio sin que nadie diga nada.
type ErrataInvisible struct {
	NombreCompleto string `json:"nombre_complet"` // falta la "o" final
	Edad           int    `json:"edda"`           // "edad" con las letras cambiadas
}

// ────────────────────────────────────── 3 · any mata el tipado (video 212) ──

func devuelveAny() any { return "esto es un texto, no un numero" }

// ──────────────────────────────── 4 · goroutines que nunca mueren (video 209) ──

func fuga() {
	ch := make(chan int) // sin buffer y SIN receptor
	go func() {
		ch <- 1 // se queda aqui para siempre
	}()
}

func main() {
	// ── 1
	fmt.Println("=== 1 · defer dentro de un bucle (video 214) ===")
	dir, _ := os.MkdirTemp("", "deferbucle")
	defer os.RemoveAll(dir)
	const N = 200
	for i := 0; i < N; i++ {
		os.WriteFile(filepath.Join(dir, fmt.Sprintf("f%d.txt", i)), []byte("x"), 0o644)
	}
	fmt.Printf("  con defer en el bucle : %d ficheros abiertos A LA VEZ\n", conDeferEnBucle(dir, N))
	fmt.Printf("  sacando el cuerpo     : %d fichero abierto a la vez\n", sinDeferEnBucle(dir, N))
	fmt.Println("  -> el defer del bucle NO cierra al acabar la vuelta: espera al final de la funcion")

	// ── 2
	fmt.Println()
	fmt.Println("=== 2 · los struct tags son cadenas magicas (video 213) ===")
	b, _ := json.Marshal(BienEscrito{"Ana Lopez", 33})
	m, _ := json.Marshal(MalEscrito{"Ana Lopez", 33})
	e, _ := json.Marshal(ErrataInvisible{"Ana Lopez", 33})
	fmt.Println("  etiqueta bien escrita :", string(b))
	fmt.Println("  sintaxis rota         :", string(m), "  <- go vet SI la caza")
	fmt.Println("  errata en el nombre   :", string(e), "  <- go vet NO puede cazarla")
	fmt.Println("  -> ni error, ni panic, ni aviso. Solo un JSON distinto del que esperabas.")

	// ── 3
	fmt.Println()
	fmt.Println("=== 3 · any mata el tipado (video 212) ===")
	v := devuelveAny()
	if n, vale := v.(int); vale {
		fmt.Println("  era un entero:", n)
	} else {
		fmt.Println("  la asercion a int NO vale. Y esto COMPILA sin una sola queja.")
	}
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("  y la version sin comprobar revienta en EJECUCION:")
			fmt.Println("    panic:", r)
			seccion4()
		}
	}()
	_ = v.(int) // <- esto compila y explota aqui
}

func seccion4() {
	fmt.Println()
	fmt.Println("=== 4 · goroutines que nunca mueren (video 209) ===")
	fmt.Printf("  al empezar            : %d goroutines\n", runtime.NumGoroutine())
	for i := 0; i < 1000; i++ {
		fuga()
	}
	time.Sleep(200 * time.Millisecond)
	fmt.Printf("  tras 1.000 llamadas   : %d goroutines\n", runtime.NumGoroutine())
	runtime.GC()
	time.Sleep(200 * time.Millisecond)
	fmt.Printf("  y despues de un GC    : %d goroutines\n", runtime.NumGoroutine())
	fmt.Println("  -> el recolector NO se las lleva: no son basura, estan ESPERANDO.")
}
