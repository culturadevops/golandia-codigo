package main

import (
	"context"
	"errors"
	"net"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// Cada error de pgx tiene que acabar en el código HTTP correcto.
func TestClasificar(t *testing.T) {
	casos := []struct {
		nombre string
		err    error
		quiero error // nil = ninguno de los nuestros: 500
	}{
		{"sin filas", pgx.ErrNoRows, ErrNoEncontrada},
		{"tiempo agotado", context.DeadlineExceeded, ErrBDLenta},
		{"check de la tabla", &pgconn.PgError{Code: "23514"}, nil},
		{"Postgres apagándose", &pgconn.PgError{Code: "57P01"}, ErrBDCaida},
		{"conexión rota", &pgconn.PgError{Code: "08006"}, ErrBDCaida},
		{"red", &net.OpError{Op: "dial", Err: errors.New("connection refused")}, ErrBDCaida},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			got := clasificar(c.err)
			for _, nuestro := range []error{ErrNoEncontrada, ErrBDLenta, ErrBDCaida} {
				if errors.Is(got, nuestro) != (nuestro == c.quiero) {
					t.Errorf("clasificar(%v) = %v, quiero %v", c.err, got, c.quiero)
				}
			}
		})
	}
}
