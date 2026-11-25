// Package conexion: el DSN sale de una variable de entorno y se hace Ping al arrancar.
package conexion

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib" // registra el driver "pgx" para database/sql
)

func Abrir(ctx context.Context) (*sql.DB, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return nil, errors.New("falta DATABASE_URL")
	}
	db, err := sql.Open("pgx", dsn) // no conecta todavía: solo prepara el pool
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(30 * time.Minute)

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil { // aquí sí se conecta
		db.Close()
		return nil, fmt.Errorf("no llegó a la base: %w", err)
	}
	return db, nil
}
