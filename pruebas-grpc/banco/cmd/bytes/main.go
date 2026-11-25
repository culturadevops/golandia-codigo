package main

// ¿Cuantos bytes viajan de verdad? Cuatro combinaciones y cuatro tamaños de respuesta.
//
// Metodo: una llamada de calentamiento (que paga el saludo TLS-menos, el SETTINGS de HTTP/2 y
// la tabla HPACK), se pone el contador a cero, y se miden 20 llamadas. Se divide.
// ⚠️ Sin el calentamiento, gRPC sale MUCHO peor de lo que es: el coste de abrir la conexion
// HTTP/2 se repartia entre 1 llamada en vez de entre las miles que hace una app de verdad.

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"

	"banco/contar"
	"banco/datos"
	"banco/pb"
	"banco/servidores"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	gz "google.golang.org/grpc/encoding/gzip"
)

const VUELTAS = 20

func clienteREST(comp bool, c *contar.Contador) *http.Client {
	return &http.Client{Transport: &http.Transport{
		DisableCompression: !comp,
		DialContext: func(ctx context.Context, red, dir string) (net.Conn, error) {
			cn, err := (&net.Dialer{}).DialContext(ctx, red, dir)
			if err != nil {
				return nil, err
			}
			return c.Marcar(cn), nil
		},
	}}
}

func pideREST(cl *http.Client, dir string, n int) int {
	r, err := cl.Get(fmt.Sprintf("http://%s/pedidos?cuantos=%d", dir, n))
	if err != nil {
		panic(err)
	}
	defer r.Body.Close()
	var xs []datos.Pedido
	var cuerpo io.Reader = r.Body
	if r.Header.Get("Content-Encoding") == "gzip" {
		z, err := gzip.NewReader(r.Body)
		if err != nil {
			panic(err)
		}
		defer z.Close()
		cuerpo = z
	}
	if err := json.NewDecoder(cuerpo).Decode(&xs); err != nil {
		panic(err)
	}
	return len(xs)
}

func main() {
	dirREST, paraR := servidores.REST()
	defer paraR()
	dirGRPC, paraG := servidores.GRPC()
	defer paraG()

	var cR, cRz, cG, cGz contar.Contador
	rest := clienteREST(false, &cR)
	restz := clienteREST(true, &cRz)

	marca := func(c *contar.Contador) grpc.DialOption {
		return grpc.WithContextDialer(func(ctx context.Context, dir string) (net.Conn, error) {
			cn, err := (&net.Dialer{}).DialContext(ctx, "tcp", dir)
			if err != nil {
				return nil, err
			}
			return c.Marcar(cn), nil
		})
	}
	abre := func(c *contar.Contador) pb.PedidosClient {
		cc, err := grpc.NewClient(dirGRPC, grpc.WithTransportCredentials(insecure.NewCredentials()), marca(c))
		if err != nil {
			panic(err)
		}
		return pb.NewPedidosClient(cc)
	}
	gr, grz := abre(&cG), abre(&cGz)

	fmt.Printf("%8s | %14s | %14s | %14s | %14s | %s\n",
		"pedidos", "REST/JSON", "REST+gzip", "gRPC/proto", "gRPC+gzip", "JSON / gRPC")
	fmt.Println("---------|----------------|----------------|----------------|----------------|------------")

	for _, n := range []int{1, 10, 100, 1000} {
		ctx := context.Background()
		// calentamiento
		pideREST(rest, dirREST, n)
		pideREST(restz, dirREST, n)
		gr.Listar(ctx, &pb.Peticion{Cuantos: int32(n)})
		grz.Listar(ctx, &pb.Peticion{Cuantos: int32(n)}, grpc.UseCompressor(gz.Name))

		cR.Cero()
		cRz.Cero()
		cG.Cero()
		cGz.Cero()
		for i := 0; i < VUELTAS; i++ {
			pideREST(rest, dirREST, n)
			pideREST(restz, dirREST, n)
			if _, err := gr.Listar(ctx, &pb.Peticion{Cuantos: int32(n)}); err != nil {
				panic(err)
			}
			if _, err := grz.Listar(ctx, &pb.Peticion{Cuantos: int32(n)},
				grpc.UseCompressor(gz.Name)); err != nil {
				panic(err)
			}
		}
		a := float64(cR.Total()) / VUELTAS
		b := float64(cRz.Total()) / VUELTAS
		c := float64(cG.Total()) / VUELTAS
		d := float64(cGz.Total()) / VUELTAS
		fmt.Printf("%8d | %11.0f B | %11.0f B | %11.0f B | %11.0f B | %10.2fx\n", n, a, b, c, d, a/c)
	}

	fmt.Println()
	fmt.Println("Bytes por llamada, contados en el socket (cabeceras incluidas), media de", VUELTAS, "llamadas")
	fmt.Println("sobre una conexion ya abierta. La ultima columna es cuantas veces mas gasta REST/JSON")
	fmt.Println("que gRPC sin comprimir.")
}
