// Lección 3 · Plazos: sin deadline se espera para siempre; y el servidor tiene que mirarlo.
package main

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"

	"curso-grpc/comun"
	inv "curso-grpc/gen/inventario"
)

// servidor hace 10 pasos de 100 ms (una consulta lenta a la base de datos, por ejemplo).
type servidor struct {
	inv.UnimplementedInventarioServer
	colgado    chan struct{} // si no es nil, el handler se queda esperando aquí
	respeta    bool          // ¿mira el ctx entre paso y paso?
	pasos      atomic.Int32
	plazoVisto atomic.Int64 // lo que le queda al plazo cuando entra la llamada, en ms
}

func (s *servidor) Consultar(ctx context.Context, p *inv.ConsultarPedido) (*inv.Producto, error) {
	if dl, ok := ctx.Deadline(); ok {
		s.plazoVisto.Store(time.Until(dl).Milliseconds())
	} else {
		s.plazoVisto.Store(-1)
	}
	if s.colgado != nil {
		<-s.colgado
		return &inv.Producto{Sku: p.Sku}, nil
	}
	for i := 0; i < 10; i++ {
		if s.respeta && ctx.Err() != nil {
			return nil, status.FromContextError(ctx.Err()).Err()
		}
		time.Sleep(100 * time.Millisecond)
		s.pasos.Add(1)
	}
	return &inv.Producto{Sku: p.Sku, Stock: 12}, nil
}

func levantar(s *servidor) (inv.InventarioClient, func()) {
	addr, srv := comun.Arrancar(func(g *grpc.Server) { inv.RegisterInventarioServer(g, s) })
	conn := comun.Cliente(addr)
	return inv.NewInventarioClient(conn), func() { conn.Close(); srv.Stop() }
}

func main() {
	comun.Titulo("3.1 · La trampa: sin plazo, contra un servidor colgado")
	colgado := &servidor{colgado: make(chan struct{})}
	c, cerrar := levantar(colgado)
	hecho := make(chan error, 1)
	t0 := time.Now()
	go func() {
		_, err := c.Consultar(context.Background(), &inv.ConsultarPedido{Sku: "TAZA-01"})
		hecho <- err
	}()
	select {
	case err := <-hecho:
		fmt.Printf("  respondió: %v\n", err)
	case <-time.After(3 * time.Second):
		fmt.Printf("  a los %s: sin respuesta y sin error. gRPC no pone plazo por defecto.\n", comun.Ms(time.Since(t0)))
		fmt.Printf("  el servidor vio la llamada sin plazo: %v\n", colgado.plazoVisto.Load() == -1)
	}
	close(colgado.colgado)
	<-hecho
	cerrar()

	comun.Titulo("3.2 · Con plazo: context.WithTimeout de 300 ms")
	lento := &servidor{respeta: true}
	c, cerrar = levantar(lento)
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	t0 = time.Now()
	_, err := c.Consultar(ctx, &inv.ConsultarPedido{Sku: "TAZA-01"})
	cancel()
	fmt.Printf("  Consultar -> código %v en %s\n", status.Code(err), comun.Ms(time.Since(t0)))
	fmt.Printf("  el plazo viaja con la llamada: al servidor le llegó con %d ms por delante\n", lento.plazoVisto.Load())
	time.Sleep(1200 * time.Millisecond)
	fmt.Printf("  y como el servidor mira el ctx, paró en el paso %d de 10\n", lento.pasos.Load())
	cerrar()

	comun.Titulo("3.3 · La otra trampa: el servidor que no mira el ctx")
	sordo := &servidor{respeta: false}
	c, cerrar = levantar(sordo)
	ctx, cancel = context.WithTimeout(context.Background(), 300*time.Millisecond)
	t0 = time.Now()
	_, err = c.Consultar(ctx, &inv.ConsultarPedido{Sku: "TAZA-01"})
	cancel()
	fmt.Printf("  el cliente se rinde: código %v en %s\n", status.Code(err), comun.Ms(time.Since(t0)))
	time.Sleep(1200 * time.Millisecond)
	fmt.Printf("  el servidor siguió trabajando para nadie: %d pasos de 10\n", sordo.pasos.Load())
	cerrar()
}
