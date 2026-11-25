// Lección 6 · El límite de 4 MB: funciona con los datos de prueba y revienta con los de verdad.
package main

import (
	"context"
	"fmt"
	"io"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	"curso-grpc/comun"
	inv "curso-grpc/gen/inventario"
)

type servidor struct {
	inv.UnimplementedInventarioServer
	ultimo string
}

func catalogo(desde, n int) []*inv.Producto {
	out := make([]*inv.Producto, 0, n)
	for i := desde; i < desde+n; i++ {
		out = append(out, &inv.Producto{
			Sku:            fmt.Sprintf("SKU-%07d", i),
			Nombre:         fmt.Sprintf("Producto %07d de la tienda de Golandia", i),
			Stock:          int32(i % 50),
			PrecioCentavos: int64(1000 + i%9000),
		})
	}
	return out
}

func (s *servidor) Listar(_ context.Context, p *inv.ListarPedido) (*inv.Catalogo, error) {
	s.ultimo = "el servidor lo envió sin error"
	return &inv.Catalogo{Productos: catalogo(0, int(p.Productos))}, nil
}

func (s *servidor) ListarPorTrozos(p *inv.ListarPedido, flujo grpc.ServerStreamingServer[inv.Catalogo]) error {
	for desde := 0; desde < int(p.Productos); desde += int(p.PorTrozo) {
		n := min(int(p.PorTrozo), int(p.Productos)-desde)
		if err := flujo.Send(&inv.Catalogo{Productos: catalogo(desde, n)}); err != nil {
			return err
		}
	}
	return nil
}

func main() {
	s := &servidor{}
	addr, srv := comun.Arrancar(func(g *grpc.Server) { inv.RegisterInventarioServer(g, s) })
	defer srv.Stop()
	c := inv.NewInventarioClient(comun.Cliente(addr))
	ctx := context.Background()

	comun.Titulo("6.1 · La misma llamada, con catálogos cada vez más grandes")
	for _, n := range []int{100, 10_000, 50_000, 100_000} {
		tam := proto.Size(&inv.Catalogo{Productos: catalogo(0, n)})
		s.ultimo = ""
		_, err := c.Listar(ctx, &inv.ListarPedido{Productos: int32(n)})
		res := "OK"
		if err != nil {
			res = status.Code(err).String()
		}
		fmt.Printf("  %7s productos · %s · %s\n", miles(n), bytes(tam), res)
	}

	comun.Titulo("6.2 · El error, entero, y quién lo ve")
	_, err := c.Listar(ctx, &inv.ListarPedido{Productos: 100_000})
	fmt.Printf("  cliente: %s\n", status.Convert(err).Message())
	fmt.Printf("  servidor: «%s». Su límite de envío por defecto es math.MaxInt32 (2 GB): no se entera.\n", s.ultimo)

	comun.Titulo("6.3 · Subir el límite (grpc.MaxCallRecvMsgSize): mueve el muro, no lo quita")
	_, err = c.Listar(ctx, &inv.ListarPedido{Productos: 100_000}, grpc.MaxCallRecvMsgSize(16<<20))
	fmt.Printf("  100.000 productos con 16 MB de límite -> %v\n", status.Code(err))
	_, err = c.Listar(ctx, &inv.ListarPedido{Productos: 400_000}, grpc.MaxCallRecvMsgSize(16<<20))
	fmt.Printf("  400.000 productos con el mismo límite -> %v\n", status.Code(err))

	comun.Titulo("6.4 · Lo correcto: un flujo, a trozos de 5.000")
	flujo, _ := c.ListarPorTrozos(ctx, &inv.ListarPedido{Productos: 400_000, PorTrozo: 5000})
	trozos, total, mayor := 0, 0, 0
	for {
		t, err := flujo.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println("  error:", status.Code(err))
			break
		}
		trozos++
		total += len(t.Productos)
		mayor = max(mayor, proto.Size(t))
	}
	fmt.Printf("  %s productos en %d trozos · el trozo más grande: %s · límite intacto (4 MB)\n",
		miles(total), trozos, bytes(mayor))
}

func bytes(n int) string {
	if n >= 1<<20 {
		return strings.Replace(fmt.Sprintf("%.1f MB", float64(n)/(1<<20)), ".", ",", 1)
	}
	return fmt.Sprintf("%d KB", n/1024)
}

func miles(n int) string {
	s := fmt.Sprint(n)
	for i := len(s) - 3; i > 0; i -= 3 {
		s = s[:i] + "." + s[i:]
	}
	return s
}
