package web

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// MaxCuerpo es el tamaño máximo de un cuerpo JSON. Una tarea cabe en 1 KB.
const MaxCuerpo = 64 << 10

// Leer decodifica UN objeto JSON del cuerpo, sin campos desconocidos y con tope de tamaño.
// Los tres detalles faltan en el json.NewDecoder(r.Body).Decode(&v) de siempre, y ninguno da
// error: el campo mal escrito se ignora, lo que viene detrás del objeto se ignora, y el cuerpo
// de 50 MB se lee entero.
func Leer(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, MaxCuerpo)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return traducir(err)
	}
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return PeticionMala("el cuerpo debe tener un solo objeto JSON")
	}
	return nil
}

func traducir(err error) error {
	var (
		sintaxis *json.SyntaxError
		tipo     *json.UnmarshalTypeError
		grande   *http.MaxBytesError
	)
	switch {
	case errors.As(err, &grande):
		return &Error{Estado: 413, Codigo: "demasiado_grande",
			Mensaje: fmt.Sprintf("el cuerpo pasa de %d bytes", grande.Limit)}
	case errors.As(err, &sintaxis), errors.Is(err, io.ErrUnexpectedEOF):
		return PeticionMala("JSON mal formado")
	case errors.As(err, &tipo):
		return PeticionMala(fmt.Sprintf("el campo %q debe ser %s", tipo.Field, tipo.Type))
	case errors.Is(err, io.EOF):
		return PeticionMala("el cuerpo está vacío")
	case strings.HasPrefix(err.Error(), "json: unknown field "):
		// encoding/json no exporta este error: el texto es lo único que hay.
		return PeticionMala("campo desconocido " + strings.TrimPrefix(err.Error(), "json: unknown field "))
	}
	return PeticionMala("JSON inválido")
}
