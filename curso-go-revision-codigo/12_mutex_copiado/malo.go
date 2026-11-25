//go:build malo

package pedidos

import "sync"

// Contador cuenta los pedidos atendidos; lo comparten todos los handlers.
type Contador struct {
	mu sync.Mutex
	n  int
}

func (c Contador) Sumar() {
	c.mu.Lock()
	c.n++
	c.mu.Unlock()
}

func (c *Contador) Valor() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.n
}
