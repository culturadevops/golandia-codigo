// Package tareas es el dominio de la API: el modelo, sus reglas y dónde se guarda.
package tareas

import (
	"context"
	"strings"
	"time"
	"unicode/utf8"

	"tareas/internal/web"
)

// Tarea es lo que se guarda y lo que se devuelve.
type Tarea struct {
	ID        int64     `json:"id"`
	Titulo    string    `json:"titulo"`
	Prioridad int       `json:"prioridad"`
	Hecha     bool      `json:"hecha"`
	Creada    time.Time `json:"creada"`
}

// Nueva es el cuerpo de POST /tareas.
type Nueva struct {
	Titulo    string `json:"titulo"`
	Prioridad int    `json:"prioridad"`
}

// Cambios es el cuerpo de PATCH /tareas/{id}: lo que no viene (nil) no se toca.
type Cambios struct {
	Titulo    *string `json:"titulo"`
	Prioridad *int    `json:"prioridad"`
	Hecha     *bool   `json:"hecha"`
}

// ErrNoExiste: la tarea no existe, o es de otro usuario. Para el cliente es lo mismo.
var ErrNoExiste = web.NoEncontrado("esa tarea no existe")

// Validar normaliza y devuelve todos los problemas a la vez.
func (n *Nueva) Validar() error {
	if n.Prioridad == 0 {
		n.Prioridad = 2
	}
	c := map[string]string{}
	revisarTitulo(&n.Titulo, c)
	revisarPrioridad(n.Prioridad, c)
	return web.Validacion(c)
}

// Validar revisa solo los campos que vienen.
func (c *Cambios) Validar() error {
	m := map[string]string{}
	if c.Titulo == nil && c.Prioridad == nil && c.Hecha == nil {
		return web.PeticionMala("no hay nada que cambiar")
	}
	if c.Titulo != nil {
		revisarTitulo(c.Titulo, m)
	}
	if c.Prioridad != nil {
		revisarPrioridad(*c.Prioridad, m)
	}
	return web.Validacion(m)
}

func revisarTitulo(t *string, m map[string]string) {
	*t = strings.TrimSpace(*t)
	switch n := utf8.RuneCountInString(*t); {
	case n == 0:
		m["titulo"] = "obligatorio"
	case n > 200:
		m["titulo"] = "máximo 200 caracteres"
	}
}

func revisarPrioridad(p int, m map[string]string) {
	if p < 1 || p > 3 {
		m["prioridad"] = "1, 2 o 3"
	}
}

// Filtro para listar: pagina por id («después de»), nunca con OFFSET.
type Filtro struct {
	Hecha   *bool
	Despues int64
	Limite  int
}

// Almacen es lo que los handlers necesitan de la base de datos. Todas las operaciones
// llevan el usuario: no hay forma de pedir una tarea sin decir de quién es.
type Almacen interface {
	Crear(ctx context.Context, usuario int64, n Nueva) (Tarea, error)
	Obtener(ctx context.Context, usuario, id int64) (Tarea, error)
	Listar(ctx context.Context, usuario int64, f Filtro) ([]Tarea, error)
	Cambiar(ctx context.Context, usuario, id int64, c Cambios) (Tarea, error)
	Borrar(ctx context.Context, usuario, id int64) error
}
