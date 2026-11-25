```go
package tarea

import (
	"sync"
)

type Cache struct {
	data      map[string]string
	accidentes int
	mutex     sync.RWMutex
}

func NuevaCache() *Cache {
	return &Cache{
		data: make(map[string]string),
	}
}

func (c *Cache) Set(clave, valor string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.data[clave] = valor
}

func (c *Cache) Get(clave string) (string, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	valor, existe := c.data[clave]
	if existe {
		c.accidentes++
	}
	return valor, existe
}

func (c *Cache) Aciertos() int {
	return c.accidentes
}
```