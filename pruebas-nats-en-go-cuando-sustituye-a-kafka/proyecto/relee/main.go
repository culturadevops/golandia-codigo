// relee: en Kafka leer no borra. Otro grupo de consumidores vuelve a leer desde el principio.
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
)

// leer consume con un grupo durante unos segundos y confirma lo leído.
func leer(grupo string) {
	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
	defer cancel()
	cl, err := kgo.NewClient(kgo.SeedBrokers(os.Getenv("KAFKA")), kgo.ConsumeTopics("facturas"),
		kgo.ConsumerGroup(grupo), kgo.ConsumeResetOffset(kgo.NewOffset().AtStart()))
	if err != nil {
		panic(err)
	}
	defer cl.Close()
	n := 0
	for ctx.Err() == nil {
		f := cl.PollFetches(ctx)
		f.EachRecord(func(r *kgo.Record) {
			n++
			fmt.Printf("  %-11s leyó %s (partición %d, offset %d)\n", grupo, r.Value, r.Partition, r.Offset)
		})
		if n > 0 {
			cl.CommitUncommittedOffsets(context.Background())
		}
	}
	fmt.Printf("  %-11s total: %d\n", grupo, n)
}

func main() {
	ctx := context.Background()
	cl, err := kgo.NewClient(kgo.SeedBrokers(os.Getenv("KAFKA")))
	if err != nil {
		panic(err)
	}
	for i := 1; i <= 3; i++ {
		v := fmt.Sprintf("factura %d", i)
		if err := cl.ProduceSync(ctx, &kgo.Record{Topic: "facturas", Key: []byte("cliente-ana"), Value: []byte(v)}).FirstErr(); err != nil {
			panic(err)
		}
	}
	cl.Close()
	fmt.Println("producidas 3 facturas")

	fmt.Println("grupo facturacion, primera vez:")
	leer("facturacion")
	fmt.Println("grupo facturacion, otra vez:")
	leer("facturacion")
	fmt.Println("grupo auditoria, nuevo:")
	leer("auditoria")
}
