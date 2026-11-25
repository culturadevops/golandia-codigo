package main

// Que pasa cuando el contrato cambia y las dos partes no se actualizan a la vez.
// Es la pregunta que decide si puedes desplegar tus servicios por separado o no.

import (
	"fmt"

	"banco/pb"
	"banco/pbmalo"
	"banco/pbv2"

	"google.golang.org/protobuf/proto"
)

func main() {
	// El servidor NUEVO manda un pedido con el campo 8, que el cliente viejo no conoce.
	nuevo := &pbv2.Pedido{Id: 42, Cliente: "cliente-00042", Articulo: "articulo-0042",
		Unidades: 3, Centimos: 1999, Estado: "enviado", Creado: "2026-08-24T10:00:00Z",
		Almacen: "MADRID-3"}
	crudo, err := proto.Marshal(nuevo)
	if err != nil {
		panic(err)
	}
	fmt.Printf("1) el servidor NUEVO manda %d bytes, con el campo 8 (almacen=%q)\n\n",
		len(crudo), nuevo.Almacen)

	// El cliente VIEJO, generado del .proto de 7 campos, lo lee.
	viejo := &pb.Pedido{}
	if err := proto.Unmarshal(crudo, viejo); err != nil {
		fmt.Println("   el cliente viejo FALLO:", err)
	} else {
		fmt.Printf("2) el cliente VIEJO lo lee sin enterarse: id=%d cliente=%q estado=%q\n",
			viejo.Id, viejo.Cliente, viejo.Estado)
		fmt.Println("   no conoce el campo 8, y no le importa. Nadie se cae.")
	}

	// Y lo mejor: si ese cliente viejo lo vuelve a serializar, el campo que no entiende
	// NO SE PIERDE. protobuf guarda los campos desconocidos y los devuelve tal cual.
	otraVez, _ := proto.Marshal(viejo)
	rebote := &pbv2.Pedido{}
	proto.Unmarshal(otraVez, rebote)
	fmt.Printf("3) y si el viejo lo reenvia, el campo 8 SIGUE AHI: almacen=%q\n", rebote.Almacen)
	fmt.Println("   protobuf conserva los campos que no conoce. Por eso un servicio viejo")
	fmt.Println("   en medio no destroza los datos de uno nuevo.")

	// Y ahora lo que SI rompe: reutilizar un numero para otro tipo.
	fmt.Println()
	malo := &pbmalo.Pedido{}
	err = proto.Unmarshal(crudo, malo)
	fmt.Println("4) ahora alguien reutiliza el campo 2 (era string cliente) para un entero:")
	if err != nil {
		fmt.Printf("   proto.Unmarshal devuelve error: %v\n", err)
	} else {
		fmt.Printf("   NO da error. descuento=%d, y el nombre del cliente se perdio\n",
			malo.Descuento)
	}
	fmt.Println("   Este es el bug silencioso: anadir campos es gratis, reutilizar numeros no.")
}
