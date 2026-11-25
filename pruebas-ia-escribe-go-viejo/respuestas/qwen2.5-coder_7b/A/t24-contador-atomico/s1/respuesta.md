```go
package tarea

import (
	"sync/atomic"
)

type Contador int64

func (c *Contador) Incrementar() {
	atomic.AddInt64((*int64)(c), 1)
}

func (c *Contador) Valor() int64 {
	return atomic.LoadInt64((*int64)(c))
}
```