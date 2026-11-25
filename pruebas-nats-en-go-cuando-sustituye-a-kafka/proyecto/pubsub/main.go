// pubsub: NATS básico. Publicar sin nadie escuchando, y publicar con un suscriptor.
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/nats-io/nats.go"
)

func main() {
	nc, err := nats.Connect(os.Getenv("NATS_URL"))
	if err != nil {
		panic(err)
	}
	defer nc.Close()

	// 1) Sin suscriptor: el mensaje se publica… y no lo recibe nadie, nunca.
	nc.Publish("pedidos.nuevo", []byte("pedido 1"))
	nc.Flush()

	// 2) Con suscriptor.
	recibidos := make(chan string, 10)
	sub, _ := nc.Subscribe("pedidos.*", func(m *nats.Msg) { recibidos <- m.Subject + " → " + string(m.Data) })
	defer sub.Unsubscribe()
	nc.Publish("pedidos.nuevo", []byte("pedido 2"))
	nc.Publish("pedidos.pagado", []byte("pedido 2"))
	nc.Flush()

	for i := 0; i < 2; i++ {
		select {
		case m := <-recibidos:
			fmt.Println("recibido:", m)
		case <-time.After(time.Second):
			fmt.Println("nada más")
		}
	}
	select {
	case m := <-recibidos:
		fmt.Println("recibido:", m)
	case <-time.After(500 * time.Millisecond):
		fmt.Println("el «pedido 1» no llegó: se publicó antes de que hubiera suscriptor")
	}
}
