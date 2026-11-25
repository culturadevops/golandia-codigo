// Un servidor HTTP mínimo para tener conexiones que observar.
package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { fmt.Fprintln(w, "hola") })
	http.ListenAndServe(":8080", nil)
}
