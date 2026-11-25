package tareas_test

import (
	"context"
	"errors"
	"testing"

	"tareas/internal/cuentas"
	"tareas/internal/pruebabd"
	"tareas/internal/tareas"
)

func TestPostgres(t *testing.T) {
	pool := pruebabd.Abrir(t)
	ctx := context.Background()
	c := cuentas.Nuevo(pool, 10)
	ana, err := c.Registrar(ctx, "ana@ejemplo.com", "una clave larga")
	if err != nil {
		t.Fatal(err)
	}
	beto, _ := c.Registrar(ctx, "beto@ejemplo.com", "otra clave larga")

	a := tareas.Postgres{BD: pool}
	var ultima tareas.Tarea
	for _, ti := range []string{"uno", "dos", "tres"} {
		if ultima, err = a.Crear(ctx, ana, tareas.Nueva{Titulo: ti, Prioridad: 2}); err != nil {
			t.Fatal(err)
		}
	}

	pag, err := a.Listar(ctx, ana, tareas.Filtro{Limite: 2})
	if err != nil || len(pag) != 2 {
		t.Fatalf("página 1: %v %v", pag, err)
	}
	resto, _ := a.Listar(ctx, ana, tareas.Filtro{Limite: 2, Despues: pag[1].ID})
	if len(resto) != 1 || resto[0].Titulo != "tres" {
		t.Fatalf("página 2: %v", resto)
	}

	if _, err := a.Obtener(ctx, beto, ultima.ID); !errors.Is(err, tareas.ErrNoExiste) {
		t.Fatalf("beto ve la tarea de ana: %v", err)
	}
	if err := a.Borrar(ctx, beto, ultima.ID); !errors.Is(err, tareas.ErrNoExiste) {
		t.Fatalf("beto borra la tarea de ana: %v", err)
	}

	hecha := true
	if got, err := a.Cambiar(ctx, ana, ultima.ID, tareas.Cambios{Hecha: &hecha}); err != nil || !got.Hecha || got.Titulo != "tres" {
		t.Fatalf("cambiar: %+v %v", got, err)
	}
	solo, _ := a.Listar(ctx, ana, tareas.Filtro{Limite: 10, Hecha: &hecha})
	if len(solo) != 1 {
		t.Fatalf("filtro hecha: %v", solo)
	}
}
