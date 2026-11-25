package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"io"
)

// El UNICO sitio donde recover tiene sentido: el borde.
func recuperar(siguiente http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if p := recover(); p != nil {
				fmt.Println("   [middleware] atrapado:", p)
				http.Error(w, "error interno", http.StatusInternalServerError)
			}
		}()
		siguiente.ServeHTTP(w, r)
	})
}

func main() {
	roto := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var m map[string]int
		m["boom"] = 1 // panic: escribir en un mapa nil (el del episodio 2)
	})

	srv := httptest.NewServer(recuperar(roto))
	defer srv.Close()

	res, err := http.Get(srv.URL)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	defer res.Body.Close()
	cuerpo, _ := io.ReadAll(res.Body)
	fmt.Println("el cliente recibio:", res.StatusCode, string(cuerpo))
	fmt.Println("y el servidor SIGUE VIVO. Sin el recover, se caia entero.")
}
