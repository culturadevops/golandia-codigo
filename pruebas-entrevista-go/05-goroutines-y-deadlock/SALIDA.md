# Pruebas ejecutadas — Episodio 5/8 — canales, deadlock y goroutines que no mueren

> Salida **literal** de esta máquina. La genera `probar.py` y se vuelve a generar sola.
> ⛔ Lo que no esté aquí no puede salir en pantalla (canon: ninguna cifra inventada,
> y lo que se muestra como código se ejecuta antes).

**Entorno citable en el vídeo:** contenedor `golang:1.26` · `go version go1.26.7 linux/amd64`

---

## 1. Cinco lineas y se cuelga

**La pregunta:** Por que muere este programa?

`p01/main.go`

```go
package main

import "fmt"

func main() {
	ch := make(chan int) // SIN buffer

	ch <- 1 // envio... y nadie esta esperando al otro lado

	fmt.Println("esto no se imprime nunca:", <-ch)
}
```

```console
$ go run ./p01   # ⚠️ falla A PROPÓSITO
fatal error: all goroutines are asleep - deadlock!

goroutine 1 [chan send]:
main.main()
	/app/p01/main.go:8 +0x36
exit status 2
[codigo de salida: 1]
```

> El runtime lo detecta porque TODAS las goroutines estan dormidas. Si una sola
> estuviera viva, no habria mensaje: se quedaria colgado en silencio. Eso hay que decirlo.

---

## 2. Dos arreglos, y no son lo mismo

**La pregunta:** Como lo arreglas? Y por que hay dos respuestas?

`p02/main.go`

```go
package main

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
```

```console
$ go run ./p02
con buffer: 1
con goroutine: 1

No son intercambiables: el buffer desacopla en el tiempo,
la goroutine mantiene la cita. Elegir mal el arreglo es la trampa.
[codigo de salida: 0]
```

---

## 3. El WaitGroup con el Add en el sitio equivocado

**La pregunta:** Donde va el Add(1)?

`p03/main.go`

```go
package main

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
```

```console
$ go run ./p03
con Add DENTRO -> esperaba 1000, hechas: 963
con Add FUERA   -> esperaba 1000, hechas: 1000
[codigo de salida: 0]
```

```console
$ go run -race ./p03
con Add DENTRO -> esperaba 1000, hechas: 977
con Add FUERA   -> esperaba 1000, hechas: 1000
[codigo de salida: 0]
```

---

## 4. Canal cerrado: recibir va, enviar no

**La pregunta:** Que pasa si recibes de un canal cerrado? Y si envias?

`p04/main.go`

```go
package main

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
```

```console
$ go run ./p04   # ⚠️ falla A PROPÓSITO
primero:  1 true
ya vacio: 0 false  <-- cero y false, NO bloquea
ahora envio a un canal cerrado...
panic: send on closed channel

goroutine 1 [running]:
main.main()
	/app/p04/main.go:18 +0x21b
exit status 2
[codigo de salida: 1]
```

> De aqui sale la regla: cierra SIEMPRE el que envia, nunca el que recibe.
> El que recibe no puede saber si queda alguien enviando.

---

## 5. La goroutine que nunca muere

**La pregunta:** El programa termina bien. Hay algun problema?

`p05/main.go`

```go
package main

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
```

```console
$ go run ./p05
goroutines al empezar: 1
goroutines tras 100 llamadas: 101

El programa acaba sin error y sin aviso. Nadie te dice nada.
En un servidor, esto crece hasta que el proceso se muere por memoria.
[codigo de salida: 0]
```
