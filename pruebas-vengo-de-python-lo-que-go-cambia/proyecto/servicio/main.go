// El servicio más pequeño posible: responde «ok» en /salud. Solo biblioteca estándar.
package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/salud", func(w http.ResponseWriter, r *http.Request) { fmt.Fprintln(w, "ok") })
	http.ListenAndServe(":8080", nil)
}
