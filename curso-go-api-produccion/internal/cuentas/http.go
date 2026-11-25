package cuentas

import (
	"context"
	"net/http"
	"strings"
	"time"

	"tareas/internal/web"
)

type claveUsuario struct{}

// UsuarioDe devuelve el id del usuario autenticado. Solo vale detrás de Requerir.
func UsuarioDe(ctx context.Context) int64 {
	id, _ := ctx.Value(claveUsuario{}).(int64)
	return id
}

// ConUsuario mete un usuario en el contexto. Para los tests de los handlers.
func ConUsuario(ctx context.Context, id int64) context.Context {
	return context.WithValue(ctx, claveUsuario{}, id)
}

func tokenDe(r *http.Request) string {
	t, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
	if !ok {
		return ""
	}
	return strings.TrimSpace(t)
}

// Requerir deja pasar solo peticiones con un token de sesión vigente.
func (s *Servicio) Requerir(next http.Handler) http.Handler {
	return web.Manejador(func(w http.ResponseWriter, r *http.Request) error {
		t := tokenDe(r)
		if t == "" {
			return web.NoAutenticado("falta Authorization: Bearer <token>")
		}
		id, err := s.Usuario(r.Context(), t)
		if err != nil {
			return err
		}
		next.ServeHTTP(w, r.WithContext(ConUsuario(r.Context(), id)))
		return nil
	})
}

type credenciales struct {
	Email string `json:"email"`
	Clave string `json:"clave"`
}

// Rutas registra las rutas de cuentas.
func (s *Servicio) Rutas(mux *http.ServeMux) {
	mux.Handle("POST /usuarios", web.Manejador(s.registrar))
	mux.Handle("POST /sesiones", web.Manejador(s.entrar))
	mux.Handle("DELETE /sesiones", s.Requerir(web.Manejador(s.salir)))
}

func (s *Servicio) registrar(w http.ResponseWriter, r *http.Request) error {
	var c credenciales
	if err := web.Leer(w, r, &c); err != nil {
		return err
	}
	id, err := s.Registrar(r.Context(), c.Email, c.Clave)
	if err != nil {
		return err
	}
	web.JSON(w, http.StatusCreated, map[string]any{"id": id, "email": Normalizar(c.Email)})
	return nil
}

func (s *Servicio) entrar(w http.ResponseWriter, r *http.Request) error {
	var c credenciales
	if err := web.Leer(w, r, &c); err != nil {
		return err
	}
	token, vence, err := s.IniciarSesion(r.Context(), c.Email, c.Clave)
	if err != nil {
		return err
	}
	web.JSON(w, http.StatusCreated, map[string]any{"token": token, "vence": vence.Format(time.RFC3339)})
	return nil
}

func (s *Servicio) salir(w http.ResponseWriter, r *http.Request) error {
	if err := s.Cerrar(r.Context(), tokenDe(r)); err != nil {
		return err
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}
