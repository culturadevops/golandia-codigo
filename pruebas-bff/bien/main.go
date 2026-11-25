package main

import (
	"encoding/json"
	"net/http"

	"pruebas/api"
)

type Servidor struct{}

// La firma la impone la interfaz generada del openapi.yaml. Si no coincide, no compila.
func (s Servidor) GetPantalla(w http.ResponseWriter, r *http.Request) {
	saldo := -248.60
	json.NewEncoder(w).Encode(api.PantallaMovil{
		Nombre: "Ana Rivas", Plan: "premium", Saldo: saldo, SinLeer: 3,
	})
}

var _ api.ServerInterface = Servidor{}

func main() {}
