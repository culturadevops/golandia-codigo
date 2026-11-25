package main

import (
	"cmp"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// sondearSalud es el healthcheck del contenedor: la imagen es scratch y no trae curl,
// así que el propio binario pregunta a GET /salud y sale con 0 (sana) o 1 (no sana).
//
//	tareas salud
func sondearSalud() int {
	puerto := cmp.Or(os.Getenv("TAREAS_PUERTO"), "8080")
	cliente := http.Client{Timeout: 2 * time.Second}
	resp, err := cliente.Get("http://127.0.0.1:" + puerto + "/salud")
	if err != nil {
		fmt.Println("salud:", err)
		return 1
	}
	defer resp.Body.Close()
	cuerpo, _ := io.ReadAll(resp.Body)
	fmt.Printf("salud: %d %s", resp.StatusCode, cuerpo)
	if resp.StatusCode != http.StatusOK {
		return 1
	}
	return 0
}
