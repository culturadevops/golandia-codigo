package consultas_test

import (
	"context"
	"database/sql"
	"os"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"tienda"
	"tienda/consultas"
	"tienda/db"
)

// Test de integración: corre contra una base REAL (la de tests, no la de la app).
// Sin TEST_DATABASE_URL se salta.
func abrirTest(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("sin TEST_DATABASE_URL")
	}
	sqldb, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { sqldb.Close() })
	if err := sqldb.Ping(); err != nil {
		t.Fatal(err)
	}
	goose.SetBaseFS(tienda.Migraciones)
	goose.SetLogger(goose.NopLogger())
	if err := goose.SetDialect("postgres"); err != nil {
		t.Fatal(err)
	}
	if err := goose.Up(sqldb, "migrations"); err != nil { // las MISMAS migraciones que producción
		t.Fatal(err)
	}
	// base limpia en cada corrida
	_, err = sqldb.Exec(`TRUNCATE productos RESTART IDENTITY;
		INSERT INTO productos (nombre, categoria, precio_centavos) VALUES
		('Café molido 500 g', 'cafe', 12900), ('Café en grano 1 kg', 'cafe', 24900),
		('Taza', 'cocina', 8900), ('Café soluble', 'cafe', 9900)`)
	if err != nil {
		t.Fatal(err)
	}
	return sqldb
}

func TestLasTresFormasDanLoMismo(t *testing.T) {
	sqldb := abrirTest(t)
	ctx := context.Background()

	a, err := consultas.PorCategoriaString(ctx, sqldb, "cafe", 2)
	if err != nil {
		t.Fatal(err)
	}
	b, err := consultas.PorCategoriaEmbed(ctx, sqldb, "cafe", 2)
	if err != nil {
		t.Fatal(err)
	}
	c, err := db.New(sqldb).PorCategoria(ctx, db.PorCategoriaParams{Categoria: "cafe", Limit: 2})
	if err != nil {
		t.Fatal(err)
	}
	if len(a) != 2 || len(b) != 2 || len(c) != 2 {
		t.Fatalf("filas: a=%d b=%d c=%d, quería 2", len(a), len(b), len(c))
	}
	for i := range a {
		if a[i].Nombre != b[i].Nombre || a[i].Nombre != c[i].Nombre {
			t.Errorf("fila %d distinta: %q %q %q", i, a[i].Nombre, b[i].Nombre, c[i].Nombre)
		}
	}
	if a[0].Nombre != "Café soluble" {
		t.Errorf("el más barato: %q, quería Café soluble", a[0].Nombre)
	}
}
