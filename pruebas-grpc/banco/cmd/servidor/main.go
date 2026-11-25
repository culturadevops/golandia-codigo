package main

// Deja el servicio escuchando en un puerto fijo para poder interrogarlo desde fuera con grpcurl.

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"banco/servidores"
)

func main() {
	dir, para := servidores.GRPCEn("127.0.0.1:50051")
	defer para()
	fmt.Println("gRPC escuchando en", dir)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
}
