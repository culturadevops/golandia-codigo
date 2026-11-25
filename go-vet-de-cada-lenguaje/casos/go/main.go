package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"sync"
)

// 1 · Envolver un error con %w pasando un PUNTERO a un tipo cuyo método Error es de valor.
type ErrNoEncontrado struct{ ID int }

func (e ErrNoEncontrado) Error() string { return fmt.Sprintf("no encontrado: %d", e.ID) }

func buscar(id int) error {
	return fmt.Errorf("buscar: %w", &ErrNoEncontrado{ID: id})
}

// 2 · Un contador con mutex y receptor de VALOR: cada llamada copia el mutex y el número.
type Contador struct {
	mu sync.Mutex
	n  int
}

func (c Contador) Inc() {
	c.mu.Lock()
	c.n++
	c.mu.Unlock()
}

// 3 · Una API vieja: io/ioutil está obsoleto (staticcheck SA1019).
func leer(ruta string) ([]byte, error) {
	return ioutil.ReadFile(ruta)
}

func main() {
	err := buscar(7)
	var e ErrNoEncontrado
	fmt.Println("errors.As(err, &ErrNoEncontrado{}) =", errors.As(err, &e))

	var c Contador
	var wg sync.WaitGroup
	for range 1000 {
		wg.Add(1)
		go func() { defer wg.Done(); c.Inc() }()
	}
	wg.Wait()
	fmt.Println("después de 1000 Inc, n =", c.n)

	_, err = leer("no-existe.txt")
	fmt.Println("leer:", err != nil)
}
