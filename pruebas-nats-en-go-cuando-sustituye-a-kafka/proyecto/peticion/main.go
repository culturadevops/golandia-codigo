// peticion: petición-respuesta sobre NATS, sin servidor HTTP.
package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/nats-io/nats.go"
)

func main() {
	nc, err := nats.Connect(os.Getenv("NATS_URL"))
	if err != nil {
		panic(err)
	}
	defer nc.Close()

	nc.Subscribe("precios.consultar", func(m *nats.Msg) {
		m.Respond([]byte("precio de " + strings.ToUpper(string(m.Data)) + ": 12,50"))
	})
	inicio := time.Now()
	r, err := nc.Request("precios.consultar", []byte("café"), time.Second)
	if err != nil {
		panic(err)
	}
	fmt.Printf("respuesta: %s (en %v)\n", r.Data, time.Since(inicio).Round(time.Microsecond))

	_, err = nc.Request("inventario.consultar", []byte("café"), time.Second)
	fmt.Println("sin nadie que responda:", err)
}
