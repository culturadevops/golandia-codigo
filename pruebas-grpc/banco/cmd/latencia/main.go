package main

// ¿Y en tiempo? 2.000 llamadas seguidas de cada uno, con 100 pedidos de respuesta.
// Se ordenan y se sacan p50 y p99: la media esconde justo lo que duele.

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"time"

	"banco/datos"
	"banco/pb"
	"banco/servidores"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	LLAMADAS = 2000
	CUANTOS  = 100
)

func percentil(xs []time.Duration, p float64) time.Duration {
	i := int(float64(len(xs)-1) * p)
	return xs[i]
}

func resumen(nombre string, xs []time.Duration) {
	sort.Slice(xs, func(i, j int) bool { return xs[i] < xs[j] })
	var suma time.Duration
	for _, d := range xs {
		suma += d
	}
	fmt.Printf("%-12s  p50 %8.3f ms   p99 %8.3f ms   media %8.3f ms   min %8.3f ms\n",
		nombre,
		float64(percentil(xs, .50).Microseconds())/1000,
		float64(percentil(xs, .99).Microseconds())/1000,
		float64(suma.Microseconds())/1000/float64(len(xs)),
		float64(xs[0].Microseconds())/1000)
}

func main() {
	dirREST, paraR := servidores.REST()
	defer paraR()
	dirGRPC, paraG := servidores.GRPC()
	defer paraG()

	cl := &http.Client{Transport: &http.Transport{DisableCompression: true, MaxIdleConnsPerHost: 4}}
	cc, err := grpc.NewClient(dirGRPC, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	gr := pb.NewPedidosClient(cc)
	ctx := context.Background()
	url := fmt.Sprintf("http://%s/pedidos?cuantos=%d", dirREST, CUANTOS)

	unaREST := func() {
		r, err := cl.Get(url)
		if err != nil {
			panic(err)
		}
		var xs []datos.Pedido
		json.NewDecoder(r.Body).Decode(&xs)
		r.Body.Close()
	}
	unaGRPC := func() {
		if _, err := gr.Listar(ctx, &pb.Peticion{Cuantos: CUANTOS}); err != nil {
			panic(err)
		}
	}

	for i := 0; i < 200; i++ { // calentamiento
		unaREST()
		unaGRPC()
	}

	tr := make([]time.Duration, LLAMADAS)
	tg := make([]time.Duration, LLAMADAS)
	for i := 0; i < LLAMADAS; i++ {
		t := time.Now()
		unaREST()
		tr[i] = time.Since(t)
		t = time.Now()
		unaGRPC()
		tg[i] = time.Since(t)
	}

	fmt.Printf("%d llamadas de cada uno, %d pedidos por respuesta, mismo proceso y localhost\n\n",
		LLAMADAS, CUANTOS)
	resumen("REST/JSON", tr)
	resumen("gRPC/proto", tg)
}
