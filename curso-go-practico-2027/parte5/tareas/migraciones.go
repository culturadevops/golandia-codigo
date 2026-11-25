package main

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"path"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migraciones/*.sql
var archivosMigracion embed.FS

// Migrar aplica, en orden, los .sql que todavía no estén en la tabla migraciones.
// Los archivos van dentro del binario (go:embed): no hay que copiarlos al servidor.
func Migrar(ctx context.Context, pool *pgxpool.Pool) ([]string, error) {
	_, err := pool.Exec(ctx, `CREATE TABLE IF NOT EXISTS migraciones (
		nombre      text PRIMARY KEY,
		aplicada_en timestamptz NOT NULL DEFAULT now())`)
	if err != nil {
		return nil, err
	}
	archivos, err := fs.Glob(archivosMigracion, "migraciones/*.sql")
	if err != nil {
		return nil, err
	}
	aplicadas := []string{}
	for _, archivo := range archivos {
		nombre := path.Base(archivo)
		var hecha bool
		err := pool.QueryRow(ctx,
			`SELECT EXISTS (SELECT 1 FROM migraciones WHERE nombre = $1)`, nombre).Scan(&hecha)
		if err != nil {
			return aplicadas, err
		}
		if hecha {
			continue
		}
		sql, err := archivosMigracion.ReadFile(archivo)
		if err != nil {
			return aplicadas, err
		}
		// El .sql y su registro van en la misma transacción: o las dos cosas o ninguna.
		tx, err := pool.Begin(ctx)
		if err != nil {
			return aplicadas, err
		}
		if _, err := tx.Exec(ctx, string(sql)); err != nil {
			tx.Rollback(ctx)
			return aplicadas, fmt.Errorf("%s: %w", nombre, err)
		}
		if _, err := tx.Exec(ctx, `INSERT INTO migraciones (nombre) VALUES ($1)`, nombre); err != nil {
			tx.Rollback(ctx)
			return aplicadas, err
		}
		if err := tx.Commit(ctx); err != nil {
			return aplicadas, err
		}
		aplicadas = append(aplicadas, nombre)
	}
	return aplicadas, nil
}
