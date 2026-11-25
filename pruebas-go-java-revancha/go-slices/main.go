// VARIANTE del experimento (2026-10-19): igual que go/main.go pero ordenando con slices.SortFunc (sin reflexión).
// La revancha de Go contra Java (2026-10-19): el mismo servidor del vídeo 148 más /trabajo, una carga
// que reserva memoria (objetos, mapa, orden y JSON armado a mano), para que no sea un «hola mundo».
// SOLO librería estándar (net/http). El JSON se arma a mano igual que en Java, para hacer el mismo trabajo.
package main

import (
	"fmt"
	"net/http"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"
)

var arranque = time.Now()

type item struct {
	ID     int
	Nombre string
	Valor  int
}

func num(r *http.Request, k string) int {
	n, _ := strconv.Atoi(r.URL.Query().Get(k))
	return n
}

func responder(w http.ResponseWriter, s string) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(s))
}

func suma(w http.ResponseWriter, r *http.Request) {
	n := num(r, "n")
	t0 := time.Now()
	var s uint64
	for i := 1; i <= n; i++ {
		s += uint64(i) * uint64(i) % 1000003
	}
	responder(w, fmt.Sprintf(`{"lang":"go","n":%d,"suma":%d,"ns":%d}`, n, s, time.Since(t0).Nanoseconds()))
}

func trabajo(w http.ResponseWriter, r *http.Request) {
	n := num(r, "n")
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
	slices.SortFunc(lista, func(a, b *item) int {
		if a.Valor != b.Valor {
			return a.Valor - b.Valor
		}
		return a.ID - b.ID
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
	responder(w, b.String())
}

func espera(w http.ResponseWriter, r *http.Request) {
	ms := num(r, "ms")
	time.Sleep(time.Duration(ms) * time.Millisecond)
	responder(w, fmt.Sprintf(`{"lang":"go","n":%d,"suma":0,"ns":0}`, ms))
}

func main() {
	http.HandleFunc("/suma", suma)
	http.HandleFunc("/trabajo", trabajo)
	http.HandleFunc("/espera", espera)
	fmt.Fprintf(os.Stderr, "LISTO_EN_MS %.3f\n", float64(time.Since(arranque).Microseconds())/1000)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
