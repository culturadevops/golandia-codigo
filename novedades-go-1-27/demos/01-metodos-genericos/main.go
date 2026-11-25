// Go 1.27: un método puede declarar sus propios parámetros de tipo.
package main

import (
	"fmt"
	"math/rand/v2"
	"time"
)

type Cache struct{ datos map[string]any }

// Get es un método genérico: T es del método, no del tipo Cache.
func (c *Cache) Get[T any](clave string) (T, bool) {
	v, ok := c.datos[clave].(T)
	return v, ok
}

func main() {
	c := &Cache{datos: map[string]any{"puerto": 8080, "host": "localhost"}}

	puerto, ok := c.Get[int]("puerto")
	fmt.Println("puerto:", puerto, ok)

	host, ok := c.Get[string]("host")
	fmt.Println("host:", host, ok)

	_, ok = c.Get[string]("puerto") // existe, pero no es string
	fmt.Println("puerto como string:", ok)

	// math/rand/v2 ahora trae N como método de *Rand.
	r := rand.New(rand.NewPCG(1, 2))
	fmt.Println("r.N(10 * time.Second):", r.N(10*time.Second))
}
