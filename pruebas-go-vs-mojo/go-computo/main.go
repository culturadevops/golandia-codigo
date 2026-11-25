// El cómputo de «¿Go o Mojo?» (golandia 904): suma y trabajo copiados TAL CUAL del servidor de Go de la serie
// (recursos/pruebas-go-java-revancha/go/main.go), sin net/http, con el mismo arnés que computo.mojo y computo.py:
//   computo una    → una llamada de cada función (para comprobar que dan lo mismo)
//   computo medir  → 50 de calentamiento y 300 medidas de cada una; imprime los 300 tiempos en ns
//   computo        → nada (para medir el arranque del proceso)
package main

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

type item struct {
	ID     int
	Nombre string
	Valor  int
}

func suma(n int) string {
	t0 := time.Now()
	var s uint64
	for i := 1; i <= n; i++ {
		s += uint64(i) * uint64(i) % 1000003
	}
	return fmt.Sprintf(`{"lang":"go","n":%d,"suma":%d,"ns":%d}`, n, s, time.Since(t0).Nanoseconds())
}

func trabajo(n int) string {
	t0 := time.Now()
	m := make(map[string]*item, n)
	for i := 0; i < n; i++ {
		it := &item{ID: i, Nombre: "item-" + strconv.Itoa(i), Valor: (i * 7919) % 1000}
		m[it.Nombre] = it
	}
	lista := make([]*item, 0, len(m))
	total := 0
	for _, it := range m {
		lista = append(lista, it)
		total += it.Valor
	}
	sort.Slice(lista, func(a, b int) bool {
		if lista[a].Valor != lista[b].Valor {
			return lista[a].Valor < lista[b].Valor
		}
		return lista[a].ID < lista[b].ID
	})
	var b strings.Builder
	b.WriteString(`{"lang":"go","n":`)
	b.WriteString(strconv.Itoa(n))
	b.WriteString(`,"total":`)
	b.WriteString(strconv.Itoa(total))
	b.WriteString(`,"primeros":[`)
	for i := 0; i < 50 && i < len(lista); i++ {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(`{"id":`)
		b.WriteString(strconv.Itoa(lista[i].ID))
		b.WriteString(`,"nombre":"`)
		b.WriteString(lista[i].Nombre)
		b.WriteString(`","valor":`)
		b.WriteString(strconv.Itoa(lista[i].Valor))
		b.WriteByte('}')
	}
	b.WriteString(`],"ns":`)
	b.WriteString(strconv.FormatInt(time.Since(t0).Nanoseconds(), 10))
	b.WriteByte('}')
	return b.String()
}

func medir(nombre string, f func(int) string, n int) {
	var b strings.Builder
	b.WriteString(nombre)
	for i := 0; i < 350; i++ {
		t0 := time.Now()
		r := f(n)
		dt := time.Since(t0).Nanoseconds()
		if i >= 50 {
			b.WriteString(" " + strconv.FormatInt(dt, 10))
		}
		if len(r) == 0 {
			fmt.Println(r)
		}
	}
	fmt.Println(b.String())
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "una" {
		fmt.Println(suma(200000))
		fmt.Println(trabajo(20000))
	} else if len(os.Args) > 1 && os.Args[1] == "medir" {
		medir("SUMA", suma, 200000)
		medir("TRABAJO", trabajo, 20000)
	}
}
