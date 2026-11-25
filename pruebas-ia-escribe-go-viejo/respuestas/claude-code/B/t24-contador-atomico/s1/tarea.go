package tarea

import "sync/atomic"

// Contador es un contador seguro para uso concurrente desde varias goroutines.
// El valor cero está listo para usarse.
type Contador struct {
	n atomic.Int64
}

// Incrementar suma uno al contador de forma atómica.
func (c *Contador) Incrementar() {
	c.n.Add(1)
}

// Valor devuelve el valor actual del contador.
func (c *Contador) Valor() int64 {
	return c.n.Load()
}
