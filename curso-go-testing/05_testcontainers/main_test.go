//go:build integracion

// Lección 5 · testcontainers-go con un Postgres de verdad.
//
// UN contenedor para todo el paquete (TestMain), una foto de la base de datos recién migrada
// (Snapshot) y cada test empieza restaurándola (Restore): aislados sin pagar un arranque por test.
//
//	go test -tags integracion -v ./05_testcontainers
package contenedores

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib" // registra «pgx» en database/sql: Restore lo usa
	"github.com/testcontainers/testcontainers-go"
	tcpg "github.com/testcontainers/testcontainers-go/modules/postgres"

	"curso-testing/tienda/postgres"
)

const Imagen = "postgres:18.6-alpine"

var pg *tcpg.PostgresContainer

func arrancar(ctx context.Context, extra ...testcontainers.ContainerCustomizer) (*tcpg.PostgresContainer, error) {
	opts := append([]testcontainers.ContainerCustomizer{
		tcpg.WithDatabase("tienda"), tcpg.WithUsername("curso"), tcpg.WithPassword("curso"),
		tcpg.BasicWaitStrategies(), tcpg.WithSQLDriver("pgx"),
	}, extra...)
	return tcpg.Run(ctx, Imagen, opts...)
}

func TestMain(m *testing.M) {
	ctx := context.Background()
	t0 := time.Now()
	c, err := arrancar(ctx)
	if err != nil {
		fmt.Println("no pude arrancar Postgres:", err)
		os.Exit(1)
	}
	pg = c
	fmt.Printf("arranque del contenedor (%s): %.2f s\n", Imagen, time.Since(t0).Seconds())

	dsn := pg.MustConnectionString(ctx, "sslmode=disable")
	db, err := pgxpool.New(ctx, dsn)
	if err == nil {
		err = postgres.Migrar(ctx, db)
		db.Close()
	}
	if err == nil {
		t0 = time.Now()
		err = pg.Snapshot(ctx, tcpg.WithSnapshotName("recien_migrada"))
		fmt.Printf("Snapshot de la base migrada: %.0f ms\n", ms(time.Since(t0)))
	}
	codigo := 1
	if err != nil {
		fmt.Println("preparando la base:", err)
	} else {
		codigo = m.Run()
	}
	testcontainers.TerminateContainer(pg)
	os.Exit(codigo)
}

func ms(d time.Duration) float64 { return float64(d.Microseconds()) / 1000 }

// repo da a cada test un repositorio sobre la base recién migrada, y la restaura al terminar.
func repo(t *testing.T) *postgres.Repo {
	t.Helper()
	ctx := context.Background()
	t.Cleanup(func() { // se registra primero: corre DESPUÉS de cerrar el pool
		t0 := time.Now()
		if err := pg.Restore(ctx); err != nil {
			t.Errorf("Restore: %v", err)
		}
		t.Logf("Restore: %.0f ms", ms(time.Since(t0)))
	})
	db, err := pgxpool.New(ctx, pg.MustConnectionString(ctx, "sslmode=disable"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(db.Close)
	return &postgres.Repo{DB: db}
}
