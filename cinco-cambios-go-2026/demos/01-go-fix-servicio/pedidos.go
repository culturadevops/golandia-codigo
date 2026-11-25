// Un servicio de pedidos escrito «como siempre»: compila, pasa go vet y sus pruebas.
// Golandia 714: go fix de Go 1.27 tiene que decir qué partes son Go viejo.
package pedidos

import (
	"errors"
	"io/fs"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type Pedido struct {
	ID      string    `json:"id"`
	Cliente string    `json:"cliente"`
	Total   int       `json:"total"`
	Alta    time.Time `json:"alta,omitempty"`
}

var procesados int64

// Etiquetas convierte "a,b,c" en una lista sin espacios.
func Etiquetas(s string) []string {
	var res []string
	for _, p := range strings.Split(s, ",") {
		res = append(res, strings.TrimSpace(p))
	}
	return res
}

// Cliente saca el cliente de una línea "cliente=ana".
func Cliente(linea string) (string, bool) {
	if strings.HasPrefix(linea, "cliente=") {
		return strings.TrimPrefix(linea, "cliente="), true
	}
	return "", false
}

// EsVIP dice si el cliente está en la lista.
func EsVIP(vips []string, cliente string) bool {
	for _, v := range vips {
		if v == cliente {
			return true
		}
	}
	return false
}

// Copiar pasa los totales de un mapa a otro.
func Copiar(dst, src map[string]int) {
	for k, v := range src {
		dst[k] = v
	}
}

// Ordenar ordena los totales de menor a mayor.
func Ordenar(totales []int) {
	sort.Slice(totales, func(i, j int) bool { return totales[i] < totales[j] })
}

// Reintentos devuelve los tiempos de espera: 100 ms, 200 ms, 400 ms...
func Reintentos(n int) []time.Duration {
	var res []time.Duration
	for i := 0; i < n; i++ {
		res = append(res, time.Duration(100<<i)*time.Millisecond)
	}
	return res
}

// Procesar aplica f a cada pedido en paralelo.
func Procesar(ps []Pedido, f func(Pedido) int) []int {
	res := make([]int, len(ps))
	var wg sync.WaitGroup
	for i, p := range ps {
		i, p := i, p
		wg.Add(1)
		go func() {
			defer wg.Done()
			res[i] = f(p)
			atomic.AddInt64(&procesados, 1)
		}()
	}
	wg.Wait()
	return res
}

// Procesados cuenta los pedidos procesados desde que arrancó.
func Procesados() int64 { return atomic.LoadInt64(&procesados) }

// RutaFallida saca la ruta de un error de archivo.
func RutaFallida(err error) (string, bool) {
	var pe *fs.PathError
	if errors.As(err, &pe) {
		return pe.Path, true
	}
	return "", false
}

// Campos devuelve los datos de un pedido como mapa genérico.
func Campos(p Pedido) map[string]interface{} {
	return map[string]interface{}{"id": p.ID, "total": p.Total}
}
