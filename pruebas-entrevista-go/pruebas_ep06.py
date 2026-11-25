# -*- coding: utf-8 -*-
import sys
from pathlib import Path
sys.path.insert(0, str(Path(__file__).resolve().parent))
from probar import episodio

P1 = '''package main

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
'''

P2 = '''package main

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
'''

BENCH = '''package p03

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
'''

P4 = r'''package main

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
'''


P5 = r'''package main

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
'''


episodio(
    "06-carreras-mutex-y-race",
    "Episodio 6/8 — carreras de datos, mutex, atomic y -race",
    [
        {"nombre": "1. Dos goroutines sumando: cuanto sale? (tres corridas)",
         "pregunta": "Dos goroutines suman 100.000 cada una. Cuanto vale el contador?",
         "archivos": {"p01/main.go": P1},
         "correr": [(["go", "run", "./p01"], False),
                    (["go", "run", "./p01"], False),
                    (["go", "run", "./p01"], False)],
         "nota": "Tres corridas, tres numeros. Lo peligroso no es que falle: es que A VECES da el\nnumero bueno, y entonces el bug se va a produccion."},
        {"nombre": "2. El detector de carreras lo senala con el dedo",
         "pregunta": "Como demuestras que hay una carrera?",
         "archivos": {},
         "correr": [(["go", "run", "-race", "./p01"], True)],
         "nota": "-race no adivina: instrumenta los accesos y solo ve las carreras que OCURREN en esa\nejecucion. Que pase limpio no demuestra que no haya carrera. Eso hay que decirlo."},
        {"nombre": "3. Los tres arreglos, y el tercero es el que Go prefiere",
         "pregunta": "Mutex o canal?",
         "archivos": {"p02/main.go": P2},
         "correr": [(["go", "run", "-race", "./p02"], False)],
         "nota": "Los tres dan 200000 y los tres pasan -race. La respuesta buena no elige bando: el\nmutex protege ESTADO compartido, el canal transfiere PROPIEDAD."},
        {"nombre": "4. Cuanto cuesta cada uno (medido en paralelo)",
         "pregunta": "Y cual es mas rapido?",
         "archivos": {"p03/bench_test.go": BENCH},
         "correr": [(["go", "test", "-bench=.", "-count=2", "./p03"], False)]},
        {"nombre": "5. La repregunta: por que no dejas -race en produccion?",
         "pregunta": "Cuanto cuesta el detector?",
         "archivos": {},
         "correr": [(["go", "test", "-bench=BenchmarkMutex", "-count=2", "./p03"], False),
                    (["go", "test", "-race", "-bench=BenchmarkMutex", "-count=2", "./p03"], False)],
         "nota": "Comparar los dos ns/op da el factor REAL de esta maquina. Va en CI y en los tests,\nno en el binario que despliegas."},
        {"nombre": "6. La carrera LOGICA: todo protegido y el resultado mal",
         "pregunta": "Si todos los accesos pasan por el candado, ya esta bien?",
         "archivos": {"p04/main.go": P4},
         "correr": [(["go", "run", "./p04"], False),
                    (["go", "run", "-race", "./p04"], False)],
         "nota": "El detector pasa LIMPIO y el resultado esta mal. -race encuentra carreras de "
                 "DATOS, no errores de razonamiento. Es el limite que hay que saber decir."},
        {"nombre": "7. Cuantas veces acierta por casualidad (500 corridas)",
         "pregunta": "Con que frecuencia el bug da el numero correcto?",
         "archivos": {"p05/main.go": P5},
         "correr": [(["go", "run", "./p05"], False)],
         "nota": "Esto convierte la anecdota en dato: no es que a veces acierte, es que acierta un porcentaje medible de las veces. Y ese porcentaje es la probabilidad de que el bug pase tus pruebas."},
    ])
