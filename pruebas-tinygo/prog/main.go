// El cálculo de la revancha (recursos/pruebas-go-java-revancha/go/main.go), sin servidor: /suma y /trabajo como funciones,
// repetidas y medidas con el reloj del propio proceso. El MISMO archivo se compila con go build y con tinygo build
// (golandia 808, «TinyGo: el mismo Go en un binario mucho más chico»).
//
//	banco suma 300      → mediana y p99 de 300 llamadas a suma(200000), tras 50 de calentamiento
//	banco trabajo 300   → lo mismo con trabajo(20000)
package main

import (
	"fmt"
	"os"
	"runtime"
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

func suma(n int) uint64 {
	var s uint64
	for i := 1; i <= n; i++ {
		s += uint64(i) * uint64(i) % 1000003
	}
	return s
}

func trabajo(n int) string {
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
	b.WriteString(`{"total":`)
	b.WriteString(strconv.Itoa(total))
	b.WriteString(`,"primeros":[`)
	for i := 0; i < 5 && i < len(lista); i++ {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(strconv.Itoa(lista[i].ID))
	}
	b.WriteString(`]}`)
	return b.String()
}

func entero(k string, def int) int {
	if v, err := strconv.Atoi(os.Getenv(k)); err == nil {
		return v
	}
	return def
}

func main() {
	modo := os.Args[1]
	reps, _ := strconv.Atoi(os.Args[2])
	// n se lee al ejecutar (variables N_SUMA y N_TRABAJO): con una constante, LLVM (TinyGo) calcula suma(200000) al compilar.
	nSuma, nTrabajo := entero("N_SUMA", 200000), entero("N_TRABAJO", 20000)
	const cal = 50
	var res string
	ns := make([]int64, 0, reps)
	for i := 0; i < cal+reps; i++ {
		t0 := time.Now()
		if modo == "suma" {
			res = strconv.FormatUint(suma(nSuma), 10)
		} else {
			res = trabajo(nTrabajo)
		}
		if i >= cal {
			ns = append(ns, time.Since(t0).Nanoseconds())
		}
	}
	sort.Slice(ns, func(a, b int) bool { return ns[a] < ns[b] })
	fmt.Printf("%s resultado=%s mediana_ns=%d p99_ns=%d compilador=%s\n", modo, res, ns[len(ns)/2], ns[len(ns)*99/100-1],
		runtime.Compiler)
}
