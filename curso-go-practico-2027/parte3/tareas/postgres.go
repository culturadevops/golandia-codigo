package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrBDLenta = errors.New("la base de datos no respondió a tiempo")
	ErrBDCaida = errors.New("base de datos no disponible")
)

// ConectarPostgres abre el pool y comprueba que Postgres conteste.
func ConectarPostgres(ctx context.Context, url string, maxConexiones int32) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, fmt.Errorf("DATABASE_URL: %w", err)
	}
	cfg.MaxConns = maxConexiones
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}
	ctxPing, cancelar := context.WithTimeout(ctx, 5*time.Second)
	defer cancelar()
	if err := pool.Ping(ctxPing); err != nil {
		pool.Close()
		return nil, fmt.Errorf("no llego a Postgres: %w", err)
	}
	return pool, nil
}

// Postgres guarda las tareas en la tabla tareas.
// Cada consulta tiene un tiempo máximo: timeout.
type Postgres struct {
	pool    *pgxpool.Pool
	timeout time.Duration
}

func (p *Postgres) Listar(ctx context.Context) ([]Tarea, error) {
	ctx, cancelar := context.WithTimeout(ctx, p.timeout)
	defer cancelar()
	filas, err := p.pool.Query(ctx, `SELECT id, titulo, hecha FROM tareas ORDER BY id`)
	if err != nil {
		return nil, clasificar(err)
	}
	lista, err := pgx.CollectRows(filas, pgx.RowToStructByPos[Tarea])
	return lista, clasificar(err)
}

func (p *Postgres) Crear(ctx context.Context, titulo string) (Tarea, error) {
	ctx, cancelar := context.WithTimeout(ctx, p.timeout)
	defer cancelar()
	var t Tarea
	err := p.pool.QueryRow(ctx,
		`INSERT INTO tareas (titulo) VALUES ($1) RETURNING id, titulo, hecha`, titulo,
	).Scan(&t.ID, &t.Titulo, &t.Hecha)
	return t, clasificar(err)
}

func (p *Postgres) Obtener(ctx context.Context, id int64) (Tarea, error) {
	ctx, cancelar := context.WithTimeout(ctx, p.timeout)
	defer cancelar()
	var t Tarea
	err := p.pool.QueryRow(ctx,
		`SELECT id, titulo, hecha FROM tareas WHERE id = $1`, id,
	).Scan(&t.ID, &t.Titulo, &t.Hecha)
	return t, clasificar(err)
}

func (p *Postgres) Marcar(ctx context.Context, id int64, hecha bool) (Tarea, error) {
	ctx, cancelar := context.WithTimeout(ctx, p.timeout)
	defer cancelar()
	var t Tarea
	err := p.pool.QueryRow(ctx,
		`UPDATE tareas SET hecha = $2 WHERE id = $1 RETURNING id, titulo, hecha`, id, hecha,
	).Scan(&t.ID, &t.Titulo, &t.Hecha)
	return t, clasificar(err)
}

func (p *Postgres) Borrar(ctx context.Context, id int64) error {
	ctx, cancelar := context.WithTimeout(ctx, p.timeout)
	defer cancelar()
	res, err := p.pool.Exec(ctx, `DELETE FROM tareas WHERE id = $1`, id)
	if err != nil {
		return clasificar(err)
	}
	if res.RowsAffected() == 0 {
		return ErrNoEncontrada
	}
	return nil
}

// Salud hace un ping corto y devuelve cómo está el pool.
func (p *Postgres) Salud(ctx context.Context) (Estado, error) {
	ctx, cancelar := context.WithTimeout(ctx, time.Second)
	defer cancelar()
	s := p.pool.Stat()
	e := Estado{BD: "ok", Abiertas: s.TotalConns(), EnUso: s.AcquiredConns(),
		Max: s.MaxConns(), Esperas: s.EmptyAcquireCount()}
	if err := p.pool.Ping(ctx); err != nil {
		e.BD = "caida"
		return e, clasificar(err)
	}
	return e, nil
}

// clasificar traduce los errores de pgx a los de la API.
func clasificar(err error) error {
	var pgErr *pgconn.PgError
	switch {
	case err == nil:
		return nil
	case errors.Is(err, pgx.ErrNoRows):
		return ErrNoEncontrada
	case errors.Is(err, context.DeadlineExceeded):
		return fmt.Errorf("%w: %w", ErrBDLenta, err)
	case errors.As(err, &pgErr):
		return err // Postgres contestó con un error: es un fallo nuestro
	default:
		return fmt.Errorf("%w: %w", ErrBDCaida, err)
	}
}
