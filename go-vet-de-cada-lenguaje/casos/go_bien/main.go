package main

import (
	"errors"
	"fmt"
	"os"
	"sync"
)

type ErrNoEncontrado struct{ ID int }

func (e ErrNoEncontrado) Error() string { return fmt.Sprintf("no encontrado: %d", e.ID) }

func buscar(id int) error {
	return fmt.Errorf("buscar: %w", ErrNoEncontrado{ID: id})
}

type Contador struct {
	mu sync.Mutex
	n  int
}

func (c *Contador) Inc() {
	c.mu.Lock()
	c.n++
	c.mu.Unlock()
}

func leer(ruta string) ([]byte, error) {
	return os.ReadFile(ruta)
}

func main() {
	err := buscar(7)
	var e ErrNoEncontrado
	fmt.Println("errors.As(err, &ErrNoEncontrado{}) =", errors.As(err, &e))

	var c Contador
	var wg sync.WaitGroup
	for range 1000 {
		wg.Go(c.Inc)
	}
	wg.Wait()
	fmt.Println("después de 1000 Inc, n =", c.n)

	_, err = leer("no-existe.txt")
	fmt.Println("leer:", err != nil)
}
