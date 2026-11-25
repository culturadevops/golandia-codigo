// Lección 7: validación y errores JSON. El mismo cuerpo, contra un handler ingenuo y contra
// la plantilla.
//
//	APP_BD_URL=postgres://app:app@localhost:5432/tareas go run ./lecciones/07_errores
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"tareas/internal/bd"
	"tareas/internal/cuentas"
	"tareas/internal/servidor"
	"tareas/internal/tareas"
)

// ingenuo es el handler de los tutoriales: Decode y adelante, y el error tal cual al cliente.
func ingenuo(pool *pgxpool.Pool, a tareas.Almacen, usuario int64) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /tareas", func(w http.ResponseWriter, r *http.Request) {
		var n tareas.Nueva
		if err := json.NewDecoder(r.Body).Decode(&n); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		if n.Prioridad == 0 {
			n.Prioridad = 2
		}
		t, err := a.Crear(r.Context(), usuario, n)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.WriteHeader(201)
		_ = json.NewEncoder(w).Encode(t)
	})
	mux.HandleFunc("POST /usuarios", func(w http.ResponseWriter, r *http.Request) {
		var m map[string]string
		_ = json.NewDecoder(r.Body).Decode(&m)
		_, err := pool.Exec(r.Context(), `INSERT INTO usuarios (email, clave_hash) VALUES ($1, 'x')`, m["email"])
		if err != nil {
			http.Error(w, err.Error(), 500) // el error de Postgres, tal cual, al cliente
			return
		}
		w.WriteHeader(201)
	})
	return mux
}

func pedir(url, token, ruta, cuerpo string) string {
	req, _ := http.NewRequest("POST", url+ruta, strings.NewReader(cuerpo))
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "error: " + err.Error()
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	s := strings.TrimSpace(string(b))
	s = quitar(s, `"creada":"`, `"`)
	s = quitar(s, `"request_id":"`, `"`)
	if len(s) > 150 {
		s = s[:150] + "…"
	}
	return fmt.Sprintf("%d %s", resp.StatusCode, s)
}

func quitar(s, desde, hasta string) string {
	i := strings.Index(s, desde)
	if i < 0 {
		return s
	}
	j := strings.Index(s[i+len(desde):], hasta)
	return s[:i+len(desde)] + "…" + s[i+len(desde)+j:]
}

func main() {
	ctx := context.Background()
	pool, err := bd.Abrir(ctx, os.Getenv("APP_BD_URL"))
	if err != nil {
		panic(err)
	}
	defer pool.Close()
	c := cuentas.Nuevo(pool, 10)
	email := "carla" + time.Now().Format("150405") + "@ejemplo.com"
	id, _ := c.Registrar(ctx, email, "una clave larga")
	token, _, _ := c.IniciarSesion(ctx, email, "una clave larga")

	a := tareas.Postgres{BD: pool}
	malo := httptest.NewServer(ingenuo(pool, a, id))
	defer malo.Close()
	bueno := httptest.NewServer(servidor.Rutas(servidor.Deps{
		Log: slog.New(slog.NewJSONHandler(io.Discard, nil)), Cuentas: c, Tareas: a,
		Salud: &servidor.Salud{}, Plazo: 5 * time.Second,
	}))
	defer bueno.Close()

	casos := []struct{ nombre, ruta, cuerpo string }{
		{"campo mal escrito (priorida)", "/tareas", `{"titulo":"llamar al banco","priorida":1}`},
		{"dos objetos seguidos", "/tareas", `{"titulo":"uno"}{"titulo":"dos","prioridad":3}`},
		{"cuerpo de 5 MB", "/tareas", `{"titulo":"grande","x":"` + strings.Repeat("a", 5<<20) + `"}`},
		{"título vacío y prioridad 9", "/tareas", `{"titulo":"","prioridad":9}`},
		{"correo ya registrado", "/usuarios", `{"email":"` + strings.ToUpper(email) + `","clave":"una clave larga"}`},
	}
	for _, k := range casos {
		fmt.Println(k.nombre + ":")
		fmt.Println("  ingenuo   ->", pedir(malo.URL, token, k.ruta, k.cuerpo))
		fmt.Println("  plantilla ->", pedir(bueno.URL, token, k.ruta, k.cuerpo))
	}
}
