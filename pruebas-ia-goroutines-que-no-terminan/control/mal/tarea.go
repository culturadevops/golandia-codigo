// Control del banco (705): fallos clásicos escritos a mano a propósito. El banco tiene que
// cazarlos: si alguno sale limpio, el banco no mide lo que dice.
package tarea

import (
	"context"
	"errors"
	"sync"
	"time"
)

// Canal sin búfer: si vence el plazo, la goroutine se queda para siempre en ch <- f().
func ConTimeout(d time.Duration, f func() int) (int, error) {
	ch := make(chan int)
	go func() { ch <- f() }()
	select {
	case v := <-ch:
		return v, nil
	case <-time.After(d):
		return 0, errors.New("timeout")
	}
}

// Canal sin búfer: las que no ganan se quedan bloqueadas al enviar.
func ElMasRapido(consultas []func() string) string {
	ch := make(chan string)
	for _, q := range consultas {
		go func() { ch <- q() }()
	}
	return <-ch
}

// No mira ctx al enviar: si el que lee se va, se queda bloqueada.
func Generar(ctx context.Context, n int) <-chan int {
	c := make(chan int)
	go func() {
		defer close(c)
		for i := range n {
			c <- i
		}
	}()
	return c
}

// Canal sin búfer y vuelve con el primer error: las demás se quedan al enviar.
func ProcesarTodos(items []string, f func(string) error) error {
	errs := make(chan error)
	for _, it := range items {
		go func() { errs <- f(it) }()
	}
	for range items {
		if err := <-errs; err != nil {
			return err
		}
	}
	return nil
}

// aciertos++ con RLock: varias lecturas a la vez escriben el contador → data race.
type Cache struct {
	mu       sync.RWMutex
	m        map[string]string
	aciertos int
}

func NuevaCache() *Cache { return &Cache{m: map[string]string{}} }

func (c *Cache) Set(k, v string) { c.mu.Lock(); c.m[k] = v; c.mu.Unlock() }

func (c *Cache) Get(k string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	v, ok := c.m[k]
	if ok {
		c.aciertos++
	}
	return v, ok
}

func (c *Cache) Aciertos() int { c.mu.RLock(); defer c.mu.RUnlock(); return c.aciertos }

// total++ desde varias goroutines sin protección → data race.
func ContarPrimos(nums []int, workers int) int {
	total := 0
	var wg sync.WaitGroup
	for w := range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := w; i < len(nums); i += workers {
				n, primo := nums[i], nums[i] >= 2
				for d := 2; d*d <= n; d++ {
					if n%d == 0 {
						primo = false
						break
					}
				}
				if primo {
					total++
				}
			}
		}()
	}
	wg.Wait()
	return total
}
