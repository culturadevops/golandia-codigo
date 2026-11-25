package usuarios

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

const imagenPG = "postgres:17-alpine"

// probarStore es UNA batería de pruebas que se corre contra los dos almacenes.
func probarStore(t *testing.T, s Store) {
	ctx := context.Background()
	t.Run("crear y contar", func(t *testing.T) {
		if err := s.Crear(ctx, "ana@ejemplo.com"); err != nil {
			t.Fatal(err)
		}
		if n, _ := s.Contar(ctx); n != 1 {
			t.Fatalf("Contar = %d, quiero 1", n)
		}
	})
	t.Run("duplicado", func(t *testing.T) {
		if err := s.Crear(ctx, "ana@ejemplo.com"); !errors.Is(err, ErrDuplicado) {
			t.Fatalf("err = %v, quiero ErrDuplicado", err)
		}
	})
	t.Run("correo sin arroba", func(t *testing.T) {
		if err := s.Crear(ctx, "sin-arroba"); err == nil {
			t.Fatal("aceptó un correo sin arroba")
		}
	})
}

// arrancarPG levanta un Postgres con el módulo de testcontainers y dice cuánto tardó.
func arrancarPG(t testing.TB) (*PG, func()) {
	t.Helper()
	ctx := context.Background()
	t0 := time.Now()
	c, err := postgres.Run(ctx, imagenPG,
		postgres.WithDatabase("app"), postgres.WithUsername("app"), postgres.WithPassword("app"),
		postgres.BasicWaitStrategies())
	if err != nil {
		t.Fatal(err)
	}
	arranque := time.Since(t0)
	dsn, err := c.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	pg, err := NuevoPG(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("contenedor listo en %v", arranque.Round(10*time.Millisecond))
	return pg, func() {
		pg.Cerrar()
		_ = testcontainers.TerminateContainer(c)
	}
}

func TestMem(t *testing.T) {
	probarStore(t, NuevoMem())
}

func TestPG(t *testing.T) {
	pg, fin := arrancarPG(t)
	defer fin()
	probarStore(t, pg)
}

// ── Un contenedor por prueba contra uno compartido ──────────────────────────

func TestCincoPorPrueba(t *testing.T) {
	for i := 1; i <= 5; i++ {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			pg, fin := arrancarPG(t)
			defer fin()
			if err := pg.Crear(context.Background(), fmt.Sprintf("u%d@ejemplo.com", i)); err != nil {
				t.Fatal(err)
			}
		})
	}
}

var (
	compartido     *PG
	compartidoOnce sync.Once
)

func TestCincoCompartido(t *testing.T) {
	var fin func()
	compartidoOnce.Do(func() { compartido, fin = arrancarPG(t) })
	defer fin()
	for i := 1; i <= 5; i++ {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			if err := compartido.Crear(context.Background(), fmt.Sprintf("u%d@ejemplo.com", i)); err != nil {
				t.Fatal(err)
			}
		})
	}
}

// ── La trampa de la espera ──────────────────────────────────────────────────
//
// Postgres escribe «database system is ready to accept connections» DOS veces al arrancar por
// primera vez: una con el servidor temporal del script de inicio, y otra con el definitivo. Si
// esperas a la primera, te conectas a un servidor que se va a apagar.

func TestEsperaUnaVez(t *testing.T) { probarEspera(t, 1) }
func TestEsperaDosVeces(t *testing.T) { probarEspera(t, 2) }

func probarEspera(t *testing.T, veces int) {
	ctx := context.Background()
	fallos := 0
	const intentos = 5
	for i := 0; i < intentos; i++ {
		c, err := testcontainers.Run(ctx, imagenPG,
			testcontainers.WithEnv(map[string]string{"POSTGRES_PASSWORD": "app"}),
			testcontainers.WithExposedPorts("5432/tcp"),
			testcontainers.WithWaitStrategy(
				wait.ForLog("database system is ready to accept connections").
					WithOccurrence(veces).WithStartupTimeout(60*time.Second)))
		if err != nil {
			t.Fatal(err)
		}
		host, _ := c.Host(ctx)
		port, _ := c.MappedPort(ctx, "5432/tcp")
		dsn := fmt.Sprintf("postgres://postgres:app@%s:%s/postgres?sslmode=disable", host, port.Port())
		conn, err := pgx.Connect(ctx, dsn)
		if err == nil {
			_, err = conn.Exec(ctx, "SELECT 1")
			conn.Close(ctx)
		}
		if err != nil {
			fallos++
			t.Logf("intento %d: %v", i+1, err)
		}
		_ = testcontainers.TerminateContainer(c)
	}
	t.Logf("esperando el log %d vez/veces: %d de %d conexiones fallaron", veces, fallos, intentos)
}
