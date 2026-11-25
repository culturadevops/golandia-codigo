package main

import (
	"log"
	"net/http"
)

func main() {
	api := &API{almacen: NuevaMemoria()}
	log.Println("escuchando en :8080")
	log.Fatal(http.ListenAndServe(":8080", api.Rutas()))
}
