// Lección 4 · httptest para servidores y para clientes.
//
// Este paquete tiene un fallo a propósito en Rutas: el handler está bien, la ruta no.
// La versión buena es tienda/api, que se prueba con el servidor entero (api_test.go).
package httpprueba

import (
	"encoding/json"
	"net/http"
)

// VerPedido contesta GET /pedidos/{id}.
func VerPedido(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "falta el id", http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"id": id, "estado": "pagado"})
}

// Rutas registra el handler... con la ruta en singular. La documentación (y el cliente)
// dicen /pedidos/{id}.
func Rutas() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /pedido/{id}", VerPedido)
	return mux
}
