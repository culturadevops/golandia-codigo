// Lección 5 · Streaming: Send se bloquea si el cliente no lee, y un flujo sin cancelar no termina.
package main

import (
	"context"
	"fmt"
	"io"
	"sync/atomic"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"

	"curso-grpc/comun"
	inv "curso-grpc/gen/inventario"
)

type servidor struct {
	inv.UnimplementedInventarioServer
	enviados atomic.Int32
	vivos    atomic.Int32
	fin      chan error
}

func (s *servidor) Vigilar(p *inv.VigilarPedido, flujo grpc.ServerStreamingServer[inv.Evento]) error {
	s.vivos.Add(1)
	defer s.vivos.Add(-1)
	relleno := make([]byte, p.BytesPorEvento)
	for i := int32(0); i < p.Eventos; i++ {
		if err := flujo.Send(&inv.Evento{N: i, Relleno: relleno}); err != nil {
			s.fin <- err
			return err
		}
		s.enviados.Store(i + 1)
	}
	s.fin <- nil
	return nil
}

func main() {
	s := &servidor{fin: make(chan error, 4)}
	addr, srv := comun.Arrancar(func(g *grpc.Server) { inv.RegisterInventarioServer(g, s) })
	defer srv.Stop()
	c := inv.NewInventarioClient(comun.Cliente(addr))

	comun.Titulo("5.1 · Un flujo del servidor: 20 eventos pequeños, leídos hasta io.EOF")
	flujo, _ := c.Vigilar(context.Background(), &inv.VigilarPedido{Eventos: 20, BytesPorEvento: 10})
	n := 0
	for {
		_, err := flujo.Recv()
		if err == io.EOF {
			break
		}
		n++
	}
	<-s.fin
	fmt.Printf("  recibidos: %d · el final del flujo es io.EOF, no un error\n", n)

	comun.Titulo("5.2 · El cliente pide 1.000 eventos de 16 KB y no lee")
	s.enviados.Store(0)
	ctx, cancel := context.WithCancel(context.Background())
	flujo, _ = c.Vigilar(ctx, &inv.VigilarPedido{Eventos: 1000, BytesPorEvento: 16 * 1024})
	time.Sleep(time.Second)
	a := s.enviados.Load()
	fmt.Printf("  tras 1 s, el servidor lleva enviados %d de 1.000 (unos %d KB) y su Send está bloqueado\n", a, int(a)*16)
	time.Sleep(time.Second)
	fmt.Printf("  tras 2 s, sigue en %d: es el control de flujo de HTTP/2, no un error\n", s.enviados.Load())
	for i := 0; i < 50; i++ {
		flujo.Recv()
	}
	time.Sleep(200 * time.Millisecond)
	fmt.Printf("  el cliente lee 50 -> el servidor avanza a %d\n", s.enviados.Load())

	comun.Titulo("5.3 · La trampa: el cliente deja de leer y no cancela")
	time.Sleep(3 * time.Second)
	fmt.Printf("  3 s después: handlers vivos en el servidor = %d, parado en %d\n", s.vivos.Load(), s.enviados.Load())
	fmt.Println("  ese handler (con su gorrutina y sus buffers) sigue ahí mientras la conexión viva: una fuga por cada cliente así")

	comun.Titulo("5.4 · Lo correcto: cancelar el contexto (defer cancel())")
	t0 := time.Now()
	cancel()
	err := <-s.fin
	tarda := time.Since(t0)
	time.Sleep(50 * time.Millisecond)
	fmt.Printf("  cancel() -> el Send del servidor devuelve %v en %s · handlers vivos = %d\n",
		status.Code(err), comun.Ms(tarda), s.vivos.Load())
}
