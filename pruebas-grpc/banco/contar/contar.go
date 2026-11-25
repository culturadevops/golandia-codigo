package contar

// Cuenta los bytes que pasan POR EL SOCKET, no los del cuerpo serializado.
//
// Es la unica forma honesta de comparar: si contaramos solo el JSON contra solo el protobuf,
// dejariamos fuera las cabeceras de HTTP/1.1 y el marco de HTTP/2 con su HPACK, que es
// justamente donde gRPC se gasta parte de lo que ahorra.

import (
	"net"
	"sync/atomic"
)

type Contador struct {
	Enviados  atomic.Int64
	Recibidos atomic.Int64
}

func (c *Contador) Cero() { c.Enviados.Store(0); c.Recibidos.Store(0) }

func (c *Contador) Total() int64 { return c.Enviados.Load() + c.Recibidos.Load() }

type marcada struct {
	net.Conn
	c *Contador
}

func (m marcada) Read(p []byte) (int, error) {
	n, err := m.Conn.Read(p)
	m.c.Recibidos.Add(int64(n))
	return n, err
}

func (m marcada) Write(p []byte) (int, error) {
	n, err := m.Conn.Write(p)
	m.c.Enviados.Add(int64(n))
	return n, err
}

func (c *Contador) Marcar(base net.Conn) net.Conn { return marcada{base, c} }
