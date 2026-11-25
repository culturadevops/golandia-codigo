// jetstream: el mismo mensaje, pero guardado en un stream y leído después.
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
		Name: "PEDIDOS", Subjects: []string{"pedidos.>"}, Storage: jetstream.FileStorage, MaxAge: 7 * 24 * time.Hour,
	})
	if err != nil {
		panic(err)
	}

	// Publicar SIN ningún consumidor todavía.
	for i := 1; i <= 3; i++ {
		ack, err := js.Publish(ctx, "pedidos.nuevo", []byte(fmt.Sprintf("pedido %d", i)))
		if err != nil {
			panic(err)
		}
		fmt.Printf("publicado pedido %d · secuencia %d\n", i, ack.Sequence)
	}
	info, _ := s.Info(ctx)
	fmt.Printf("en el stream: %d mensajes, %d bytes\n", info.State.Msgs, info.State.Bytes)

	// Ahora llega un consumidor duradero y lee desde el principio.
	c, _ := s.CreateOrUpdateConsumer(ctx, jetstream.ConsumerConfig{Durable: "facturacion", AckPolicy: jetstream.AckExplicitPolicy})
	lote, _ := c.Fetch(3, jetstream.FetchMaxWait(2*time.Second))
	for m := range lote.Messages() {
		md, _ := m.Metadata()
		fmt.Printf("consumidor facturacion: %s (secuencia %d)\n", m.Data(), md.Sequence.Stream)
		m.Ack()
	}
}
