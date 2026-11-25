package main

// El contrato lo comprueba el COMPILADOR. Este servidor se deja un metodo del `.proto` sin
// implementar y ademas cambia el tipo que devuelve el otro. En REST esto es un 500 en produccion
// un martes; aqui es un error de compilacion antes de que exista el binario.

import (
	"context"

	"banco/pb"

	"google.golang.org/grpc"
)

type servicioIncompleto struct{} // ⛔ a proposito: SIN pb.UnimplementedPedidosServer

// Y la firma esta mal: el .proto dice que Listar devuelve Respuesta, no Pedido.
func (servicioIncompleto) Listar(context.Context, *pb.Peticion) (*pb.Pedido, error) {
	return &pb.Pedido{Id: 1}, nil
}

func main() {
	s := grpc.NewServer()
	pb.RegisterPedidosServer(s, servicioIncompleto{})
}
