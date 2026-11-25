package tarea

import (
	"sync"
)

type Cache struct {
	data    map[string]string
	mu      sync.RWMutex
	aciertos int
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
	valor, encontrado := c.data[clave]
	if encontrado {
		c.aciertos++
	}
	return valor, encontrado
}

func (c *Cache) Aciertos() int {
	return c.aciertos
}
