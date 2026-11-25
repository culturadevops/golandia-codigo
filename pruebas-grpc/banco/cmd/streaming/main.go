package main

// 10.000 pedidos de golpe contra 10.000 en streaming.
// Lo que importa no es el total: es CUANDO LLEGA EL PRIMERO, que es lo que ve el usuario.

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"banco/datos"
	"banco/pb"
	"banco/servidores"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const CUANTOS = 10000

func main() {
	dirREST, paraR := servidores.REST()
	defer paraR()
	dirGRPC, paraG := servidores.GRPC()
	defer paraG()

	cl := &http.Client{Transport: &http.Transport{DisableCompression: true}}
	cc, err := grpc.NewClient(dirGRPC, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	gr := pb.NewPedidosClient(cc)
	ctx := context.Background()

	// calentamiento
	r0, _ := cl.Get(fmt.Sprintf("http://%s/pedidos?cuantos=%d", dirREST, CUANTOS))
	io.Copy(io.Discard, r0.Body)
	r0.Body.Close()
	gr.Listar(ctx, &pb.Peticion{Cuantos: CUANTOS})

	// 1) REST: hay que esperar a que llegue el JSON entero antes de tener el primer pedido
	t := time.Now()
	r, err := cl.Get(fmt.Sprintf("http://%s/pedidos?cuantos=%d", dirREST, CUANTOS))
	if err != nil {
		panic(err)
	}
	var xs []datos.Pedido
	json.NewDecoder(r.Body).Decode(&xs)
	r.Body.Close()
	restTotal := time.Since(t)

	// 2) gRPC unario: mismo problema, el mensaje llega entero o no llega
	t = time.Now()
	resp, err := gr.Listar(ctx, &pb.Peticion{Cuantos: CUANTOS})
	if err != nil {
		panic(err)
	}
	unarioTotal := time.Since(t)

	// 3) gRPC en streaming: el primero llega mucho antes
	t = time.Now()
	st, err := gr.ListarEnStreaming(ctx, &pb.Peticion{Cuantos: CUANTOS})
	if err != nil {
		panic(err)
	}
	var primero time.Duration
	n := 0
	for {
		_, err := st.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}
		if n == 0 {
			primero = time.Since(t)
		}
		n++
	}
	streamTotal := time.Since(t)

	ms := func(d time.Duration) float64 { return float64(d.Microseconds()) / 1000 }
	fmt.Printf("%d pedidos\n\n", CUANTOS)
	fmt.Printf("%-26s %10s %14s\n", "", "total", "primer pedido")
	fmt.Printf("%-26s %9.1f ms %13.1f ms   (%d pedidos)\n", "REST/JSON de una vez", ms(restTotal), ms(restTotal), len(xs))
	fmt.Printf("%-26s %9.1f ms %13.1f ms   (%d pedidos)\n", "gRPC unario", ms(unarioTotal), ms(unarioTotal), len(resp.Pedidos))
	fmt.Printf("%-26s %9.1f ms %13.1f ms   (%d pedidos)\n", "gRPC en streaming", ms(streamTotal), ms(primero), n)
	fmt.Println()
	fmt.Printf("El primer pedido llega %.0fx antes en streaming que esperando el JSON entero.\n",
		ms(restTotal)/ms(primero))
}
