// Demo 705: el perfil en un servicio. Con net/http/pprof importado, Go 1.27 sirve
// /debug/pprof/goroutineleak. El handler /buscar tiene la fuga clásica del timeout.
package main

import (
	"fmt"
	"io"
	"net/http"
	_ "net/http/pprof"
	"strings"
	"time"
)

func buscar(w http.ResponseWriter, r *http.Request) {
	ch := make(chan string) // sin búfer
	go func() {
		time.Sleep(50 * time.Millisecond) // la consulta lenta
		ch <- "resultado"
	}()
	select {
	case v := <-ch:
		fmt.Fprintln(w, v)
	case <-time.After(10 * time.Millisecond):
		http.Error(w, "timeout", http.StatusGatewayTimeout)
	}
}

func main() {
	http.HandleFunc("/buscar", buscar)
	go http.ListenAndServe("127.0.0.1:6060", nil)
	time.Sleep(200 * time.Millisecond)
	for range 5 {
		r, _ := http.Get("http://127.0.0.1:6060/buscar")
		r.Body.Close()
	}
	time.Sleep(200 * time.Millisecond)
	r, err := http.Get("http://127.0.0.1:6060/debug/pprof/goroutineleak?debug=1")
	if err != nil {
		panic(err)
	}
	b, _ := io.ReadAll(r.Body)
	for _, l := range strings.Split(string(b), "\n")[:4] {
		fmt.Println(l)
	}
}
