// atras: 1.000 mensajes y un consumidor que solo lee 200. Ni Kafka ni JetStream pierden el resto:
// lo cuentan como retraso.
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/twmb/franz-go/pkg/kgo"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Kafka: producir 1.000 y leer 200 con el grupo «lento».
	p, err := kgo.NewClient(kgo.SeedBrokers(os.Getenv("KAFKA")))
	if err != nil {
		panic(err)
	}
	for i := 0; i < 1000; i++ {
		p.Produce(ctx, &kgo.Record{Topic: "lecturas", Value: []byte("21.5")}, nil)
	}
	if err := p.Flush(ctx); err != nil {
		panic(err)
	}
	p.Close()
	c, err := kgo.NewClient(kgo.SeedBrokers(os.Getenv("KAFKA")), kgo.ConsumeTopics("lecturas"),
		kgo.ConsumerGroup("lento"), kgo.ConsumeResetOffset(kgo.NewOffset().AtStart()), kgo.DisableAutoCommit())
	if err != nil {
		panic(err)
	}
	var leidos []*kgo.Record
	for len(leidos) < 200 {
		f := c.PollRecords(ctx, 200-len(leidos))
		leidos = append(leidos, f.Records()...)
	}
	if err := c.CommitRecords(ctx, leidos...); err != nil {
		panic(err)
	}
	c.Close()
	fmt.Printf("kafka:     producidos 1000 · el grupo lento leyó y confirmó %d\n", len(leidos))

	// JetStream: lo mismo.
	nc, err := nats.Connect(os.Getenv("NATS_URL"))
	if err != nil {
		panic(err)
	}
	defer nc.Close()
	js, _ := jetstream.New(nc)
	s, err := js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{Name: "LECTURAS2", Subjects: []string{"lect2.>"}})
	if err != nil {
		panic(err)
	}
	for i := 0; i < 1000; i++ {
		js.PublishAsync("lect2.temp", []byte("21.5"))
	}
	<-js.PublishAsyncComplete()
	cons, _ := s.CreateOrUpdateConsumer(ctx, jetstream.ConsumerConfig{Durable: "lento", AckPolicy: jetstream.AckExplicitPolicy})
	n := 0
	for n < 200 {
		lote, _ := cons.Fetch(200-n, jetstream.FetchMaxWait(2*time.Second))
		for m := range lote.Messages() {
			m.DoubleAck(ctx)
			n++
		}
	}
	info, _ := cons.Info(ctx)
	fmt.Printf("jetstream: publicados 1000 · el consumidor lento leyó y confirmó %d · pendientes: %d\n",
		n, info.NumPending)
}
