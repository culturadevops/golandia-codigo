// Videos 204 (Go no tiene excepciones) y 208 (panic en una libreria).
//
// LO QUE DEMUESTRA, y es lo que desactiva el malentendido de siempre: panic/recover NO son
// try/catch. Tienen tres limites que una excepcion de verdad no tiene:
//
//	1. recover() fuera de un defer devuelve nil: no captura nada
//	2. no puedes "recuperar" en mitad de una funcion y seguir donde estabas
//	3. un panic en OTRA goroutine se lleva el programa entero, aunque tengas recover en main
//
// El tercero es el que decide: en cualquier lenguaje con excepciones, un hilo que revienta no
// tumba el proceso. Aqui si. Por eso no son excepciones.
//
// El limite 3 MATA el programa, asi que va al final y con aviso.
//
//	go run .
package main

import (
	"fmt"
	"os"
	"time"
)

// Una "libreria" maleducada: hace panic en vez de devolver un error (video 208).
func libreriaMaleducada(n int) int {
	if n < 0 {
		panic("libreria: no acepto negativos")
	}
	return n * 2
}

// La misma, educada: devuelve un error, como manda Effective Go.
func libreriaEducada(n int) (int, error) {
	if n < 0 {
		return 0, fmt.Errorf("no acepto negativos, me diste %d", n)
	}
	return n * 2, nil
}

func main() {
	fmt.Println("=== 1 · recover() FUERA de un defer no captura nada ===")
	r := recover()
	fmt.Printf("  recover() llamado directamente -> %v  (nil: no ha capturado nada)\n", r)

	fmt.Println()
	fmt.Println("=== 2 · una libreria que hace panic se lleva a QUIEN LA USA ===")
	func() {
		defer func() {
			if p := recover(); p != nil {
				fmt.Println("  el que la usa tuvo que poner un recover para sobrevivir:")
				fmt.Println("    recuperado:", p)
			}
		}()
		fmt.Println("  llamo a libreriaMaleducada(-1)...")
		_ = libreriaMaleducada(-1)
		fmt.Println("  ESTA LINEA NO SE EJECUTA")
	}()
	fmt.Println("  -> y fijate: no se puede seguir donde estabas. La funcion entera se abandona.")

	fmt.Println()
	fmt.Println("=== 3 · la version educada: un error normal, y el flujo sigue ===")
	if _, err := libreriaEducada(-1); err != nil {
		fmt.Println("  error:", err)
		fmt.Println("  -> aqui SI sigo donde estaba, decido, y continuo")
	}

	fmt.Println()
	fmt.Println("=== 4 · el limite que lo decide: panic en OTRA goroutine ===")
	fmt.Println("  main tiene un recover puesto. Vamos a ver si sirve de algo.")
	defer func() {
		fmt.Println("  ESTE recover de main NO se va a ejecutar:", recover())
	}()
	go func() {
		panic("panic desde otra goroutine")
	}()
	time.Sleep(300 * time.Millisecond)
	fmt.Println("  si lees esto, el panic no se propago (no deberia pasar)")
	os.Exit(0)
}
