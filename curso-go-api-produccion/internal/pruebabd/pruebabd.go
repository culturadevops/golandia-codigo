// Package pruebabd prepara un Postgres de verdad para los tests de integración.
//
// Variables: TEST_BD_URL (usuario app) y TEST_BD_URL_MIGRAR (usuario migrador), contra la base
// tareas_test. Si faltan, el test se salta... salvo en CI (CI=true), donde falla: un CI en
// verde con los tests de integración saltados es el peor resultado posible.
package pruebabd

import (
	"context"
	"io"
	"os"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"tareas/internal/bd"
)

// Abrir migra la base de tests, la vacía y devuelve un pool con el usuario de la app.
func Abrir(t testing.TB) *pgxpool.Pool {
	t.Helper()
	app, mig := os.Getenv("TEST_BD_URL"), os.Getenv("TEST_BD_URL_MIGRAR")
	if app == "" || mig == "" {
		if os.Getenv("CI") != "" {
			t.Fatal("CI sin TEST_BD_URL / TEST_BD_URL_MIGRAR: los tests de integración no pueden saltarse")
		}
		t.Skip("sin TEST_BD_URL: test de integración saltado")
	}
	ctx := context.Background()
	candado(t, mig)
	if err := bd.Migrar(ctx, mig, io.Discard); err != nil {
		t.Fatalf("migrar: %v", err)
	}
	vaciar(t, mig)
	pool, err := bd.Abrir(ctx, app)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	return pool
}

// candado: go test ./... corre los paquetes EN PARALELO, y todos usan la misma base. Sin
// esto, el TRUNCATE de un paquete borra las filas que otro está comprobando. El candado
// de Postgres los pone en fila; se suelta al acabar el test.
func candado(t testing.TB, dsn string) {
	t.Helper()
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := conn.Exec(ctx, `SELECT pg_advisory_lock(685)`); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close(ctx) }) // cerrar la conexión suelta el candado
}

func vaciar(t testing.TB, dsn string) {
	t.Helper()
	pool, err := bd.Abrir(context.Background(), dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	if _, err := pool.Exec(context.Background(),
		`TRUNCATE tareas, sesiones, usuarios RESTART IDENTITY`); err != nil {
		t.Fatal(err)
	}
}
