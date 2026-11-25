package tarea

import (
	"sync/atomic"
)

type Contador struct {
	value atomic.Int64
}

func (c *Contador) Incrementar() {
	c.value.Add(1)
}

func (c *Contador) Valor() int64 {
	return c.value.Load()
}
