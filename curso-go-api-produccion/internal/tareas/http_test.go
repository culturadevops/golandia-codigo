package tareas

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"tareas/internal/cuentas"
)

// enMemoria es un Almacen falso para probar los handlers sin base de datos.
type enMemoria struct {
	mu sync.Mutex
	ts map[int64]Tarea
	de map[int64]int64 // id de tarea -> usuario
}

func nuevaMemoria() *enMemoria {
	return &enMemoria{ts: map[int64]Tarea{}, de: map[int64]int64{}}
}

func (m *enMemoria) Crear(_ context.Context, u int64, n Nueva) (Tarea, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	t := Tarea{ID: int64(len(m.ts) + 1), Titulo: n.Titulo, Prioridad: n.Prioridad, Creada: time.Now()}
	m.ts[t.ID], m.de[t.ID] = t, u
	return t, nil
}

func (m *enMemoria) Obtener(_ context.Context, u, id int64) (Tarea, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if t, ok := m.ts[id]; ok && m.de[id] == u {
		return t, nil
	}
	return Tarea{}, ErrNoExiste
}

func (m *enMemoria) Listar(context.Context, int64, Filtro) ([]Tarea, error) { return nil, nil }
func (m *enMemoria) Cambiar(context.Context, int64, int64, Cambios) (Tarea, error) {
	return Tarea{}, ErrNoExiste
}
func (m *enMemoria) Borrar(context.Context, int64, int64) error { return ErrNoExiste }

// comoUsuario sustituye a la autenticación real: el usuario sale de la cabecera.
func comoUsuario(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u := int64(1)
		if r.Header.Get("X-Usuario") == "2" {
			u = 2
		}
		next.ServeHTTP(w, r.WithContext(cuentas.ConUsuario(r.Context(), u)))
	})
}

func pedir(h http.Handler, metodo, ruta, cuerpo, usuario string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(metodo, ruta, strings.NewReader(cuerpo))
	r.Header.Set("X-Usuario", usuario)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, r)
	return rec
}

func TestHandlers(t *testing.T) {
	mux := http.NewServeMux()
	Rutas(mux, nuevaMemoria(), comoUsuario)

	if rec := pedir(mux, "POST", "/tareas", `{"titulo":"  ","prioridad":7}`, "1"); rec.Code != 422 ||
		!strings.Contains(rec.Body.String(), `"titulo":"obligatorio"`) ||
		!strings.Contains(rec.Body.String(), `"prioridad":"1, 2 o 3"`) {
		t.Fatalf("validación: %d %s", rec.Code, rec.Body)
	}
	if rec := pedir(mux, "POST", "/tareas", `{"titulo":"comprar pan"}`, "1"); rec.Code != 201 ||
		!strings.Contains(rec.Body.String(), `"prioridad":2`) {
		t.Fatalf("crear: %d %s", rec.Code, rec.Body)
	}
	if rec := pedir(mux, "GET", "/tareas/1", "", "1"); rec.Code != 200 {
		t.Fatalf("la dueña: %d", rec.Code)
	}
	if rec := pedir(mux, "GET", "/tareas/1", "", "2"); rec.Code != 404 {
		t.Fatalf("otro usuario: %d, esperaba 404", rec.Code)
	}
	if rec := pedir(mux, "GET", "/tareas/abc", "", "1"); rec.Code != 404 {
		t.Fatalf("id no numérico: %d", rec.Code)
	}
}
