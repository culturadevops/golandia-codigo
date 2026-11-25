// Control del banco (705): implementaciones correctas escritas a mano. Si alguna de estas da fuga,
// race o falla, el que falla es el banco, no la IA.
package tarea

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

func ProcesarConWorkers(ctx context.Context, trabajos []int, workers int, f func(int) int) ([]int, error) {
	res := make([]int, len(trabajos))
	idx := make(chan int)
	var wg sync.WaitGroup
	for range workers {
		wg.Go(func() {
			for i := range idx {
				res[i] = f(trabajos[i])
			}
		})
	}
	var err error
enviar:
	for i := range trabajos {
		select {
		case idx <- i:
		case <-ctx.Done():
			err = ctx.Err()
			break enviar
		}
	}
	close(idx)
	if err != nil {
		return nil, err // los workers acaban el trabajo en curso y salen: idx está cerrado
	}
	wg.Wait()
	return res, nil
}

func ConTimeout(d time.Duration, f func() int) (int, error) {
	ch := make(chan int, 1)
	go func() { ch <- f() }()
	select {
	case v := <-ch:
		return v, nil
	case <-time.After(d):
		return 0, errors.New("timeout")
	}
}

func LlamarConContexto(ctx context.Context, f func() (string, error)) (string, error) {
	type r struct {
		s   string
		err error
	}
	ch := make(chan r, 1)
	go func() { s, err := f(); ch <- r{s, err} }()
	select {
	case v := <-ch:
		return v.s, v.err
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

func ElMasRapido(consultas []func() string) string {
	ch := make(chan string, len(consultas))
	for _, q := range consultas {
		go func() { ch <- q() }()
	}
	return <-ch
}

func FanOutFanIn(ctx context.Context, entrada []int, workers int, f func(int) int) <-chan int {
	in := make(chan int)
	out := make(chan int)
	go func() {
		defer close(in)
		for _, x := range entrada {
			select {
			case in <- x:
			case <-ctx.Done():
				return
			}
		}
	}()
	var wg sync.WaitGroup
	for range workers {
		wg.Go(func() {
			for x := range in {
				select {
				case out <- f(x):
				case <-ctx.Done():
					return
				}
			}
		})
	}
	go func() { wg.Wait(); close(out) }()
	return out
}

func Generar(ctx context.Context, n int) <-chan int {
	c := make(chan int)
	go func() {
		defer close(c)
		for i := range n {
			select {
			case c <- i:
			case <-ctx.Done():
				return
			}
		}
	}()
	return c
}

func Cuadrados(ctx context.Context, in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for x := range in {
			select {
			case out <- x * x:
			case <-ctx.Done():
				return
			}
		}
	}()
	return out
}

func Unir(ctx context.Context, canales ...<-chan int) <-chan int {
	out := make(chan int)
	var wg sync.WaitGroup
	for _, c := range canales {
		wg.Go(func() {
			for v := range c {
				select {
				case out <- v:
				case <-ctx.Done():
					return
				}
			}
		})
	}
	go func() { wg.Wait(); close(out) }()
	return out
}

func ProcesarTodos(items []string, f func(string) error) error {
	errs := make(chan error, len(items))
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

type Cache struct {
	mu       sync.RWMutex
	m        map[string]string
	aciertos atomic.Int64
}

func NuevaCache() *Cache { return &Cache{m: map[string]string{}} }

func (c *Cache) Set(k, v string) { c.mu.Lock(); c.m[k] = v; c.mu.Unlock() }

func (c *Cache) Get(k string) (string, bool) {
	c.mu.RLock()
	v, ok := c.m[k]
	c.mu.RUnlock()
	if ok {
		c.aciertos.Add(1)
	}
	return v, ok
}

func (c *Cache) Aciertos() int { return int(c.aciertos.Load()) }

func esPrimo(n int) bool {
	if n < 2 {
		return false
	}
	for d := 2; d*d <= n; d++ {
		if n%d == 0 {
			return false
		}
	}
	return true
}

func ContarPrimos(nums []int, workers int) int {
	var total atomic.Int64
	var wg sync.WaitGroup
	for w := range workers {
		wg.Go(func() {
			for i := w; i < len(nums); i += workers {
				if esPrimo(nums[i]) {
					total.Add(1)
				}
			}
		})
	}
	wg.Wait()
	return int(total.Load())
}

func Consumir(ctx context.Context, trabajos <-chan int, consumidores int, f func(int)) error {
	var wg sync.WaitGroup
	for range consumidores {
		wg.Go(func() {
			for {
				select {
				case <-ctx.Done():
					return
				case t, ok := <-trabajos:
					if !ok {
						return
					}
					f(t)
				}
			}
		})
	}
	wg.Wait()
	return ctx.Err()
}
