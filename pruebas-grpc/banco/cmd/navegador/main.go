package main

// ⛔ El «cuando NO» mas caro: un navegador NO puede llamar a un servidor gRPC.
// No es una limitacion de permisos ni de CORS: fetch() no deja controlar los marcos de HTTP/2
// que gRPC necesita. Aqui se hace exactamente lo que haria un navegador —una peticion HTTP/1.1
// normal y corriente— contra el puerto de gRPC, y se enseña lo que contesta.

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"time"

	"banco/servidores"
)

func main() {
	dir, para := servidores.GRPC()
	defer para()

	cn, err := net.Dial("tcp", dir)
	if err != nil {
		panic(err)
	}
	defer cn.Close()

	peticion := "GET /pedidos.Pedidos/Listar HTTP/1.1\r\nHost: " + dir + "\r\nAccept: application/json\r\n\r\n"
	fmt.Println("Lo que manda un navegador:")
	fmt.Println("  " + strings.ReplaceAll(strings.TrimSpace(peticion), "\r\n", "\n  "))
	fmt.Println()

	cn.Write([]byte(peticion))
	cn.SetReadDeadline(time.Now().Add(3 * time.Second))
	lector := bufio.NewReader(cn)
	buf := make([]byte, 512)
	n, err := lector.Read(buf)

	fmt.Println("Lo que contesta el servidor gRPC:")
	if n > 0 {
		fmt.Printf("  %d bytes: %q\n", n, string(buf[:n]))
	} else {
		fmt.Println("  nada. Ni una cabecera, ni un 400, ni un 404.")
	}
	if err != nil {
		fmt.Printf("  error de lectura: %v\n", err)
	}
	fmt.Println()
	fmt.Println("Por eso hace falta un proxy (grpc-web / Envoy) para hablar con gRPC desde el navegador.")
}
