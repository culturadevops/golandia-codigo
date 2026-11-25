package pedidos

import (
	"errors"
	"fmt"
	"sync"
)

// Pool es un pool de conexiones con tope, como el de database/sql. Aquel espera turno
// cuando no quedan libres; este devuelve un error, para que la prueba no se cuelgue.
type Pool struct {
	mu     sync.Mutex
	libres int
	tope   int
}

type Conn struct{ p *Pool }

var ErrSinConexiones = errors.New("no quedan conexiones libres")

func NuevoPool(n int) *Pool { return &Pool{libres: n, tope: n} }

func (p *Pool) Tomar() (*Conn, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.libres == 0 {
		return nil, fmt.Errorf("%w (%d de %d en uso)", ErrSinConexiones, p.tope, p.tope)
	}
	p.libres--
	return &Conn{p: p}, nil
}

func (c *Conn) Soltar() {
	c.p.mu.Lock()
	c.p.libres++
	c.p.mu.Unlock()
}

// Escribir manda el pedido por la conexión (aquí no hace nada).
func (c *Conn) Escribir(id int) { _ = id }
