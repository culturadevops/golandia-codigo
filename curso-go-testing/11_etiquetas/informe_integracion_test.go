//go:build integracion

package etiquetas

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	tcpg "github.com/testcontainers/testcontainers-go/modules/postgres"

	"curso-testing/tienda"
	"curso-testing/tienda/postgres"
)

// El mismo test contra Postgres de verdad. Solo con -tags integracion.
func TestPagadosDePostgres(t *testing.T) {
	ctx := context.Background()
	t0 := time.Now()
	c, err := tcpg.Run(ctx, "postgres:18.6-alpine", tcpg.WithDatabase("tienda"),
		tcpg.WithUsername("curso"), tcpg.WithPassword("curso"), tcpg.BasicWaitStrategies())
	testcontainers.CleanupContainer(t, c)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Postgres listo en %.2f s", time.Since(t0).Seconds())
	db, err := pgxpool.New(ctx, c.MustConnectionString(ctx, "sslmode=disable"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(db.Close)
	if err := postgres.Migrar(ctx, db); err != nil {
		t.Fatal(err)
	}
	r := &postgres.Repo{DB: db}
	for _, e := range []string{"pagado", "rechazado", "pagado"} {
		r.Guardar(ctx, &tienda.Pedido{Cliente: "ana", Estado: e, Creado: time.Now()})
	}
	if n, _ := PagadosDe(ctx, r, "ana"); n != 2 {
		t.Errorf("n = %d; quiero 2", n)
	}
}
