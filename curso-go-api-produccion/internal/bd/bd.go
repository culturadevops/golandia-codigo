// Package bd abre el pool de Postgres y aplica las migraciones.
package bd

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib" // driver «pgx» para database/sql, solo para goose
	"github.com/pressly/goose/v3"
	"github.com/pressly/goose/v3/lock"

	"tareas/migraciones"
)

// Abrir crea el pool y hace Ping con plazo: pgxpool.New no conecta, y una clave mal
// puesta no daría error hasta la primera petición.
func Abrir(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("la base de datos no contesta: %w", err)
	}
	return pool, nil
}

// Migrar aplica las migraciones pendientes con el usuario dueño del esquema (no el de la app).
// El candado de sesión de Postgres impide que dos réplicas migren a la vez.
func Migrar(ctx context.Context, dsn string, w io.Writer) error {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return err
	}
	defer db.Close()
	candado, err := lock.NewPostgresSessionLocker()
	if err != nil {
		return err
	}
	p, err := goose.NewProvider(goose.DialectPostgres, db, migraciones.FS, goose.WithSessionLocker(candado))
	if err != nil {
		return err
	}
	res, err := p.Up(ctx)
	for _, r := range res {
		fmt.Fprintf(w, "aplicada %s (%d ms)\n", r.Source.Path, r.Duration.Milliseconds())
	}
	if err != nil {
		return err
	}
	v, err := p.GetDBVersion(ctx)
	if err != nil {
		return err
	}
	fmt.Fprintf(w, "esquema en la versión %d (%d migraciones nuevas)\n", v, len(res))
	return nil
}

// Sonda adapta el pool a lo que pide la readiness.
type Sonda struct{ *pgxpool.Pool }

func (s Sonda) Version(ctx context.Context) (int64, error) { return VersionEsquema(ctx, s.Pool) }

// VersionEsquema lee la versión aplicada. La app solo tiene SELECT sobre la tabla de goose.
func VersionEsquema(ctx context.Context, pool *pgxpool.Pool) (int64, error) {
	var v int64
	err := pool.QueryRow(ctx,
		`SELECT coalesce(max(version_id), 0) FROM goose_db_version WHERE is_applied`).Scan(&v)
	return v, err
}
