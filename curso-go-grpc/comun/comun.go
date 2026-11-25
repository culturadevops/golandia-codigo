// Package comun tiene lo que repiten todas las lecciones: levantar un servidor en un puerto
// libre, abrir un cliente sin TLS y escribir los títulos de cada apartado.
package comun

import (
	"fmt"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Arrancar escucha en un puerto libre de 127.0.0.1 y sirve en segundo plano.
func Arrancar(registrar func(*grpc.Server), opciones ...grpc.ServerOption) (string, *grpc.Server) {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}
	s := grpc.NewServer(opciones...)
	registrar(s)
	go s.Serve(lis)
	return lis.Addr().String(), s
}

// PuertoLibre devuelve una dirección en la que ahora mismo no escucha nadie.
func PuertoLibre() string {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}
	addr := lis.Addr().String()
	lis.Close()
	return addr
}

// Cliente abre una conexión sin TLS. Las lecciones corren en local; la de producción lleva TLS.
func Cliente(addr string, opciones ...grpc.DialOption) *grpc.ClientConn {
	opciones = append([]grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}, opciones...)
	c, err := grpc.NewClient(addr, opciones...)
	if err != nil {
		panic(err)
	}
	return c
}

func Titulo(t string) { fmt.Printf("\n== %s ==\n", t) }

// Ms redondea una duración a milisegundos enteros, que es lo que se lee en pantalla.
func Ms(d time.Duration) string {
	return fmt.Sprintf("%d ms", d.Round(time.Millisecond).Milliseconds())
}
