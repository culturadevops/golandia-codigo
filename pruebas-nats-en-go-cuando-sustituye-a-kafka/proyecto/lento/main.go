// lento: un suscriptor de NATS básico que no da abasto. El cliente descarta lo que no cabe.
package main

import (
	"fmt"
	"os"
	"sync/atomic"
	"time"

	"github.com/nats-io/nats.go"
)

func main() {
	avisos := atomic.Int64{}
	sub, err := nats.Connect(os.Getenv("NATS_URL"), nats.ErrorHandler(func(_ *nats.Conn, _ *nats.Subscription, err error) {
		if avisos.Add(1) == 1 {
			fmt.Println("aviso del cliente:", err)
		}
	}))
	if err != nil {
		panic(err)
	}
	defer sub.Close()

	var procesados atomic.Int64
	s, _ := sub.Subscribe("lecturas", func(m *nats.Msg) {
		time.Sleep(time.Millisecond) // 1 ms por mensaje: unos 1.000 por segundo como mucho
		procesados.Add(1)
	})
	s.SetPendingLimits(1000, 1024*1024) // caben 1.000 mensajes esperando
	sub.Flush()

	pub, _ := nats.Connect(os.Getenv("NATS_URL"))
	defer pub.Close()
	t0 := time.Now()
	for i := 0; i < 20000; i++ {
		pub.Publish("lecturas", []byte("21.5"))
	}
	pub.Flush()
	fmt.Printf("publicados 20000 en %v\n", time.Since(t0).Round(time.Millisecond))

	time.Sleep(3 * time.Second)
	caidos, _ := s.Dropped()
	fmt.Printf("procesados: %d · descartados por el cliente: %d\n", procesados.Load(), caidos)
}
