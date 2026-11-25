// Package sonda abre una conexión TLS y lee la fecha de vencimiento del certificado del servidor.
package sonda

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"time"
)

type Resultado struct {
	Vence    time.Time     // NotAfter del certificado del servidor
	Duracion time.Duration // lo que tardó la sonda
	Err      error
}

// Sondear conecta a addr (host:puerto) y devuelve el NotAfter del certificado.
// timeout = 0 significa «sin límite» (solo para la demo: no lo hagas).
func Sondear(ctx context.Context, addr string, timeout time.Duration) Resultado {
	inicio := time.Now()
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}
	host, _, _ := net.SplitHostPort(addr)
	d := tls.Dialer{Config: &tls.Config{
		ServerName: host,
		// Queremos leer la fecha también de un certificado vencido: no validamos la cadena aquí.
		InsecureSkipVerify: true,
	}}
	conn, err := d.DialContext(ctx, "tcp", addr)
	if err != nil {
		return Resultado{Duracion: time.Since(inicio), Err: err}
	}
	defer conn.Close()
	certs := conn.(*tls.Conn).ConnectionState().PeerCertificates
	if len(certs) == 0 {
		return Resultado{Duracion: time.Since(inicio), Err: errors.New("sin certificado")}
	}
	return Resultado{Vence: certs[0].NotAfter, Duracion: time.Since(inicio)}
}
