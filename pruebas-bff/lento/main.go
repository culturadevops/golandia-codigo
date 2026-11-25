package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"time"

	"pruebas/banco"
)

// El coste REAL del BFF no es el salto de red: es que ACOPLA la pantalla entera
// al servicio mas lento. Aqui avisos tarda 2 segundos.
func main() {
	u := banco.Nuevo(60*time.Millisecond, banco.UsuarioDemo())
	m := banco.Nuevo(90*time.Millisecond, banco.MovimientosDemo(20))
	a := banco.Nuevo(2000*time.Millisecond, banco.AvisosDemo(6))
	defer u.Close()
	defer m.Close()
	defer a.Close()

	pedir := func(ctx context.Context, url string, destino any) error {
		req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		defer res.Body.Close()
		return json.NewDecoder(res.Body).Decode(destino)
	}

	// BFF INGENUO: espera a los tres. Un servicio lento tumba la pantalla entera.
	ingenuo := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var wg sync.WaitGroup
		var usr banco.Usuario
		var movs []banco.Movimiento
		var avs []banco.Aviso
		wg.Add(3)
		go func() { defer wg.Done(); pedir(r.Context(), u.URL, &usr) }()
		go func() { defer wg.Done(); pedir(r.Context(), m.URL, &movs) }()
		go func() { defer wg.Done(); pedir(r.Context(), a.URL, &avs) }()
		wg.Wait()
		json.NewEncoder(w).Encode(map[string]any{"nombre": usr.Nombre, "avisos": len(avs)})
	}))
	defer ingenuo.Close()

	// BFF CON PLAZO POR DEPENDENCIA: lo que no llega a tiempo se devuelve como ausente.
	// La pantalla se pinta igual y el usuario ve un hueco, no una rueda infinita.
	sensato := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var wg sync.WaitGroup
		var usr banco.Usuario
		var movs []banco.Movimiento
		var avs []banco.Aviso
		avisosOK := true
		ctx, cancel := context.WithTimeout(r.Context(), 250*time.Millisecond)
		defer cancel()
		wg.Add(3)
		go func() { defer wg.Done(); pedir(ctx, u.URL, &usr) }()
		go func() { defer wg.Done(); pedir(ctx, m.URL, &movs) }()
		go func() {
			defer wg.Done()
			if err := pedir(ctx, a.URL, &avs); err != nil {
				avisosOK = false
			}
		}()
		wg.Wait()
		json.NewEncoder(w).Encode(map[string]any{
			"nombre": usr.Nombre, "movimientos": len(movs),
			"avisos_disponibles": avisosOK,
		})
	}))
	defer sensato.Close()

	medir := func(nombre, url string) {
		t := time.Now()
		res, err := http.Get(url)
		if err != nil {
			fmt.Println("  error:", err)
			return
		}
		b, _ := io.ReadAll(res.Body)
		res.Body.Close()
		fmt.Printf("  %-26s %-7v  %s", nombre, time.Since(t).Round(time.Millisecond), string(b))
	}

	fmt.Println("UN SERVICIO LENTO (avisos tarda 2s) - aqui SI duele")
	medir("BFF ingenuo:", ingenuo.URL)
	medir("BFF con plazo (250ms):", sensato.URL)
	fmt.Println()
	fmt.Println("El BFF ingenuo acopla la pantalla entera al servicio mas lento.")
	fmt.Println("Sin BFF, el movil habria pintado lo demas y dejado ese hueco cargando.")
}
