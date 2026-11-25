package pedidos_test

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"

	"tienda/internal/adaptadores/memoria"
	"tienda/internal/adaptadores/postgres"
	"tienda/internal/pedidos"
)

// Las MISMAS reglas, contra cualquier adaptador del puerto.
func probarReglas(t *testing.T, repo pedidos.Repositorio) {
	s := pedidos.NuevoServicio(repo)
	ctx := context.Background()
	casos := []struct {
		nombre            string
		cliente           string
		unidades, precio  int
		err               error
	}{
		{"válido", "Ana", 3, 4500, nil},
		{"sin cliente", "", 3, 4500, pedidos.ErrSinCliente},
		{"sin unidades", "Ana", 0, 4500, pedidos.ErrSinUnidades},
		{"demasiado", "Ana", 1000, 4500, pedidos.ErrDemasiado},
	}
	for _, c := range casos {
		p, err := s.Crear(ctx, c.cliente, c.unidades, c.precio)
		if !errors.Is(err, c.err) {
			t.Fatalf("%s: error %v, quería %v", c.nombre, err, c.err)
		}
		if err == nil {
			got, err := s.Buscar(ctx, p.ID)
			if err != nil || got.Total() != 13500 {
				t.Fatalf("%s: buscar %+v %v", c.nombre, got, err)
			}
		}
	}
}

func TestReglasEnMemoria(t *testing.T) {
	for i := 0; i < 1000; i++ {
		probarReglas(t, memoria.Nuevo())
	}
}

func TestReglasEnPostgres(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("sin DATABASE_URL")
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	for i := 0; i < 1000; i++ {
		probarReglas(t, postgres.Nuevo(db))
	}
}
