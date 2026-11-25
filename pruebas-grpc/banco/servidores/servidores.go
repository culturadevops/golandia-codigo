package servidores

// Levanta LOS DOS servicios con el mismo dato detras. Cada programa de medida los arranca en su
// propio proceso y en puertos que pide el sistema, para que nada se pise.

import (
	"encoding/json"
	"net"
	"net/http"

	"banco/datos"
	"banco/pb"

	"google.golang.org/grpc"
	_ "google.golang.org/grpc/encoding/gzip" // registra el compresor, no lo activa
	"google.golang.org/grpc/reflection"
)

// ── REST ──────────────────────────────────────────────────────────────────────────────
func REST() (string, func()) {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/pedidos", func(w http.ResponseWriter, r *http.Request) {
		n := 100
		if v := r.URL.Query().Get("cuantos"); v != "" {
			n = 0
			for _, c := range v {
				n = n*10 + int(c-'0')
			}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(datos.Muchos(n))
	})
	// La compresion la decide el cliente con Accept-Encoding; el servidor la sirve si se la piden.
	s := &http.Server{Handler: gzipSiLaPiden(mux)}
	go s.Serve(lis)
	return lis.Addr().String(), func() { s.Close() }
}

// ── gRPC ──────────────────────────────────────────────────────────────────────────────
type servicio struct{ pb.UnimplementedPedidosServer }

func aPB(p datos.Pedido) *pb.Pedido {
	return &pb.Pedido{Id: p.ID, Cliente: p.Cliente, Articulo: p.Articulo,
		Unidades: p.Unidades, Centimos: p.Centimos, Estado: p.Estado, Creado: p.Creado}
}

func (servicio) Listar(_ context0, in *pb.Peticion) (*pb.Respuesta, error) {
	xs := datos.Muchos(int(in.Cuantos))
	out := make([]*pb.Pedido, len(xs))
	for i, p := range xs {
		out[i] = aPB(p)
	}
	return &pb.Respuesta{Pedidos: out}, nil
}

func (servicio) ListarEnStreaming(in *pb.Peticion, srv pb.Pedidos_ListarEnStreamingServer) error {
	for i := 0; i < int(in.Cuantos); i++ {
		if err := srv.Send(aPB(datos.Uno(i))); err != nil {
			return err
		}
	}
	return nil
}

func GRPC() (string, func()) { return GRPCEn("127.0.0.1:0") }

// Puerto fijo: `grpcurl` es otro proceso y necesita saber a donde llamar.
func GRPCEn(dir string) (string, func()) {
	lis, err := net.Listen("tcp", dir)
	if err != nil {
		panic(err)
	}
	s := grpc.NewServer()
	pb.RegisterPedidosServer(s, servicio{})
	reflection.Register(s) // sin esto, grpcurl no sabe que metodos hay
	go s.Serve(lis)
	return lis.Addr().String(), func() { s.Stop() }
}
