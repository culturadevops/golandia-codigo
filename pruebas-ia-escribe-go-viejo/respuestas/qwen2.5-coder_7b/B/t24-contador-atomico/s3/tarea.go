package tarea

import (
	"sync/atomic"
)

type Contador struct {
	valor int64
}

func (c *Contador) Incrementar() {
	atomic.AddInt64(&c.valor, 1)
}

func (c *Contador) Valor() int64 {
	return atomic.LoadInt64(&c.valor)
}
