package tareas

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Postgres guarda las tareas. Cada consulta filtra por usuario_id: si se te olvida en una,
// el usuario B lee las tareas del A con solo cambiar el número de la URL, y no da error.
type Postgres struct{ BD *pgxpool.Pool }

const columnas = `id, titulo, prioridad, hecha, creada`

func escanear(row pgx.Row) (Tarea, error) {
	var t Tarea
	err := row.Scan(&t.ID, &t.Titulo, &t.Prioridad, &t.Hecha, &t.Creada)
	if errors.Is(err, pgx.ErrNoRows) {
		return t, ErrNoExiste
	}
	return t, err
}

func (p Postgres) Crear(ctx context.Context, u int64, n Nueva) (Tarea, error) {
	return escanear(p.BD.QueryRow(ctx, `INSERT INTO tareas (usuario_id, titulo, prioridad)
		VALUES ($1, $2, $3) RETURNING `+columnas, u, n.Titulo, n.Prioridad))
}

func (p Postgres) Obtener(ctx context.Context, u, id int64) (Tarea, error) {
	return escanear(p.BD.QueryRow(ctx,
		`SELECT `+columnas+` FROM tareas WHERE id = $1 AND usuario_id = $2`, id, u))
}

func (p Postgres) Listar(ctx context.Context, u int64, f Filtro) ([]Tarea, error) {
	rows, err := p.BD.Query(ctx, `SELECT `+columnas+` FROM tareas
		WHERE usuario_id = $1 AND id > $2 AND ($3::boolean IS NULL OR hecha = $3)
		ORDER BY id LIMIT $4`, u, f.Despues, f.Hecha, f.Limite)
	if err != nil {
		return nil, err
	}
	// CollectRows cierra rows y mira rows.Err(): la lista cortada a medias no sale como buena.
	return pgx.CollectRows(rows, func(r pgx.CollectableRow) (Tarea, error) { return escanear(r) })
}

func (p Postgres) Cambiar(ctx context.Context, u, id int64, c Cambios) (Tarea, error) {
	return escanear(p.BD.QueryRow(ctx, `UPDATE tareas SET
		titulo = coalesce($3, titulo), prioridad = coalesce($4, prioridad), hecha = coalesce($5, hecha)
		WHERE id = $1 AND usuario_id = $2 RETURNING `+columnas, id, u, c.Titulo, c.Prioridad, c.Hecha))
}

func (p Postgres) Borrar(ctx context.Context, u, id int64) error {
	tag, err := p.BD.Exec(ctx, `DELETE FROM tareas WHERE id = $1 AND usuario_id = $2`, id, u)
	if err == nil && tag.RowsAffected() == 0 {
		return ErrNoExiste
	}
	return err
}
