// Package postgres es un ADAPTADOR del puerto pedidos.Repositorio sobre Postgres (database/sql + pgx).
package postgres

import (
	"context"
	"database/sql"
	"errors"

	"tienda/internal/pedidos"
)

type Repo struct{ db *sql.DB }

func Nuevo(db *sql.DB) *Repo { return &Repo{db: db} }

func (r *Repo) Guardar(ctx context.Context, p pedidos.Pedido) (int64, error) {
	var id int64
	err := r.db.QueryRowContext(ctx, "INSERT INTO pedidos (cliente, unidades, precio) VALUES ($1, $2, $3) RETURNING id",
		p.Cliente, p.Unidades, p.Precio).Scan(&id)
	return id, err
}

func (r *Repo) Buscar(ctx context.Context, id int64) (pedidos.Pedido, error) {
	p := pedidos.Pedido{ID: id}
	err := r.db.QueryRowContext(ctx, "SELECT cliente, unidades, precio FROM pedidos WHERE id = $1", id).
		Scan(&p.Cliente, &p.Unidades, &p.Precio)
	if errors.Is(err, sql.ErrNoRows) {
		return pedidos.Pedido{}, pedidos.ErrNoEncontrado
	}
	return p, err
}
