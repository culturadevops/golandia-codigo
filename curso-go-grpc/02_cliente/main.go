// Lección 2 · El cliente: grpc.NewClient no conecta, y una llamada no espera al servidor.
package main

import (
	"context"
	"fmt"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"

	"curso-grpc/comun"
	inv "curso-grpc/gen/inventario"
)

type servidor struct {
	inv.UnimplementedInventarioServer
}

func (servidor) Consultar(_ context.Context, p *inv.ConsultarPedido) (*inv.Producto, error) {
	return &inv.Producto{Sku: p.Sku, Nombre: "Taza de Golandia", Stock: 12}, nil
}

func main() {
	addr := comun.PuertoLibre()

	comun.Titulo("2.1 · La trampa: NewClient contra un puerto donde no escucha nadie")
	conn := comun.Cliente(addr)
	fmt.Printf("  grpc.NewClient(%q) -> error: <nil> · estado: %v\n", "127.0.0.1:<puerto vacío>", conn.GetState())
	fmt.Println("  no ha intentado conectar: NewClient solo prepara. «Conectado» no significa nada todavía.")

	comun.Titulo("2.2 · La primera llamada es la que se entera, y falla al momento")
	c := inv.NewInventarioClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t0 := time.Now()
	_, err := c.Consultar(ctx, &inv.ConsultarPedido{Sku: "TAZA-01"})
	cancel()
	st := status.Convert(err)
	fmt.Printf("  Consultar -> código %v en %s, con 5 s de plazo\n", st.Code(), comun.Ms(time.Since(t0)))
	fmt.Printf("  mensaje: %s\n", st.Message())
	fmt.Printf("  estado de la conexión ahora: %v\n", conn.GetState())
	conn.Close()

	comun.Titulo("2.3 · El servidor arranca 1 s tarde: sin WaitForReady, falla")
	conn = comun.Cliente(addr)
	c = inv.NewInventarioClient(conn)
	go arrancarEn(addr, time.Second)
	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	t0 = time.Now()
	_, err = c.Consultar(ctx, &inv.ConsultarPedido{Sku: "TAZA-01"})
	cancel()
	fmt.Printf("  Consultar -> código %v en %s\n", status.Code(err), comun.Ms(time.Since(t0)))

	comun.Titulo("2.4 · Con grpc.WaitForReady(true), la llamada espera a que haya servidor")
	conn.Close()
	addr2 := comun.PuertoLibre()
	conn = comun.Cliente(addr2)
	c = inv.NewInventarioClient(conn)
	t0 = time.Now()
	arranco := make(chan time.Duration, 1)
	go func() {
		time.Sleep(time.Second)
		arranco <- time.Since(t0)
		arrancarEn(addr2, 0)
	}()
	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	p, err := c.Consultar(ctx, &inv.ConsultarPedido{Sku: "TAZA-01"}, grpc.WaitForReady(true))
	cancel()
	total := time.Since(t0)
	listo := <-arranco
	fmt.Printf("  Consultar -> código %v en %s · %s, stock %d\n", status.Code(err), comun.Ms(total), p.GetNombre(), p.GetStock())
	fmt.Printf("  el servidor estaba listo a los %s; espera de más por los reintentos del cliente (backoff): %s\n",
		comun.Ms(listo), comun.Ms(total-listo))
	fmt.Println("  útil al arrancar; y siempre con plazo, o espera para siempre.")
	conn.Close()
}

func arrancarEn(addr string, tras time.Duration) {
	time.Sleep(tras)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}
	s := grpc.NewServer()
	inv.RegisterInventarioServer(s, servidor{})
	s.Serve(lis)
}
