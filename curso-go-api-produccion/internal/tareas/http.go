package tareas

import (
	"net/http"
	"strconv"

	"tareas/internal/cuentas"
	"tareas/internal/web"
)

// Rutas registra las rutas de tareas, todas detrás de auth.
func Rutas(mux *http.ServeMux, a Almacen, auth func(http.Handler) http.Handler) {
	h := handlers{a}
	mux.Handle("GET /tareas", auth(web.Manejador(h.listar)))
	mux.Handle("POST /tareas", auth(web.Manejador(h.crear)))
	mux.Handle("GET /tareas/{id}", auth(web.Manejador(h.obtener)))
	mux.Handle("PATCH /tareas/{id}", auth(web.Manejador(h.cambiar)))
	mux.Handle("DELETE /tareas/{id}", auth(web.Manejador(h.borrar)))
}

type handlers struct{ a Almacen }

func id(r *http.Request) (int64, error) {
	n, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || n < 1 {
		return 0, ErrNoExiste // /tareas/abc no es un error del servidor
	}
	return n, nil
}

func (h handlers) crear(w http.ResponseWriter, r *http.Request) error {
	var n Nueva
	if err := web.Leer(w, r, &n); err != nil {
		return err
	}
	if err := n.Validar(); err != nil {
		return err
	}
	t, err := h.a.Crear(r.Context(), cuentas.UsuarioDe(r.Context()), n)
	if err != nil {
		return err
	}
	web.JSON(w, http.StatusCreated, t)
	return nil
}

func (h handlers) obtener(w http.ResponseWriter, r *http.Request) error {
	n, err := id(r)
	if err != nil {
		return err
	}
	t, err := h.a.Obtener(r.Context(), cuentas.UsuarioDe(r.Context()), n)
	if err != nil {
		return err
	}
	web.JSON(w, http.StatusOK, t)
	return nil
}

func (h handlers) listar(w http.ResponseWriter, r *http.Request) error {
	f, err := filtro(r)
	if err != nil {
		return err
	}
	ts, err := h.a.Listar(r.Context(), cuentas.UsuarioDe(r.Context()), f)
	if err != nil {
		return err
	}
	resp := map[string]any{"tareas": ts}
	if len(ts) == f.Limite {
		resp["siguiente"] = ts[len(ts)-1].ID // el cliente pide ?despues=<siguiente>
	}
	web.JSON(w, http.StatusOK, resp)
	return nil
}

func filtro(r *http.Request) (Filtro, error) {
	q := r.URL.Query()
	f := Filtro{Limite: 50}
	m := map[string]string{}
	if v := q.Get("limite"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 || n > 100 {
			m["limite"] = "de 1 a 100"
		}
		f.Limite = n
	}
	if v := q.Get("despues"); v != "" {
		n, err := strconv.ParseInt(v, 10, 64)
		if err != nil || n < 0 {
			m["despues"] = "un id"
		}
		f.Despues = n
	}
	if v := q.Get("hecha"); v != "" {
		b, err := strconv.ParseBool(v)
		if err != nil {
			m["hecha"] = "true o false"
		}
		f.Hecha = &b
	}
	return f, web.Validacion(m)
}

func (h handlers) cambiar(w http.ResponseWriter, r *http.Request) error {
	n, err := id(r)
	if err != nil {
		return err
	}
	var c Cambios
	if err := web.Leer(w, r, &c); err != nil {
		return err
	}
	if err := c.Validar(); err != nil {
		return err
	}
	t, err := h.a.Cambiar(r.Context(), cuentas.UsuarioDe(r.Context()), n, c)
	if err != nil {
		return err
	}
	web.JSON(w, http.StatusOK, t)
	return nil
}

func (h handlers) borrar(w http.ResponseWriter, r *http.Request) error {
	n, err := id(r)
	if err != nil {
		return err
	}
	if err := h.a.Borrar(r.Context(), cuentas.UsuarioDe(r.Context()), n); err != nil {
		return err
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}
