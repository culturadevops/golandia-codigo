// dedup: JetStream descarta un mensaje repetido si lleva el mismo Nats-Msg-Id.
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	nc, err := nats.Connect(os.Getenv("NATS_URL"))
	if err != nil {
		panic(err)
	}
	defer nc.Close()
	js, _ := jetstream.New(nc)
	s, err := js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name: "COBROS", Subjects: []string{"cobros.>"}, Duplicates: 2 * time.Minute,
	})
	if err != nil {
		panic(err)
	}
	// El productor reintenta porque no le llegó la confirmación: manda el MISMO cobro dos veces.
	for intento := 1; intento <= 2; intento++ {
		ack, err := js.Publish(ctx, "cobros.nuevo", []byte("cobrar pedido 7"), jetstream.WithMsgID("pedido-7"))
		if err != nil {
			panic(err)
		}
		fmt.Printf("intento %d · secuencia %d · duplicado: %v\n", intento, ack.Sequence, ack.Duplicate)
	}
	info, _ := s.Info(ctx)
	fmt.Printf("en el stream: %d mensaje\n", info.State.Msgs)
}
