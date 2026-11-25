package main

// El cliente dice «tienes 200 ms». Se entera el servidor?
//
// En HTTP el servidor se entera de que el cliente COLGO, pero no de cuanto tiempo le quedaba.
// En gRPC el plazo viaja en una cabecera y el servidor puede leerlo ANTES de empezar a trabajar.
// La diferencia es la que hay entre abortar a mitad y no llegar a empezar.

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"banco/pb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type lento struct{ pb.UnimplementedPedidosServer }

func (lento) Listar(ctx context.Context, _ *pb.Peticion) (*pb.Respuesta, error) {
	if plazo, hay := ctx.Deadline(); hay {
		fmt.Printf("   [servidor gRPC] me dicen que tengo %.0f ms\n",
			float64(time.Until(plazo).Microseconds())/1000)
	} else {
		fmt.Println("   [servidor gRPC] no me han dicho ningun plazo")
	}
	select {
	case <-time.After(2 * time.Second):
		return &pb.Respuesta{}, nil
	case <-ctx.Done():
		fmt.Println("   [servidor gRPC] me lo cancelan, paro:", ctx.Err())
		return nil, ctx.Err()
	}
}

func main() {
	lis, _ := net.Listen("tcp", "127.0.0.1:0")
	s := grpc.NewServer()
	pb.RegisterPedidosServer(s, lento{})
	go s.Serve(lis)
	defer s.Stop()

	cc, err := grpc.NewClient(lis.Addr().String(),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	fmt.Println("gRPC - el cliente da 200 ms:")
	t := time.Now()
	_, err = pb.NewPedidosClient(cc).Listar(ctx, &pb.Peticion{Cuantos: 1})
	fmt.Printf("   [cliente] corta a los %.0f ms con %v\n",
		float64(time.Since(t).Microseconds())/1000, status.Code(err))
	time.Sleep(200 * time.Millisecond)

	fmt.Println()
	fmt.Println("HTTP - el mismo cliente con 200 ms de plazo:")
	lh, _ := net.Listen("tcp", "127.0.0.1:0")
	srv := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if plazo, hay := r.Context().Deadline(); hay {
			fmt.Printf("   [servidor HTTP] me dicen que tengo %v\n", time.Until(plazo))
		} else {
			fmt.Println("   [servidor HTTP] no se cuanto tiempo tengo: la peticion NO trae plazo")
		}
		select {
		case <-time.After(2 * time.Second):
			w.Write([]byte("tarde"))
		case <-r.Context().Done():
			fmt.Println("   [servidor HTTP] me entero de que colgo, pero ya habia empezado:",
				r.Context().Err())
		}
	})}
	go srv.Serve(lh)
	defer srv.Close()

	cl := &http.Client{Timeout: 200 * time.Millisecond}
	t = time.Now()
	_, _ = cl.Get("http://" + lh.Addr().String() + "/pedidos")
	fmt.Printf("   [cliente] corta a los %.0f ms\n", float64(time.Since(t).Microseconds())/1000)
	time.Sleep(400 * time.Millisecond) // que de tiempo a imprimirse lo del servidor
}
