// Control del banco (703): Go «de 2019» escrito a mano a propósito. go fix tiene que marcar
// cada modismo; si alguno no sale, el banco no mide lo que dice.
package tarea

import (
	"errors"
	"io/fs"
	"reflect"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
)

func SumaCuadrados(n int) int {
	s := 0
	for i := 0; i < n; i++ {
		s += i * i
	}
	return s
}

func Acotar(x, lo, hi int) int {
	if x < lo {
		x = lo
	}
	if x > hi {
		x = hi
	}
	return x
}

func Describir(vals ...interface{}) []string { return nil }

func ParClaveValor(s string) (string, string, bool) {
	i := strings.Index(s, "=")
	if i < 0 {
		return "", "", false
	}
	return s[:i], s[i+1:], true
}

func TokenBearer(h string) (string, bool) {
	if strings.HasPrefix(h, "Bearer ") {
		return strings.TrimPrefix(h, "Bearer "), true
	}
	return "", false
}

func PalabrasLargas(texto string) int {
	n := 0
	for _, p := range strings.Fields(texto) {
		if len(p) > 3 {
			n++
		}
	}
	return n
}

func Mezclar(dst, src map[string]int) {
	for k, v := range src {
		dst[k] = v
	}
}

func EsAdmin(roles []string) bool {
	for _, r := range roles {
		if r == "admin" {
			return true
		}
	}
	return false
}

func Ordenar(nums []int) { sort.Slice(nums, func(i, j int) bool { return nums[i] < nums[j] }) }

func RutaDelError(err error) (string, bool) {
	var pe *fs.PathError
	if errors.As(err, &pe) {
		return pe.Path, true
	}
	return "", false
}

func ptr[T any](v T) *T { return &v }

var Timeout = ptr(30)

func ProcesarTodos(items []string, f func(string) int) []int {
	res := make([]int, len(items))
	var wg sync.WaitGroup
	for i, it := range items {
		i, it := i, it
		wg.Add(1)
		go func() {
			defer wg.Done()
			res[i] = f(it)
		}()
	}
	wg.Wait()
	return res
}

func NombreTipo[T any]() string { return reflect.TypeOf((*T)(nil)).Elem().String() }

type Contador struct{ n int64 }

func (c *Contador) Incrementar()  { atomic.AddInt64(&c.n, 1) }
func (c *Contador) Valor() int64 { return atomic.LoadInt64(&c.n) }
