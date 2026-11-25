// Lección 7 · El balanceo: una conexión HTTP/2 larga se pega a una réplica.
//
// Corre en contenedores (lo orquesta ejecutar_todo.py): dos o tres réplicas con el mismo
// nombre en la red de Docker y un cliente que les pregunta «¿quién eres?».
//
//	07_balanceo servidor          réplica: contesta con su nombre de host
//	07_balanceo cliente <fase>    pick_first | round_robin | escalar
package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"sort"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"

	"curso-grpc/comun"
	inv "curso-grpc/gen/inventario"
)

const objetivo = "dns:///inventario:50051"

type servidor struct {
	inv.UnimplementedInventarioServer
	nombre string
}

func (s servidor) Quien(context.Context, *inv.Vacio) (*inv.Replica, error) {
	return &inv.Replica{Nombre: s.nombre}, nil
}

func main() {
	switch os.Args[1] {
	case "servidor":
		servir()
	case "cliente":
		cliente(os.Args[2])
	}
}

func servir() {
	nombre, _ := os.Hostname()
	var opciones []grpc.ServerOption
	if edad := os.Getenv("MAX_CONNECTION_AGE"); edad != "" {
		d, _ := time.ParseDuration(edad)
		opciones = append(opciones, grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionAge: d, MaxConnectionAgeGrace: 5 * time.Second,
		}))
	}
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		panic(err)
	}
	s := grpc.NewServer(opciones...)
	inv.RegisterInventarioServer(s, servidor{nombre: nombre})
	s.Serve(lis)
}

// repartir hace n llamadas, una cada `cada`, y cuenta cuántas contestó cada réplica.
func repartir(c inv.InventarioClient, n int, cada time.Duration) string {
	cuenta := map[string]int{}
	for i := 0; i < n; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		r, err := c.Quien(ctx, &inv.Vacio{}, grpc.WaitForReady(true))
		cancel()
		if err != nil {
			cuenta["error"]++
		} else {
			cuenta[r.Nombre]++
		}
		time.Sleep(cada)
	}
	return pintar(cuenta)
}

func pintar(cuenta map[string]int) string {
	nombres := make([]string, 0, len(cuenta))
	for k := range cuenta {
		nombres = append(nombres, k)
	}
	sort.Strings(nombres)
	partes := make([]string, 0, len(nombres))
	for _, k := range nombres {
		partes = append(partes, fmt.Sprintf("%s: %d", k, cuenta[k]))
	}
	return strings.Join(partes, " · ")
}

func cliente(fase string) {
	var opciones []grpc.DialOption
	if fase != "pick_first" {
		opciones = append(opciones, grpc.WithDefaultServiceConfig(`{"loadBalancingConfig": [{"round_robin":{}}]}`))
	}
	c := inv.NewInventarioClient(comun.Cliente(objetivo, opciones...))
	switch fase {
	case "pick_first", "round_robin":
		fmt.Printf("  100 llamadas -> %s\n", repartir(c, 100, 0))
	case "escalar":
		fmt.Printf("  con 2 réplicas, 100 llamadas -> %s\n", repartir(c, 100, 0))
		fmt.Println("ESPERANDO_TERCERA")
		time.Sleep(2 * time.Second)
		fmt.Println("  ya hay 3 réplicas; una llamada cada 200 ms, contadas de 15 en 15 s:")
		for v := 0; v < 3; v++ {
			fmt.Printf("    de %2d a %2d s -> %s\n", v*15, (v+1)*15, repartir(c, 75, 200*time.Millisecond))
		}
	}
}
