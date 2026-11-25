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
