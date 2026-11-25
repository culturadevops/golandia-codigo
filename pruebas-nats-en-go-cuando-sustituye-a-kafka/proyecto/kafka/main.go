// kafka: producir y consumir en Kafka desde Go con franz-go. La misma clave, la misma partición.
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	cl, err := kgo.NewClient(kgo.SeedBrokers(os.Getenv("KAFKA")), kgo.ConsumeTopics("pedidos"),
		kgo.ConsumeResetOffset(kgo.NewOffset().AtStart()))
	if err != nil {
		panic(err)
	}
	defer cl.Close()

	for _, r := range []struct{ clave, valor string }{
		{"cliente-ana", "pedido 1"}, {"cliente-luis", "pedido 2"}, {"cliente-ana", "pedido 3"},
		{"cliente-luis", "pedido 4"}, {"cliente-ana", "pedido 5"}, {"cliente-eva", "pedido 6"},
	} {
		res := cl.ProduceSync(ctx, &kgo.Record{Topic: "pedidos", Key: []byte(r.clave), Value: []byte(r.valor)})
		rec, err := res.First()
		if err != nil {
			panic(err)
		}
		fmt.Printf("producido %-12s %s → partición %d, offset %d\n", r.clave, r.valor, rec.Partition, rec.Offset)
	}

	leidos := 0
	for leidos < 6 {
		f := cl.PollFetches(ctx)
		if ctx.Err() != nil {
			break
		}
		f.EachRecord(func(r *kgo.Record) {
			leidos++
			fmt.Printf("leído     partición %d: %s %s\n", r.Partition, r.Key, r.Value)
		})
	}
}
