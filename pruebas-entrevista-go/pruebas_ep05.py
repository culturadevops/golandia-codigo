# -*- coding: utf-8 -*-
import sys
from pathlib import Path
sys.path.insert(0, str(Path(__file__).resolve().parent))
from probar import episodio

P1 = '''package main

import "fmt"

func main() {
	ch := make(chan int) // SIN buffer

	ch <- 1 // envio... y nadie esta esperando al otro lado

	fmt.Println("esto no se imprime nunca:", <-ch)
}
'''

P2 = '''package main

import "fmt"

func main() {
	// ARREGLO A: darle buzon. Cambia el comportamiento: ya no hay cita.
	a := make(chan int, 1)
	a <- 1
	fmt.Println("con buffer:", <-a)

	// ARREGLO B: que haya alguien al otro lado. Es el arreglo BUENO en el 99% de los casos.
	b := make(chan int)
	go func() { b <- 1 }()
	fmt.Println("con goroutine:", <-b)

	fmt.Println()
	fmt.Println("No son intercambiables: el buffer desacopla en el tiempo,")
	fmt.Println("la goroutine mantiene la cita. Elegir mal el arreglo es la trampa.")
}
'''

P3 = '''package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

func main() {
	// EL BUG: el Add(1) esta DENTRO de la goroutine.
	var wg sync.WaitGroup
	var hechas atomic.Int64
	for i := 0; i < 1000; i++ {
		go func() {
			wg.Add(1) // <-- tarde: Wait() puede haber vuelto ya
			defer wg.Done()
			hechas.Add(1)
		}()
	}
	wg.Wait()
	fmt.Println("con Add DENTRO -> esperaba 1000, hechas:", hechas.Load())

	// BIEN: Add antes de lanzar.
	var wg2 sync.WaitGroup
	var hechas2 atomic.Int64
	for i := 0; i < 1000; i++ {
		wg2.Add(1)
		go func() {
			defer wg2.Done()
			hechas2.Add(1)
		}()
	}
	wg2.Wait()
	fmt.Println("con Add FUERA   -> esperaba 1000, hechas:", hechas2.Load())
}
'''

P4 = '''package main

import "fmt"

func main() {
	ch := make(chan int, 2)
	ch <- 1
	close(ch)

	// Recibir de un canal cerrado va bien: da el valor cero y ok=false.
	v, ok := <-ch
	fmt.Println("primero: ", v, ok)
	v, ok = <-ch
	fmt.Println("ya vacio:", v, ok, " <-- cero y false, NO bloquea")

	// Enviar a un canal cerrado, en cambio:
	fmt.Println("ahora envio a un canal cerrado...")
	ch <- 3
}
'''

P5 = '''package main

import (
	"fmt"
	"runtime"
	"time"
)

// La fuga: la goroutine se queda esperando para siempre en un canal que nadie leera.
func fuga() {
	ch := make(chan int)
	go func() {
		ch <- 42 // nadie va a recibir: esta goroutine no muere jamas
	}()
	// nos vamos sin leer
}

func main() {
	fmt.Println("goroutines al empezar:", runtime.NumGoroutine())
	for i := 0; i < 100; i++ {
		fuga()
	}
	time.Sleep(100 * time.Millisecond)
	fmt.Println("goroutines tras 100 llamadas:", runtime.NumGoroutine())
	fmt.Println()
	fmt.Println("El programa acaba sin error y sin aviso. Nadie te dice nada.")
	fmt.Println("En un servidor, esto crece hasta que el proceso se muere por memoria.")
}
'''

episodio(
    "05-goroutines-y-deadlock",
    "Episodio 5/8 — canales, deadlock y goroutines que no mueren",
    [
        {"nombre": "1. Cinco lineas y se cuelga",
         "pregunta": "Por que muere este programa?",
         "archivos": {"p01/main.go": P1},
         "correr": [(["go", "run", "./p01"], True)],
         "nota": "El runtime lo detecta porque TODAS las goroutines estan dormidas. Si una sola\nestuviera viva, no habria mensaje: se quedaria colgado en silencio. Eso hay que decirlo."},
        {"nombre": "2. Dos arreglos, y no son lo mismo",
         "pregunta": "Como lo arreglas? Y por que hay dos respuestas?",
         "archivos": {"p02/main.go": P2},
         "correr": [(["go", "run", "./p02"], False)]},
        {"nombre": "3. El WaitGroup con el Add en el sitio equivocado",
         "pregunta": "Donde va el Add(1)?",
         "archivos": {"p03/main.go": P3},
         "correr": [(["go", "run", "./p03"], False), (["go", "run", "-race", "./p03"], False)]},
        {"nombre": "4. Canal cerrado: recibir va, enviar no",
         "pregunta": "Que pasa si recibes de un canal cerrado? Y si envias?",
         "archivos": {"p04/main.go": P4},
         "correr": [(["go", "run", "./p04"], True)],
         "nota": "De aqui sale la regla: cierra SIEMPRE el que envia, nunca el que recibe.\nEl que recibe no puede saber si queda alguien enviando."},
        {"nombre": "5. La goroutine que nunca muere",
         "pregunta": "El programa termina bien. Hay algun problema?",
         "archivos": {"p05/main.go": P5},
         "correr": [(["go", "run", "./p05"], False)]},
    ])
