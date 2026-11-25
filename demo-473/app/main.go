package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/google/uuid"
)

func main() {
	puerto := os.Getenv("PORT")
	if puerto == "" {
		puerto = "8080"
	}
	http.HandleFunc("GET /salud", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"estado": "ok", "peticion": uuid.NewString()[:8]})
	})
	log.Fatal(http.ListenAndServe(":"+puerto, nil))
}
