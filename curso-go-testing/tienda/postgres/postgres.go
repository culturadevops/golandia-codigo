// Package postgres es el repositorio de verdad, con pgx.
package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"curso-testing/tienda"
)

const Esquema = `
CREATE TABLE IF NOT EXISTS pedidos (
    id       bigserial PRIMARY KEY,
    cliente  text        NOT NULL,
    total    bigint      NOT NULL,
    estado   text        NOT NULL,
    cobro_id text        NOT NULL DEFAULT '',
    creado   timestamptz NOT NULL
);
CREATE TABLE IF NOT EXISTS items (
    pedido_id bigint NOT NULL REFERENCES pedidos(id) ON DELETE CASCADE,
    pos       int    NOT NULL,
    sku       text   NOT NULL,
    cantidad  int    NOT NULL CHECK (cantidad > 0),
    precio    bigint NOT NULL,
    PRIMARY KEY (pedido_id, pos)
);
CREATE INDEX IF NOT EXISTS pedidos_cliente ON pedidos (cliente);`

type Repo struct{ DB *pgxpool.Pool }

func Migrar(ctx context.Context, db *pgxpool.Pool) error {
	_, err := db.Exec(ctx, Esquema)
	return err
}

func (r *Repo) Guardar(ctx context.Context, p *tienda.Pedido) error {
	return pgx.BeginFunc(ctx, r.DB, func(tx pgx.Tx) error {
		if p.ID == 0 {
			err := tx.QueryRow(ctx, `INSERT INTO pedidos (cliente,total,estado,cobro_id,creado)
				VALUES ($1,$2,$3,$4,$5) RETURNING id`,
				p.Cliente, p.Total, p.Estado, p.CobroID, p.Creado).Scan(&p.ID)
			if err != nil {
				return err
			}
		} else {
			tag, err := tx.Exec(ctx, `UPDATE pedidos SET cliente=$2,total=$3,estado=$4,
				cobro_id=$5,creado=$6 WHERE id=$1`,
				p.ID, p.Cliente, p.Total, p.Estado, p.CobroID, p.Creado)
			if err != nil {
				return err
			}
			if tag.RowsAffected() == 0 {
				return tienda.ErrNoExiste
			}
			if _, err := tx.Exec(ctx, `DELETE FROM items WHERE pedido_id=$1`, p.ID); err != nil {
				return err
			}
		}
		for i, it := range p.Items {
			_, err := tx.Exec(ctx, `INSERT INTO items (pedido_id,pos,sku,cantidad,precio)
				VALUES ($1,$2,$3,$4,$5)`, p.ID, i, it.SKU, it.Cantidad, it.Precio)
			if err != nil {
				return fmt.Errorf("item %s: %w", it.SKU, err)
			}
		}
		return nil
	})
}

func (r *Repo) Buscar(ctx context.Context, id int64) (*tienda.Pedido, error) {
	var p tienda.Pedido
	err := r.DB.QueryRow(ctx, `SELECT id,cliente,total,estado,cobro_id,creado
		FROM pedidos WHERE id=$1`, id).
		Scan(&p.ID, &p.Cliente, &p.Total, &p.Estado, &p.CobroID, &p.Creado)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, tienda.ErrNoExiste
	}
	if err != nil {
		return nil, err
	}
	if p.Items, err = r.items(ctx, p.ID); err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *Repo) PorCliente(ctx context.Context, cliente string) ([]tienda.Pedido, error) {
	rows, err := r.DB.Query(ctx, `SELECT id FROM pedidos WHERE cliente=$1 ORDER BY id`, cliente)
	if err != nil {
		return nil, err
	}
	ids, err := pgx.CollectRows(rows, pgx.RowTo[int64])
	if err != nil {
		return nil, err
	}
	out := make([]tienda.Pedido, 0, len(ids))
	for _, id := range ids {
		p, err := r.Buscar(ctx, id)
		if err != nil {
			return nil, err
		}
		out = append(out, *p)
	}
	return out, nil
}

func (r *Repo) items(ctx context.Context, id int64) ([]tienda.Item, error) {
	rows, err := r.DB.Query(ctx, `SELECT sku,cantidad,precio FROM items
		WHERE pedido_id=$1 ORDER BY pos`, id)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(f pgx.CollectableRow) (tienda.Item, error) {
		var it tienda.Item
		err := f.Scan(&it.SKU, &it.Cantidad, &it.Precio)
		return it, err
	})
}
