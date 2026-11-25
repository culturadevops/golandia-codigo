```go
package tarea

import (
	"sync"
	"sync/atomic"
)

// Cache es un almacén clave-valor seguro para uso concurrente.
type Cache struct {
	mu       sync.RWMutex
	datos    map[string]string
	aciertos atomic.Int64
}

// NuevaCache crea una Cache vacía lista para usarse.
func NuevaCache() *Cache {
	return &Cache{datos: make(map[string]string)}
}

// Set guarda valor bajo clave, reemplazando cualquier valor previo.
func (c *Cache) Set(clave, valor string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.datos[clave] = valor
}

// Get devuelve el valor asociado a clave y si existía.
// Las lecturas concurrentes no se bloquean entre sí; el contador de
// aciertos es atómico para poder actualizarse bajo el bloqueo de lectura.
func (c *Cache) Get(clave string) (string, bool) {
	c.mu.RLock()
	valor, ok := c.datos[clave]
	c.mu.RUnlock()
	if ok {
		c.aciertos.Add(1)
	}
	return valor, ok
}

// Aciertos devuelve cuántas veces Get encontró la clave.
func (c *Cache) Aciertos() int {
	return int(c.aciertos.Load())
}
```