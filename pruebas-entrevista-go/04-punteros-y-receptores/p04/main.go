package main

import (
	"fmt"
	"sync"
)

type Cache struct {
	mu     sync.Mutex
	datos  map[string]string
}

// El error: recibe el Cache POR VALOR, y eso copia el mutex.
func (c Cache) Leer(k string) string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.datos[k]
}

func main() {
	c := Cache{datos: map[string]string{"a": "1"}}
	fmt.Println(c.Leer("a"))
}
