// Lección 9 · ¿gRPC o REST? Lo medido, y lo que no se ve con un navegador ni con curl.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os/exec"
	"sort"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"curso-grpc/comun"
	inv "curso-grpc/gen/inventario"
)

type servidor struct {
	inv.UnimplementedInventarioServer
}

func catalogo(n int) *inv.Catalogo {
	c := &inv.Catalogo{}
	for i := 0; i < n; i++ {
		c.Productos = append(c.Productos, &inv.Producto{
			Sku: fmt.Sprintf("SKU-%07d", i), Nombre: fmt.Sprintf("Producto %07d de la tienda de Golandia", i),
			Stock: int32(i % 50), PrecioCentavos: int64(1000 + i%9000),
		})
	}
	return c
}

func (servidor) Listar(_ context.Context, p *inv.ListarPedido) (*inv.Catalogo, error) {
	return catalogo(int(p.Productos)), nil
}

const llamadas = 2000

func main() {
	comun.Titulo("9.1 · El mismo catálogo de 100 productos, en el cable")
	cat := catalogo(100)
	pb, _ := proto.Marshal(cat)
	// protojson mete espacios al azar a propósito (para que nadie dependa de sus bytes): se compacta.
	crudo, _ := protojson.Marshal(cat)
	var compacto bytes.Buffer
	json.Compact(&compacto, crudo)
	js := compacto.Bytes()
	fmt.Printf("  protobuf: %d bytes · JSON: %d bytes · JSON ocupa %s veces más\n", len(pb), len(js), strings.Replace(fmt.Sprintf("%.1f", float64(len(js))/float64(len(pb))), ".", ",", 1))

	comun.Titulo(fmt.Sprintf("9.2 · %s llamadas seguidas, una a una, en la misma máquina", "2.000"))
	addr, srv := comun.Arrancar(func(g *grpc.Server) {
		inv.RegisterInventarioServer(g, servidor{})
	})
	c := inv.NewInventarioClient(comun.Cliente(addr))
	g := medir(func() { c.Listar(context.Background(), &inv.ListarPedido{Productos: 100}) })
	srv.Stop()

	lis, _ := net.Listen("tcp", "127.0.0.1:0")
	go http.Serve(lis, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var p struct{ Productos int }
		json.NewDecoder(r.Body).Decode(&p)
		w.Header().Set("Content-Type", "application/json")
		b, _ := protojson.Marshal(catalogo(p.Productos))
		w.Write(b)
	}))
	url := "http://" + lis.Addr().String() + "/listar"
	cli := &http.Client{}
	r := medir(func() {
		resp, err := cli.Post(url, "application/json", bytes.NewReader([]byte(`{"productos":100}`)))
		if err == nil {
			var cat inv.Catalogo
			var buf bytes.Buffer
			buf.ReadFrom(resp.Body)
			resp.Body.Close()
			protojson.Unmarshal(buf.Bytes(), &cat)
		}
	})
	fmt.Printf("  gRPC: mediana %s · REST con JSON: mediana %s\n", g, r)
	fmt.Println("  en local y sin red de por medio: una diferencia que en internet se come la latencia")

	comun.Titulo("9.3 · curl no habla gRPC")
	addr, srv = comun.Arrancar(func(g *grpc.Server) { inv.RegisterInventarioServer(g, servidor{}) })
	out, _ := exec.Command("curl", "-sS", "--max-time", "3", "http://"+addr+"/inventario.v1.Inventario/Listar").CombinedOutput()
	fmt.Printf("  curl http://<servidor>/inventario.v1.Inventario/Listar -> %s\n", primera(out))

	comun.Titulo("9.4 · grpcurl sí, pero sin reflexión no sabe qué hay")
	out, _ = exec.Command("grpcurl", "-plaintext", addr, "list").CombinedOutput()
	fmt.Printf("  grpcurl list -> %s\n", primera(out))
	srv.Stop()
	addr, srv = comun.Arrancar(func(g *grpc.Server) {
		inv.RegisterInventarioServer(g, servidor{})
		reflection.Register(g)
	})
	out, _ = exec.Command("grpcurl", "-plaintext", addr, "list").CombinedOutput()
	fmt.Println("  con reflection.Register(s):")
	for _, l := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		fmt.Println("    " + l)
	}
	out, _ = exec.Command("grpcurl", "-plaintext", "-d", `{"productos": 1}`, addr, "inventario.v1.Inventario/Listar").CombinedOutput()
	fmt.Println("  grpcurl -d '{\"productos\": 1}' ... Listar:")
	for _, l := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		fmt.Println("    " + l)
	}
	srv.Stop()
}

// medir hace `llamadas` llamadas (tras 200 de calentamiento) y devuelve la mediana.
func medir(f func()) string {
	for i := 0; i < 200; i++ {
		f()
	}
	t := make([]time.Duration, llamadas)
	for i := range t {
		t0 := time.Now()
		f()
		t[i] = time.Since(t0)
	}
	sort.Slice(t, func(a, b int) bool { return t[a] < t[b] })
	return fmt.Sprintf("%d µs", t[len(t)/2].Microseconds())
}

func primera(b []byte) string {
	s := strings.TrimSpace(string(b))
	if i := strings.Index(s, "\n"); i >= 0 {
		s = s[:i]
	}
	return s
}
