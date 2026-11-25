package main

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

const capacidadReal = 5
const totalTareas = 300

var actuales int64
var pico int64
var errores int64

func resetear() {
	atomic.StoreInt64(&actuales, 0)
	atomic.StoreInt64(&pico, 0)
	atomic.StoreInt64(&errores, 0)
}

func actualizarPico(n int64) {
	for {
		p := atomic.LoadInt64(&pico)
		if n <= p {
			return
		}
		if atomic.CompareAndSwapInt64(&pico, p, n) {
			return
		}
	}
}

// llamarServicio simula un servicio downstream que solo aguanta 5 llamadas a la vez.
func llamarServicio() error {
	n := atomic.AddInt64(&actuales, 1)
	defer atomic.AddInt64(&actuales, -1)
	actualizarPico(n)
	if n > capacidadReal {
		atomic.AddInt64(&errores, 1)
		time.Sleep(2 * time.Millisecond)
		return fmt.Errorf("demasiadas conexiones simultaneas")
	}
	time.Sleep(10 * time.Millisecond)
	return nil
}

func sinControl() {
	resetear()
	var wg sync.WaitGroup
	inicio := time.Now()
	for i := 0; i < totalTareas; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			llamarServicio()
		}()
	}
	maxGo := int64(runtime.NumGoroutine())
	wg.Wait()
	duracion := time.Since(inicio)
	fmt.Printf("SIN CONTROL   -> pico de llamadas simultaneas: %d | errores: %d | goroutines lanzadas de golpe: %d | duro: %s\n",
		atomic.LoadInt64(&pico), atomic.LoadInt64(&errores), maxGo, duracion.Round(time.Millisecond))
}

func conWorkerPool(workers int) {
	resetear()
	trabajos := make(chan int, totalTareas)
	var wg sync.WaitGroup
	inicio := time.Now()
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range trabajos {
				llamarServicio()
			}
		}()
	}
	for i := 0; i < totalTareas; i++ {
		trabajos <- i
	}
	close(trabajos)
	wg.Wait()
	duracion := time.Since(inicio)
	fmt.Printf("WORKER POOL   -> pico de llamadas simultaneas: %d | errores: %d | goroutines fijas: %d           | duro: %s\n",
		atomic.LoadInt64(&pico), atomic.LoadInt64(&errores), workers, duracion.Round(time.Millisecond))
}

// fanIn: N workers, cada uno produce resultados en su propio canal.
// Un solo goroutine junta todo en un canal compartido (el fan-in de verdad).
func fanIn(canales ...<-chan int) <-chan int {
	salida := make(chan int)
	var wg sync.WaitGroup
	for _, c := range canales {
		wg.Add(1)
		go func(c <-chan int) {
			defer wg.Done()
			for v := range c {
				salida <- v
			}
		}(c)
	}
	go func() {
		wg.Wait()
		close(salida)
	}()
	return salida
}

func productor(id, cuantos int) <-chan int {
	c := make(chan int)
	go func() {
		defer close(c)
		for i := 0; i < cuantos; i++ {
			c <- id*1000 + i
		}
	}()
	return c
}

func demoFanIn() {
	const trabajadores = 5
	const porTrabajador = 40
	canales := make([]<-chan int, trabajadores)
	for i := 0; i < trabajadores; i++ {
		canales[i] = productor(i, porTrabajador)
	}
	total := 0
	vistos := make(map[int]bool)
	for v := range fanIn(canales...) {
		total++
		vistos[v] = true
	}
	fmt.Printf("FAN-IN         -> %d productores, %d valores esperados, %d recibidos, %d unicos\n",
		trabajadores, trabajadores*porTrabajador, total, len(vistos))
}

func main() {
	fmt.Printf("total de tareas: %d, capacidad real del servicio: %d\n\n", totalTareas, capacidadReal)
	sinControl()
	conWorkerPool(capacidadReal)
	fmt.Println()
	demoFanIn()
}
