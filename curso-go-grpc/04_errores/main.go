// Lección 4 · Errores: el error de Go no cruza el cable, y un pánico tumba el servidor entero.
package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime/debug"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"curso-grpc/comun"
	inv "curso-grpc/gen/inventario"
)

var ErrNoExiste = errors.New("no existe")

type servidor struct {
	inv.UnimplementedInventarioServer
	modo string
}

func (s servidor) Consultar(_ context.Context, p *inv.ConsultarPedido) (*inv.Producto, error) {
	switch {
	case p.Sku == "TAZA-01":
		return &inv.Producto{Sku: p.Sku, Nombre: "Taza de Golandia", Stock: 12}, nil
	case s.modo == "fmt":
		return nil, fmt.Errorf("sku %s: %w", p.Sku, ErrNoExiste)
	case s.modo == "status":
		return nil, status.Errorf(codes.NotFound, "sku %s: no existe", p.Sku)
	case s.modo == "nilnil":
		return nil, nil
	case s.modo == "panico":
		var precios map[string]int64
		precios[p.Sku] = 0 // escribir en un mapa nil: panic
	}
	return nil, nil
}

func main() {
	if os.Getenv("L04_HIJO") != "" {
		hijo(os.Getenv("L04_HIJO") == "recuperar")
		return
	}

	comun.Titulo("4.1 · La trampa: devolver un error de Go con fmt.Errorf")
	c, cerrar := levantar("fmt")
	_, err := c.Consultar(context.Background(), &inv.ConsultarPedido{Sku: "PLATO-99"})
	fmt.Printf("  el cliente recibe: código %v · %q\n", status.Code(err), status.Convert(err).Message())
	fmt.Printf("  errors.Is(err, ErrNoExiste): %v · el %%w se queda en el servidor\n", errors.Is(err, ErrNoExiste))
	cerrar()

	comun.Titulo("4.2 · Lo correcto: status.Errorf con su código")
	c, cerrar = levantar("status")
	_, err = c.Consultar(context.Background(), &inv.ConsultarPedido{Sku: "PLATO-99"})
	fmt.Printf("  el cliente recibe: código %v · %q\n", status.Code(err), status.Convert(err).Message())
	fmt.Printf("  y decide por el código: status.Code(err) == codes.NotFound -> %v\n", status.Code(err) == codes.NotFound)
	cerrar()

	comun.Titulo("4.3 · Devolver nil, nil")
	c, cerrar = levantar("nilnil")
	p, err := c.Consultar(context.Background(), &inv.ConsultarPedido{Sku: "PLATO-99"})
	fmt.Printf("  el cliente recibe: código %v · producto nil: %v · sku=%q stock=%d\n",
		status.Code(err), p == nil, p.GetSku(), p.GetStock())
	fmt.Println("  -> sin error y con un producto vacío: el cliente cree que existe y tiene 0 en stock")
	cerrar()

	comun.Titulo("4.4 · La trampa gorda: un pánico en un handler")
	for _, l := range conHijo("panico") {
		fmt.Println("  " + l)
	}

	comun.Titulo("4.5 · Lo correcto: un interceptor que recupera")
	for _, l := range conHijo("recuperar") {
		fmt.Println("  " + l)
	}
}

func levantar(modo string) (inv.InventarioClient, func()) {
	addr, srv := comun.Arrancar(func(g *grpc.Server) { inv.RegisterInventarioServer(g, servidor{modo: modo}) })
	conn := comun.Cliente(addr)
	return inv.NewInventarioClient(conn), func() { conn.Close(); srv.Stop() }
}

// hijo es el servidor en su propio proceso, para poder ver si muere.
func hijo(recuperar bool) {
	var opciones []grpc.ServerOption
	if recuperar {
		opciones = append(opciones, grpc.UnaryInterceptor(recuperador))
	}
	addr, _ := comun.Arrancar(func(g *grpc.Server) { inv.RegisterInventarioServer(g, servidor{modo: "panico"}) }, opciones...)
	fmt.Println(addr)
	select {}
}

// recuperador convierte un pánico en codes.Internal y deja el servidor vivo.
func recuperador(ctx context.Context, req any, info *grpc.UnaryServerInfo, h grpc.UnaryHandler) (resp any, err error) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "pánico en %s: %v\n%s", info.FullMethod, r, debug.Stack())
			err = status.Errorf(codes.Internal, "error interno")
		}
	}()
	return h(ctx, req)
}

// conHijo arranca el servidor en otro proceso, le hace tres llamadas y mira si sigue vivo.
func conHijo(modo string) []string {
	cmd := exec.Command(os.Args[0])
	cmd.Env = append(os.Environ(), "L04_HIJO="+modo)
	salida, _ := cmd.StdoutPipe()
	var errBuf strings.Builder
	cmd.Stderr = &errBuf
	cmd.Start()
	addr, _ := bufio.NewReader(salida).ReadString('\n')
	c := inv.NewInventarioClient(comun.Cliente(strings.TrimSpace(addr)))
	var out []string
	for _, sku := range []string{"TAZA-01", "PLATO-99", "TAZA-01"} {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		_, err := c.Consultar(ctx, &inv.ConsultarPedido{Sku: sku})
		cancel()
		out = append(out, fmt.Sprintf("Consultar(%s) -> %v", sku, status.Code(err)))
		time.Sleep(200 * time.Millisecond)
	}
	fin := make(chan error, 1)
	go func() { fin <- cmd.Wait() }()
	select {
	case <-fin:
		primera := strings.SplitN(strings.TrimSpace(errBuf.String()), "\n", 2)[0]
		out = append(out, fmt.Sprintf("el proceso del servidor terminó: código de salida %d", cmd.ProcessState.ExitCode()),
			"su última palabra: "+primera,
			"-> un sku raro tumbó el servidor, y con él a todos sus clientes")
	case <-time.After(time.Second):
		primera := strings.SplitN(strings.TrimSpace(errBuf.String()), "\n", 2)[0]
		out = append(out, "el proceso del servidor sigue vivo", "en su log: "+primera)
		cmd.Process.Kill()
		<-fin
	}
	return out
}
