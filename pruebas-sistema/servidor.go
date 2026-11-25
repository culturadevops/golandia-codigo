package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/salud", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "vivo")
	})
	fmt.Println("escuchando en el 8080")
	http.ListenAndServe(":8080", nil)
}
