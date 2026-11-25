```go
package tarea

import "sync/atomic"

// Contador es un contador seguro para uso concurrente desde varias goroutines.
type Contador struct {
	valor atomic.Int64
}

// Incrementar aumenta el contador en uno de forma atómica.
func (c *Contador) Incrementar() {
	c.valor.Add(1)
}

// Valor devuelve el valor actual del contador de forma atómica.
func (c *Contador) Valor() int64 {
	return c.valor.Load()
}
```