// Lección 8 · Apagar sin cortar: Stop contra GracefulStop, el health check y un interceptor de registro.
package main

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"

	"curso-grpc/comun"
	inv "curso-grpc/gen/inventario"
)

type servidor struct {
	inv.UnimplementedInventarioServer
}

func (servidor) Consultar(ctx context.Context, p *inv.ConsultarPedido) (*inv.Producto, error) {
	select {
	case <-time.After(time.Duration(p.RetrasoMs) * time.Millisecond):
		return &inv.Producto{Sku: p.Sku, Stock: 12}, nil
	case <-ctx.Done():
		return nil, status.FromContextError(ctx.Err()).Err()
	}
}

var (
	registroMu sync.Mutex
	registro   []string
)

// registrar es un interceptor: envuelve TODAS las llamadas unarias sin tocar los handlers.
func registrar(ctx context.Context, req any, info *grpc.UnaryServerInfo, h grpc.UnaryHandler) (any, error) {
	t0 := time.Now()
	resp, err := h(ctx, req)
	metodo := info.FullMethod[strings.LastIndex(info.FullMethod, "/")+1:]
	registroMu.Lock()
	registro = append(registro, fmt.Sprintf("%s · %v · %s", metodo, status.Code(err), comun.Ms(time.Since(t0))))
	registroMu.Unlock()
	return resp, err
}

// cincoEnVuelo lanza 5 llamadas de 500 ms, apaga el servidor a los 100 ms y cuenta cómo acaban.
func cincoEnVuelo(apagar func(*grpc.Server)) string {
	addr, srv := comun.Arrancar(func(g *grpc.Server) { inv.RegisterInventarioServer(g, servidor{}) })
	c := inv.NewInventarioClient(comun.Cliente(addr))
	var wg sync.WaitGroup
	var mu sync.Mutex
	cuenta := map[string]int{}
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := c.Consultar(context.Background(), &inv.ConsultarPedido{Sku: "TAZA-01", RetrasoMs: 500})
			mu.Lock()
			cuenta[status.Code(err).String()]++
			mu.Unlock()
		}()
	}
	time.Sleep(100 * time.Millisecond)
	t0 := time.Now()
	apagar(srv)
	tarda := time.Since(t0)
	wg.Wait()
	_, err := c.Consultar(context.Background(), &inv.ConsultarPedido{Sku: "TAZA-01"})
	partes := []string{}
	for k, v := range cuenta {
		partes = append(partes, fmt.Sprintf("%d %s", v, k))
	}
	return fmt.Sprintf("las 5 en vuelo: %s · apagar tardó %s · una llamada nueva después: %v",
		strings.Join(partes, ", "), comun.Ms(tarda), status.Code(err))
}

func main() {
	comun.Titulo("8.1 · La trampa: srv.Stop() con 5 llamadas a medias")
	fmt.Println("  " + cincoEnVuelo(func(s *grpc.Server) { s.Stop() }))

	comun.Titulo("8.2 · Lo correcto: srv.GracefulStop()")
	fmt.Println("  " + cincoEnVuelo(func(s *grpc.Server) { s.GracefulStop() }))
	fmt.Println("  termina lo que tiene a medias y no acepta nada nuevo. Con un tope: si tarda, Stop().")

	comun.Titulo("8.3 · El health check: avisar ANTES de apagar")
	sano := health.NewServer()
	addr, srv := comun.Arrancar(func(g *grpc.Server) {
		inv.RegisterInventarioServer(g, servidor{})
		grpc_health_v1.RegisterHealthServer(g, sano)
	})
	hc := grpc_health_v1.NewHealthClient(comun.Cliente(addr))
	r, _ := hc.Check(context.Background(), &grpc_health_v1.HealthCheckRequest{})
	fmt.Printf("  arrancado: %v\n", r.GetStatus())
	sano.Shutdown()
	r, _ = hc.Check(context.Background(), &grpc_health_v1.HealthCheckRequest{})
	fmt.Printf("  tras sano.Shutdown(): %v · el balanceador deja de mandar, y entonces GracefulStop\n", r.GetStatus())
	srv.GracefulStop()

	comun.Titulo("8.4 · Un interceptor: el registro de cada llamada, sin tocar un solo handler")
	addr, srv = comun.Arrancar(func(g *grpc.Server) { inv.RegisterInventarioServer(g, servidor{}) },
		grpc.UnaryInterceptor(registrar))
	c := inv.NewInventarioClient(comun.Cliente(addr))
	c.Consultar(context.Background(), &inv.ConsultarPedido{Sku: "TAZA-01", RetrasoMs: 20})
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	c.Consultar(ctx, &inv.ConsultarPedido{Sku: "TAZA-01", RetrasoMs: 200})
	cancel()
	time.Sleep(20 * time.Millisecond)
	c.Listar(context.Background(), &inv.ListarPedido{})
	srv.GracefulStop()
	for _, l := range registro {
		fmt.Println("  " + l)
	}
}
