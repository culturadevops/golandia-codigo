package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"pruebas/banco"
)

func main() {
	// Tres microservicios, con latencias distintas como en la vida real.
	u := banco.Nuevo(60*time.Millisecond, banco.UsuarioDemo())
	m := banco.Nuevo(90*time.Millisecond, banco.MovimientosDemo(20))
	a := banco.Nuevo(45*time.Millisecond, banco.AvisosDemo(6))
	defer u.Close()
	defer m.Close()
	defer a.Close()

	banco.Reset()
	inicio := time.Now()

	// El movil hace las tres llamadas. Y las hace EN SERIE, porque necesita el usuario
	// para saber si pedir los movimientos: es el caso normal, no el ideal.
	total := 0
	for _, url := range []string{u.URL, m.URL, a.URL} {
		r, err := http.Get(url)
		if err != nil {
			fmt.Println("error:", err)
			return
		}
		b, _ := io.ReadAll(r.Body)
		r.Body.Close()
		total += len(b)
	}

	tardo := time.Since(inicio)
	fmt.Println("SIN BFF - el cliente lo arma todo")
	fmt.Printf("  llamadas de red que hace el movil : %d\n", banco.Llamadas.Load())
	fmt.Printf("  tiempo hasta poder pintar         : %v\n", tardo.Round(time.Millisecond))
	fmt.Printf("  bytes que baja el movil           : %d\n", total)

	// Y de todo eso, ¿cuanto usa la pantalla de verdad?
	var usr banco.Usuario
	rr, _ := http.Get(u.URL)
	json.NewDecoder(rr.Body).Decode(&usr)
	rr.Body.Close()
	utiles, _ := json.Marshal(map[string]any{"nombre": usr.Nombre, "plan": usr.Plan})
	fmt.Printf("  de los cuales la pantalla usa     : %d bytes del usuario (nombre y plan)\n",
		len(utiles))
}
