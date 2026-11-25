package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"time"

	"pruebas/banco"
)

// El caso honesto: UN solo backend. Aqui el BFF no agrega nada, solo reenvia.
func main() {
	backend := banco.Nuevo(60*time.Millisecond, banco.UsuarioDemo())
	defer backend.Close()

	bff := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		res, err := http.Get(backend.URL)
		if err != nil {
			http.Error(w, "arriba", 502)
			return
		}
		defer res.Body.Close()
		var usr banco.Usuario
		json.NewDecoder(res.Body).Decode(&usr)
		json.NewEncoder(w).Encode(map[string]any{"nombre": usr.Nombre, "plan": usr.Plan})
	}))
	defer bff.Close()

	medir := func(nombre, url string, vueltas int) time.Duration {
		var total time.Duration
		for i := 0; i < vueltas; i++ {
			t := time.Now()
			res, _ := http.Get(url)
			io.Copy(io.Discard, res.Body)
			res.Body.Close()
			total += time.Since(t)
		}
		media := total / time.Duration(vueltas)
		fmt.Printf("  %-22s %v de media (%d vueltas)\n", nombre, media.Round(100*time.Microsecond), vueltas)
		return media
	}

	fmt.Println("UN SOLO SERVICIO DETRAS - el BFF no agrega nada")
	directo := medir("directo al backend:", backend.URL, 20)
	conbff := medir("pasando por el BFF:", bff.URL, 20)

	fmt.Println()
	fmt.Printf("  el BFF anade %v por peticion (%.0f%% mas lento)\n",
		(conbff - directo).Round(100*time.Microsecond),
		float64(conbff-directo)/float64(directo)*100)
	fmt.Println("  Con un solo backend, un BFF es un salto de red de mas.")
}
