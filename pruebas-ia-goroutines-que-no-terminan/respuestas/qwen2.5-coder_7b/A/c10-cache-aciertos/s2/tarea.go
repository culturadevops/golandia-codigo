package tarea

import (
	"sync"
)

type Cache struct {
	data    map[string]string
	acc     int
	accLock sync.Mutex
	mu      sync.RWMutex
}

func NuevaCache() *Cache {
	return &Cache{
		data: make(map[string]string),
	}
}

func (c *Cache) Set(clave, valor string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[clave] = valor
}

func (c *Cache) Get(clave string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	val, ok := c.data[clave]
	if ok {
		c.accLock.Lock()
		c.acc++
		c.accLock.Unlock()
	}
	return val, ok
}

func (c *Cache) Aciertos() int {
	c.accLock.Lock()
	defer c.accLock.Unlock()
	return c.acc
}
