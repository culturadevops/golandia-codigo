// Package memoria es el FAKE del repositorio: una implementación de verdad, pero en memoria.
// No es un mock: no sabe qué llamadas se esperan, solo guarda y devuelve pedidos.
package memoria

import (
	"context"
	"slices"
	"sync"

	"curso-testing/tienda"
)

type Repo struct {
	mu      sync.Mutex
	sig     int64
	pedidos map[int64]tienda.Pedido
}

func Nuevo() *Repo { return &Repo{pedidos: map[int64]tienda.Pedido{}} }

func (r *Repo) Guardar(_ context.Context, p *tienda.Pedido) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if p.ID == 0 {
		r.sig++
		p.ID = r.sig
	} else if _, ok := r.pedidos[p.ID]; !ok {
		return tienda.ErrNoExiste
	}
	c := *p
	c.Items = slices.Clone(p.Items) // una copia: que el llamador no cambie lo guardado
	r.pedidos[p.ID] = c
	return nil
}

func (r *Repo) Buscar(_ context.Context, id int64) (*tienda.Pedido, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	p, ok := r.pedidos[id]
	if !ok {
		return nil, tienda.ErrNoExiste
	}
	p.Items = slices.Clone(p.Items)
	return &p, nil
}

func (r *Repo) PorCliente(_ context.Context, cliente string) ([]tienda.Pedido, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out []tienda.Pedido
	for _, p := range r.pedidos {
		if p.Cliente == cliente {
			p.Items = slices.Clone(p.Items)
			out = append(out, p)
		}
	}
	slices.SortFunc(out, func(a, b tienda.Pedido) int { return int(a.ID - b.ID) })
	return out, nil
}

// Cobrador es el fake de la pasarela: aprueba todo salvo a los clientes de Rechazar.
type Cobrador struct {
	mu       sync.Mutex
	Rechazar map[string]bool
	Cobros   []tienda.Monto
}

func (c *Cobrador) Cobrar(_ context.Context, cliente string, m tienda.Monto) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.Rechazar[cliente] {
		return "", ErrRechazado
	}
	c.Cobros = append(c.Cobros, m)
	return "cob-" + cliente, nil
}

var ErrRechazado = errRechazado("tarjeta rechazada")

type errRechazado string

func (e errRechazado) Error() string { return string(e) }
