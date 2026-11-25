// handler: un custom handler de Azure Functions escrito con net/http y nada más.
//
// El host de Functions (un proceso .NET) arranca este binario, le pasa el puerto en FUNCTIONS_CUSTOMHANDLER_PORT y le
// reenvía cada invocación por HTTP:
//   - /api/hola: disparador HTTP con enableForwardingHttpRequest=true → llega la petición tal cual;
//   - /procesar: disparador de cola (Azurite) → llega un POST con {"Data": {...}, "Metadata": {...}} y se responde
//     {"Outputs": {...}, "Logs": [...]}; lo que va en Outputs.factura lo escribe el host como blob (binding de salida).
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync/atomic"
	"time"
)

var (
	arranque     = time.Now()
	invocaciones atomic.Int64
)

func hola(w http.ResponseWriter, r *http.Request) {
	n := invocaciones.Add(1)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"hola":         r.URL.Query().Get("nombre"),
		"pid":          os.Getpid(),
		"invocacion":   n,
		"vivo_desde_s": int(time.Since(arranque).Seconds()),
		"metodo":       r.Method,
		"ruta":         r.URL.Path,
	})
}

type invocacion struct {
	Data     map[string]json.RawMessage `json:"Data"`
	Metadata map[string]json.RawMessage `json:"Metadata"`
}

func procesar(w http.ResponseWriter, r *http.Request) {
	var inv invocacion
	if err := json.NewDecoder(r.Body).Decode(&inv); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// El mensaje de la cola llega en Data.pedido (el "name" del binding). Si era JSON, llega como objeto.
	var pedido struct {
		ID    string  `json:"id"`
		Total float64 `json:"total"`
	}
	raw := inv.Data["pedido"]
	// El host manda el mensaje como texto JSON… con el JSON del mensaje dentro, otra vez como texto: se desenvuelve
	// mientras siga siendo una cadena (la corrida del banco enseña cuántas capas trae).
	capas, cur := 0, []byte(raw)
	for {
		var s string
		if json.Unmarshal(cur, &s) != nil {
			break
		}
		capas++
		cur = []byte(s)
	}
	if err := json.Unmarshal(cur, &pedido); err != nil {
		http.Error(w, "pedido ilegible: "+err.Error(), http.StatusBadRequest)
		return
	}
	fmt.Fprintf(os.Stderr, "[handler] capas de texto quitadas: %d → id=%s total=%.2f\n", capas, pedido.ID, pedido.Total)
	claves := make([]string, 0, len(inv.Metadata))
	for k := range inv.Metadata {
		claves = append(claves, k)
	}
	fmt.Fprintf(os.Stderr, "[handler] /procesar Data.pedido=%s · claves de Metadata=%v\n", string(raw), claves)
	factura := map[string]any{"pedido": pedido.ID, "total": pedido.Total, "iva": pedido.Total * 0.16, "hecha_por_pid": os.Getpid()}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"Outputs": map[string]any{"factura": factura},
		"Logs":    []string{"factura del pedido " + pedido.ID},
	})
}

func main() {
	puerto := os.Getenv("FUNCTIONS_CUSTOMHANDLER_PORT")
	if puerto == "" {
		puerto = "8080"
	}
	http.HandleFunc("/api/hola", hola)
	http.HandleFunc("/procesar", procesar)
	log.Printf("[handler] escuchando en :%s (pid %d)", puerto, os.Getpid())
	log.Fatal(http.ListenAndServe(":"+puerto, nil))
}
