package tarea

import (
	"sync"
)

type Cache struct {
	items  map[string]string
	accets int
	mu     sync.RWMutex
}

func NuevaCache() *Cache {
	return &Cache{
		items:  make(map[string]string),
		accets: 0,
	}
}

func (c *Cache) Set(clave, valor string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[clave] = valor
}

func (c *Cache) Get(clave string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	val, found := c.items[clave]
	if found {
		c.accets++
	}
	return val, found
}

func (c *Cache) Aciertos() int {
	return c.accets
}
