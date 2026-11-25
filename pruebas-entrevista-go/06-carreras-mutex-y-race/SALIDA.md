# Pruebas ejecutadas — Episodio 6/8 — carreras de datos, mutex, atomic y -race

> Salida **literal** de esta máquina. La genera `probar.py` y se vuelve a generar sola.
> ⛔ Lo que no esté aquí no puede salir en pantalla (canon: ninguna cifra inventada,
> y lo que se muestra como código se ejecuta antes).

**Entorno citable en el vídeo:** contenedor `golang:1.26` · `go version go1.26.7 linux/amd64`

---

## 1. Dos goroutines sumando: cuanto sale? (tres corridas)

**La pregunta:** Dos goroutines suman 100.000 cada una. Cuanto vale el contador?

`p01/main.go`

```go
package main

import (
	"fmt"
	"sync"
)

func main() {
	contador := 0
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100000; j++ {
				contador++ // dos goroutines, la misma variable
			}
		}()
	}
	wg.Wait()
	fmt.Println("esperaba 200000, salio:", contador)
}
```

```console
$ go run ./p01
esperaba 200000, salio: 200000
[codigo de salida: 0]
```

```console
$ go run ./p01
esperaba 200000, salio: 200000
[codigo de salida: 0]
```

```console
$ go run ./p01
esperaba 200000, salio: 175681
[codigo de salida: 0]
```

> Tres corridas, tres numeros. Lo peligroso no es que falle: es que A VECES da el
> numero bueno, y entonces el bug se va a produccion.

---

## 2. El detector de carreras lo senala con el dedo

**La pregunta:** Como demuestras que hay una carrera?

```console
$ go run -race ./p01   # ⚠️ falla A PROPÓSITO
esperaba 200000, salio: 200000
==================
WARNING: DATA RACE
Read at 0x00c00011e028 by goroutine 8:
  main.main.func1()
      /app/p01/main.go:16 +0x99

Previous write at 0x00c00011e028 by goroutine 7:
  main.main.func1()
      /app/p01/main.go:16 +0xab

Goroutine 8 (running) created at:
  main.main()
      /app/p01/main.go:13 +0x78

Goroutine 7 (running) created at:
  main.main()
      /app/p01/main.go:13 +0x78
==================
==================
WARNING: DATA RACE
Write at 0x00c00011e028 by goroutine 8:
  main.main.func1()
      /app/p01/main.go:16 +0xab

Previous write at 0x00c00011e028 by goroutine 7:
  main.main.func1()
      /app/p01/main.go:16 +0xab

Goroutine 8 (running) created at:
  main.main()
      /app/p01/main.go:13 +0x78

Goroutine 7 (running) created at:
  main.main()
      /app/p01/main.go:13 +0x78
==================
Found 2 data race(s)
exit status 66
[codigo de salida: 1]
```

> -race no adivina: instrumenta los accesos y solo ve las carreras que OCURREN en esa
> ejecucion. Que pase limpio no demuestra que no haya carrera. Eso hay que decirlo.

---

## 3. Los tres arreglos, y el tercero es el que Go prefiere

**La pregunta:** Mutex o canal?

`p02/main.go`

```go
package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

func conMutex() int {
	var mu sync.Mutex
	c := 0
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100000; j++ {
				mu.Lock()
				c++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	return c
}

func conAtomic() int64 {
	var c atomic.Int64
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100000; j++ {
				c.Add(1)
			}
		}()
	}
	wg.Wait()
	return c.Load()
}

// La tercera via, la que Go prefiere cuando encaja: nadie comparte nada.
func conCanal() int {
	res := make(chan int, 2)
	for i := 0; i < 2; i++ {
		go func() {
			c := 0
			for j := 0; j < 100000; j++ {
				c++ // cada goroutine tiene SU contador
			}
			res <- c
		}()
	}
	return <-res + <-res
}

func main() {
	fmt.Println("con mutex: ", conMutex())
	fmt.Println("con atomic:", conAtomic())
	fmt.Println("con canal: ", conCanal())
}
```

```console
$ go run -race ./p02
con mutex:  200000
con atomic: 200000
con canal:  200000
[codigo de salida: 0]
```

> Los tres dan 200000 y los tres pasan -race. La respuesta buena no elige bando: el
> mutex protege ESTADO compartido, el canal transfiere PROPIEDAD.

---

## 4. Cuanto cuesta cada uno (medido en paralelo)

**La pregunta:** Y cual es mas rapido?

`p03/bench_test.go`

```go
package p03

import (
	"sync"
	"sync/atomic"
	"testing"
)

func BenchmarkMutex(b *testing.B) {
	var mu sync.Mutex
	c := 0
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			mu.Lock()
			c++
			mu.Unlock()
		}
	})
}

func BenchmarkAtomic(b *testing.B) {
	var c atomic.Int64
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			c.Add(1)
		}
	})
}

func BenchmarkRWMutexSoloLecturas(b *testing.B) {
	var mu sync.RWMutex
	c := 42
	var sink int
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			mu.RLock()
			sink = c
			mu.RUnlock()
		}
	})
	_ = sink
}

// Sin sincronizar: NO es correcto, esta aqui solo para ver el suelo del coste.
func BenchmarkSinProteger(b *testing.B) {
	c := 0
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			c++
		}
	})
}
```

```console
$ go test -bench=. -count=2 ./p03
goos: linux
goarch: amd64
pkg: pruebas/p03
cpu: Intel(R) Core(TM) Ultra 7 270K Plus
BenchmarkMutex-24                  	22819922	        51.03 ns/op
BenchmarkMutex-24                  	24122782	        50.72 ns/op
BenchmarkAtomic-24                 	89081978	        13.73 ns/op
BenchmarkAtomic-24                 	88714968	        13.64 ns/op
BenchmarkRWMutexSoloLecturas-24    	28913950	        41.59 ns/op
BenchmarkRWMutexSoloLecturas-24    	28974472	        41.61 ns/op
BenchmarkSinProteger-24            	993916008	         1.217 ns/op
BenchmarkSinProteger-24            	987468392	         1.212 ns/op
PASS
ok  	pruebas/p03	10.120s
[codigo de salida: 0]
```

---

## 5. La repregunta: por que no dejas -race en produccion?

**La pregunta:** Cuanto cuesta el detector?

```console
$ go test -bench=BenchmarkMutex -count=2 ./p03
goos: linux
goarch: amd64
pkg: pruebas/p03
cpu: Intel(R) Core(TM) Ultra 7 270K Plus
BenchmarkMutex-24    	23402464	        52.76 ns/op
BenchmarkMutex-24    	22559956	        52.94 ns/op
PASS
ok  	pruebas/p03	2.545s
[codigo de salida: 0]
```

```console
$ go test -race -bench=BenchmarkMutex -count=2 ./p03
goos: linux
goarch: amd64
pkg: pruebas/p03
cpu: Intel(R) Core(TM) Ultra 7 270K Plus
BenchmarkMutex-24    	 4610158	       247.2 ns/op
BenchmarkMutex-24    	 4547245	       249.8 ns/op
PASS
ok  	pruebas/p03	3.827s
[codigo de salida: 0]
```

> Comparar los dos ns/op da el factor REAL de esta maquina. Va en CI y en los tests,
> no en el binario que despliegas.

---

## 6. La carrera LOGICA: todo protegido y el resultado mal

**La pregunta:** Si todos los accesos pasan por el candado, ya esta bien?

`p04/main.go`

```go
package main

import (
	"fmt"
	"sync"
)

// Cada operacion esta protegida. El PAR no lo esta.
type Saldo struct {
	mu sync.Mutex
	n  int
}

func (s *Saldo) Leer() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.n
}

func (s *Saldo) Escribir(v int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.n = v
}

func main() {
	s := &Saldo{}
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100000; j++ {
				v := s.Leer()      // protegido
				s.Escribir(v + 1)  // protegido
			}
		}()
	}
	wg.Wait()
	fmt.Println("esperaba 200000, salio:", s.Leer())
	fmt.Println()
	fmt.Println("Cada acceso paso por el candado. No hay carrera de DATOS.")
	fmt.Println("Y el resultado esta mal igual: entre el Leer y el Escribir")
	fmt.Println("se cuela la otra goroutine.")
}
```

```console
$ go run ./p04
esperaba 200000, salio: 128377

Cada acceso paso por el candado. No hay carrera de DATOS.
Y el resultado esta mal igual: entre el Leer y el Escribir
se cuela la otra goroutine.
[codigo de salida: 0]
```

```console
$ go run -race ./p04
esperaba 200000, salio: 130556

Cada acceso paso por el candado. No hay carrera de DATOS.
Y el resultado esta mal igual: entre el Leer y el Escribir
se cuela la otra goroutine.
[codigo de salida: 0]
```

> El detector pasa LIMPIO y el resultado esta mal. -race encuentra carreras de DATOS, no errores de razonamiento. Es el limite que hay que saber decir.

---

## 7. Cuantas veces acierta por casualidad (500 corridas)

**La pregunta:** Con que frecuencia el bug da el numero correcto?

`p05/main.go`

```go
package main

import (
	"fmt"
	"sync"
)

// Repite la carrera N veces y CUENTA cuantas dan el numero correcto.
// Una anecdota ("a mi me salio bien una vez") no es un dato. Esto si.
func unaCorrida() int {
	c := 0
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100000; j++ {
				c++
			}
		}()
	}
	wg.Wait()
	return c
}

func main() {
	const N = 500
	exactos, min, max := 0, 1<<62, 0
	for i := 0; i < N; i++ {
		v := unaCorrida()
		if v == 200000 {
			exactos++
		}
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	fmt.Printf("corridas: %d\n", N)
	fmt.Printf("dieron EXACTAMENTE 200000: %d\n", exactos)
	fmt.Printf("porcentaje que parecia correcto: %.1f%%\n", float64(exactos)*100/float64(N))
	fmt.Printf("peor caso: %d   mejor caso: %d\n", min, max)
}
```

```console
$ go run ./p05
corridas: 500
dieron EXACTAMENTE 200000: 112
porcentaje que parecia correcto: 22.4%
peor caso: 100000   mejor caso: 200000
[codigo de salida: 0]
```

> Esto convierte la anecdota en dato: no es que a veces acierte, es que acierta un porcentaje medible de las veces. Y ese porcentaje es la probabilidad de que el bug pase tus pruebas.
