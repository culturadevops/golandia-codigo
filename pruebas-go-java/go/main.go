// El servidor de la comparativa, en Go: SOLO librería estándar (net/http).
// El mismo contrato que el de Java: /suma hace CPU de verdad, /espera simula E/S.
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"
)

var arranque = time.Now()

type respuesta struct {
	Lang string `json:"lang"`
	N    int    `json:"n"`
	Suma uint64 `json:"suma"`
	Ns   int64  `json:"ns"`
}

func suma(w http.ResponseWriter, r *http.Request) {
	n, _ := strconv.Atoi(r.URL.Query().Get("n"))
	t0 := time.Now()
	var s uint64
	for i := 1; i <= n; i++ {
		s += uint64(i) * uint64(i) % 1000003
	}
	ns := time.Since(t0).Nanoseconds()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(respuesta{Lang: "go", N: n, Suma: s, Ns: ns})
}

func espera(w http.ResponseWriter, r *http.Request) {
	ms, _ := strconv.Atoi(r.URL.Query().Get("ms"))
	time.Sleep(time.Duration(ms) * time.Millisecond)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(respuesta{Lang: "go", N: ms})
}

func main() {
	http.HandleFunc("/suma", suma)
	http.HandleFunc("/espera", espera)
	// La marca de «listo» del propio proceso: main → escuchando, sin el contenedor por medio.
	fmt.Fprintf(os.Stderr, "LISTO_EN_MS %.3f\n", float64(time.Since(arranque).Microseconds())/1000)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
