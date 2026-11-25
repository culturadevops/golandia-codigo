package cache

import (
	"sync"
	"time"
)

// Cache guarda valores que caducan a los ttl. Un limpiador en segundo plano borra lo caducado
// cada intervalo. Es el tipo de código que se prueba con time.Sleep... y tarda.
type Cache struct {
	mu    sync.Mutex
	ttl   time.Duration
	datos map[string]entrada
	parar chan struct{}
}

type entrada struct {
	valor  string
	expira time.Time
}

func Nueva(ttl, intervalo time.Duration) *Cache {
	c := &Cache{ttl: ttl, datos: map[string]entrada{}, parar: make(chan struct{})}
	go c.limpiar(intervalo)
	return c
}

func (c *Cache) Poner(k, v string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.datos[k] = entrada{v, time.Now().Add(c.ttl)}
}

func (c *Cache) Tam() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.datos)
}

func (c *Cache) Cerrar() { close(c.parar) }

func (c *Cache) limpiar(intervalo time.Duration) {
	t := time.NewTicker(intervalo)
	defer t.Stop()
	for {
		select {
		case <-c.parar:
			return
		case ahora := <-t.C:
			c.mu.Lock()
			for k, e := range c.datos {
				if ahora.After(e.expira) {
					delete(c.datos, k)
				}
			}
			c.mu.Unlock()
		}
	}
}
