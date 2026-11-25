// banco es el servidor de pruebas del curso: endpoints que responden bien, mal y tarde.
//
//	go run ./banco -addr 127.0.0.1:8080
//
//	/ok          200 al momento
//	/roto        500 al momento
//	/lento?ms=N  200 tras N milisegundos (3000 si no dices nada); se corta si el cliente se va
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
)

func main() {
	addr := flag.String("addr", "127.0.0.1:8080", "dirección donde escuchar")
	flag.Parse()

	http.HandleFunc("/ok", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "ok")
	})
	http.HandleFunc("/roto", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "roto", http.StatusInternalServerError)
	})
	http.HandleFunc("/lento", func(w http.ResponseWriter, r *http.Request) {
		ms, err := strconv.Atoi(r.URL.Query().Get("ms"))
		if err != nil {
			ms = 3000
		}
		select {
		case <-time.After(time.Duration(ms) * time.Millisecond):
			fmt.Fprintln(w, "ok, pero tarde")
		case <-r.Context().Done():
		}
	})
	log.Printf("banco escuchando en %s", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
