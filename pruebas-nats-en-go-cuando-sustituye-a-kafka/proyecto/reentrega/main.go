// reentrega: qué hace JetStream con un mensaje que nadie confirma, y con un stream que se llena.
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
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	nc, err := nats.Connect(os.Getenv("NATS_URL"))
	if err != nil {
		panic(err)
	}
	defer nc.Close()
	js, _ := jetstream.New(nc)

	// 1) Un consumidor que tiene 2 segundos para confirmar cada mensaje.
	s, err := js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{Name: "TAREAS", Subjects: []string{"tareas.>"}})
	if err != nil {
		panic(err)
	}
	c, _ := s.CreateOrUpdateConsumer(ctx, jetstream.ConsumerConfig{
		Durable: "trabajador", AckPolicy: jetstream.AckExplicitPolicy, AckWait: 2 * time.Second,
	})
	js.Publish(ctx, "tareas.cobrar", []byte("cobrar pedido 7"))

	t0 := time.Now()
	for intento := 1; intento <= 2; intento++ {
		lote, _ := c.Fetch(1, jetstream.FetchMaxWait(5*time.Second))
		for m := range lote.Messages() {
			md, _ := m.Metadata()
			fmt.Printf("%.1f s · recibido %q · entrega número %d\n",
				time.Since(t0).Seconds(), m.Data(), md.NumDelivered)
			if intento == 1 {
				fmt.Println("        (el trabajador se cae: no confirma)")
				continue
			}
			// DoubleAck espera a que el servidor registre la confirmación; Ack no espera.
			if err := m.DoubleAck(ctx); err != nil {
				panic(err)
			}
			fmt.Println("        confirmado")
		}
	}
	info, _ := c.Info(ctx)
	fmt.Printf("pendientes de confirmar: %d\n\n", info.NumAckPending)

	// 2) Un stream con límite: solo guarda los 3 últimos mensajes.
	l, err := js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name: "LECTURAS", Subjects: []string{"lecturas.>"}, MaxMsgs: 3, Discard: jetstream.DiscardOld,
	})
	if err != nil {
		panic(err)
	}
	for i := 1; i <= 5; i++ {
		js.Publish(ctx, "lecturas.temp", []byte(fmt.Sprintf("lectura %d", i)))
	}
	li, _ := l.Info(ctx)
	fmt.Printf("publicadas 5 · en el stream: %d (de la secuencia %d a la %d)\n",
		li.State.Msgs, li.State.FirstSeq, li.State.LastSeq)
}
