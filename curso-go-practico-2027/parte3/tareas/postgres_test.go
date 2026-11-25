package main

import (
	"context"
	"errors"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
)

// bdDePrueba se conecta al Postgres de TAREAS_TEST_BD, migra y vacía la tabla.
// Sin esa variable, la prueba se salta: go test ./... sigue funcionando sin Docker.
func bdDePrueba(t *testing.T) *Postgres {
	t.Helper()
	url := os.Getenv("TAREAS_TEST_BD")
	if url == "" {
		t.Skip("sin TAREAS_TEST_BD: me salto la prueba con Postgres")
	}
	ctx := context.Background()
	pool, err := ConectarPostgres(ctx, url, 4)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	if _, err := Migrar(ctx, pool); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `TRUNCATE tareas RESTART IDENTITY`); err != nil {
		t.Fatal(err)
	}
	return &Postgres{pool: pool, timeout: 2 * time.Second}
}

func TestPostgresAPI(t *testing.T) {
	bd := bdDePrueba(t)
	srv := httptest.NewServer((&API{almacen: bd}).Rutas())
	defer srv.Close()
	flujo(t, srv.URL)
}

// La validación de Go se puede saltar; el CHECK de la tabla, no.
func TestPostgresRechazaTituloLargo(t *testing.T) {
	bd := bdDePrueba(t)
	_, err := bd.Crear(context.Background(), strings.Repeat("x", 121))
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		t.Fatalf("quería un error de Postgres, llegó: %v", err)
	}
	t.Logf("Postgres dijo: %s (%s)", pgErr.Message, pgErr.Code)
}

func TestPostgresNoEncontrada(t *testing.T) {
	bd := bdDePrueba(t)
	if _, err := bd.Obtener(context.Background(), 12345); !errors.Is(err, ErrNoEncontrada) {
		t.Fatalf("error = %v, quiero ErrNoEncontrada", err)
	}
	if err := bd.Borrar(context.Background(), 12345); !errors.Is(err, ErrNoEncontrada) {
		t.Fatalf("error = %v, quiero ErrNoEncontrada", err)
	}
}
