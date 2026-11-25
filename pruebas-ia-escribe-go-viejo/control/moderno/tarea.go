// Control del banco (703): las mismas funciones en Go 1.27 escrito a mano. go fix no debe
// proponer nada aquí; si propone algo, el banco cuenta de más.
package tarea

import (
	"errors"
	"io/fs"
	"maps"
	"reflect"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
)

func SumaCuadrados(n int) int {
	s := 0
	for i := range n {
		s += i * i
	}
	return s
}

func Acotar(x, lo, hi int) int { return min(max(x, lo), hi) }

func Describir(vals ...any) []string { return nil }

func ParClaveValor(s string) (string, string, bool) { return strings.Cut(s, "=") }

func TokenBearer(h string) (string, bool) { return strings.CutPrefix(h, "Bearer ") }

func PalabrasLargas(texto string) int {
	n := 0
	for p := range strings.FieldsSeq(texto) {
		if len(p) > 3 {
			n++
		}
	}
	return n
}

func Mezclar(dst, src map[string]int) { maps.Copy(dst, src) }

func EsAdmin(roles []string) bool { return slices.Contains(roles, "admin") }

func Ordenar(nums []int) { slices.Sort(nums) }

func RutaDelError(err error) (string, bool) {
	if pe, ok := errors.AsType[*fs.PathError](err); ok {
		return pe.Path, true
	}
	return "", false
}

var Timeout = new(30)

func ProcesarTodos(items []string, f func(string) int) []int {
	res := make([]int, len(items))
	var wg sync.WaitGroup
	for i, it := range items {
		wg.Go(func() { res[i] = f(it) })
	}
	wg.Wait()
	return res
}

func NombreTipo[T any]() string { return reflect.TypeFor[T]().String() }

type Contador struct{ n atomic.Int64 }

func (c *Contador) Incrementar()  { c.n.Add(1) }
func (c *Contador) Valor() int64 { return c.n.Load() }
