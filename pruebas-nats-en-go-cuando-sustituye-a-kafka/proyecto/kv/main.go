// kv: el almacén clave-valor de JetStream, con historial.
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

	kv, err := js.CreateOrUpdateKeyValue(ctx, jetstream.KeyValueConfig{Bucket: "config", History: 5})
	if err != nil {
		panic(err)
	}
	kv.Put(ctx, "app.timeout", []byte("5s"))
	kv.Put(ctx, "app.timeout", []byte("10s"))
	rev, _ := kv.Put(ctx, "app.timeout", []byte("30s"))

	e, _ := kv.Get(ctx, "app.timeout")
	fmt.Printf("app.timeout = %s (revisión %d)\n", e.Value(), rev)
	hist, _ := kv.History(ctx, "app.timeout")
	for _, h := range hist {
		fmt.Printf("  historial: revisión %d → %s\n", h.Revision(), h.Value())
	}
}
