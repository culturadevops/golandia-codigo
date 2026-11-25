package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"time"

	"pruebas/banco"
)

// Lo que la pantalla del MOVIL necesita. Ni un campo mas.
type PantallaMovil struct {
	Nombre     string   `json:"nombre"`
	Plan       string   `json:"plan"`
	Saldo      float64  `json:"saldo"`
	Ultimos    []string `json:"ultimos"`
	SinLeer    int      `json:"sin_leer"`
}

func main() {
	u := banco.Nuevo(60*time.Millisecond, banco.UsuarioDemo())
	m := banco.Nuevo(90*time.Millisecond, banco.MovimientosDemo(20))
	a := banco.Nuevo(45*time.Millisecond, banco.AvisosDemo(6))
	defer u.Close()
	defer m.Close()
	defer a.Close()

	// EL BFF. Una ruta. Dentro, las tres llamadas EN PARALELO.
	bff := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			wg   sync.WaitGroup
			usr  banco.Usuario
			movs []banco.Movimiento
			avs  []banco.Aviso
		)
		pedir := func(url string, destino any) {
			defer wg.Done()
			res, err := http.Get(url)
			if err != nil {
				return
			}
			defer res.Body.Close()
			json.NewDecoder(res.Body).Decode(destino)
		}
		wg.Add(3)
		go pedir(u.URL, &usr)
		go pedir(m.URL, &movs)
		go pedir(a.URL, &avs)
		wg.Wait()

		// Agregar y TRANSFORMAR: eso es el BFF.
		p := PantallaMovil{Nombre: usr.Nombre, Plan: usr.Plan}
		for _, x := range movs {
			p.Saldo += x.Importe
		}
		for i, x := range movs {
			if i == 3 {
				break
			}
			p.Ultimos = append(p.Ultimos, x.Concepto)
		}
		for _, x := range avs {
			if !x.Leido {
				p.SinLeer++
			}
		}
		json.NewEncoder(w).Encode(p)
	}))
	defer bff.Close()

	banco.Reset()
	inicio := time.Now()
	res, _ := http.Get(bff.URL)
	b, _ := io.ReadAll(res.Body)
	res.Body.Close()
	tardo := time.Since(inicio)

	fmt.Println("CON BFF - una llamada, y el BFF agrega")
	fmt.Printf("  llamadas de red que hace el movil : 1\n")
	fmt.Printf("  llamadas que hace el BFF por detras: %d (en paralelo)\n", banco.Llamadas.Load())
	fmt.Printf("  tiempo hasta poder pintar         : %v\n", tardo.Round(time.Millisecond))
	fmt.Printf("  bytes que baja el movil           : %d\n", len(b))
	fmt.Printf("  y esto es TODO lo que recibe      : %s\n", string(b))
}
