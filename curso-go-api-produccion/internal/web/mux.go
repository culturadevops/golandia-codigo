package web

import "net/http"

// ErroresDelMux cambia el «404 page not found» y el «405 method not allowed» en texto plano
// que escribe http.ServeMux por el mismo JSON que el resto de la API. El mux no tiene un
// gancho para eso: se reconoce su respuesta (http.Error pone text/plain) y se reemplaza.
func ErroresDelMux(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(&tapaMux{ResponseWriter: w, r: r}, r)
	})
}

type tapaMux struct {
	http.ResponseWriter
	r       *http.Request
	tapando bool
}

func (t *tapaMux) WriteHeader(s int) {
	esDelMux := t.Header().Get("Content-Type") == "text/plain; charset=utf-8"
	if esDelMux && (s == http.StatusNotFound || s == http.StatusMethodNotAllowed) {
		t.tapando = true
		t.Header().Del("X-Content-Type-Options")
		e := &Error{Estado: s, Codigo: "no_encontrado", Mensaje: "no existe esa ruta"}
		if s == http.StatusMethodNotAllowed {
			e.Codigo, e.Mensaje = "metodo_no_permitido", "permitidos: "+t.Header().Get("Allow")
		}
		JSON(t.ResponseWriter, s, sobre{Error: e, RequestID: IDPeticion(t.r.Context())})
		return
	}
	t.ResponseWriter.WriteHeader(s)
}

func (t *tapaMux) Write(b []byte) (int, error) {
	if t.tapando {
		return len(b), nil // el texto del mux se descarta
	}
	return t.ResponseWriter.Write(b)
}

func (t *tapaMux) Unwrap() http.ResponseWriter { return t.ResponseWriter }
