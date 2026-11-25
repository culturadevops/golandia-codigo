// Package usuarios es el código «de producción» que probamos: un almacén de usuarios con dos
// implementaciones, una de verdad (Postgres) y un doble en memoria.
package usuarios

import (
	"context"
	"errors"
	"strings"
	"sync"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrDuplicado = errors.New("ese correo ya existe")

type Store interface {
	Crear(ctx context.Context, correo string) error
	Contar(ctx context.Context) (int, error)
}

// ── Postgres ────────────────────────────────────────────────────────────────

type PG struct{ db *pgxpool.Pool }

func NuevoPG(ctx context.Context, dsn string) (*PG, error) {
	db, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}
	_, err = db.Exec(ctx, `CREATE TABLE IF NOT EXISTS usuarios (
		id SERIAL PRIMARY KEY,
		correo TEXT NOT NULL UNIQUE CHECK (correo LIKE '%_@_%'))`)
	return &PG{db: db}, err
}

func (p *PG) Crear(ctx context.Context, correo string) error {
	_, err := p.db.Exec(ctx, `INSERT INTO usuarios (correo) VALUES ($1)`, correo)
	if err != nil && strings.Contains(err.Error(), "23505") { // unique_violation
		return ErrDuplicado
	}
	return err
}

func (p *PG) Contar(ctx context.Context) (int, error) {
	var n int
	err := p.db.QueryRow(ctx, `SELECT count(*) FROM usuarios`).Scan(&n)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, nil
	}
	return n, err
}

func (p *PG) Cerrar() { p.db.Close() }

// ── El doble en memoria ─────────────────────────────────────────────────────

// Mem imita a la tabla... casi. No sabe nada del CHECK de la columna `correo`: acepta un correo
// sin arroba. Ese es el tipo de diferencia que un doble no te avisa.
type Mem struct {
	mu     sync.Mutex
	correo map[string]bool
}

func NuevoMem() *Mem { return &Mem{correo: map[string]bool{}} }

func (m *Mem) Crear(_ context.Context, correo string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.correo[correo] {
		return ErrDuplicado
	}
	m.correo[correo] = true
	return nil
}

func (m *Mem) Contar(context.Context) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.correo), nil
}
