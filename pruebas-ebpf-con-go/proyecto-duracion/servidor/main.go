// Un servidor HTTP que tarda lo que le pidas: /?ms=300 responde a los 300 ms.
package main

import (
	"fmt"
	"net/http"
	"strconv"
	"time"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		ms, _ := strconv.Atoi(r.URL.Query().Get("ms"))
		time.Sleep(time.Duration(ms) * time.Millisecond)
		fmt.Fprintln(w, "hola")
	})
	http.ListenAndServe(":8080", nil)
}
