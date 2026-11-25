package tarea

import (
	"sync/atomic"
)

type Contador struct {
	valor atomic.Int64
}

func (c *Contador) Incrementar() {
	c.valor.Add(1)
}

func (c *Contador) Valor() int64 {
	return c.valor.Load()
}
