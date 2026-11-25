// cola: un grupo de cola. Tres trabajadores en el mismo grupo se reparten los mensajes.
package main

import (
	"fmt"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
)

func main() {
	nc, err := nats.Connect(os.Getenv("NATS_URL"))
	if err != nil {
		panic(err)
	}
	defer nc.Close()

	var mu sync.Mutex
	cuenta := map[string]int{}
	var wg sync.WaitGroup
	wg.Add(300)
	for _, nombre := range []string{"trabajador-a", "trabajador-b", "trabajador-c"} {
		n := nombre
		nc.QueueSubscribe("tareas", "trabajadores", func(m *nats.Msg) {
			mu.Lock()
			cuenta[n]++
			mu.Unlock()
			wg.Done()
		})
	}
	nc.Flush()
	for i := 0; i < 300; i++ {
		nc.Publish("tareas", []byte(fmt.Sprint(i)))
	}
	nc.Flush()
	listo := make(chan struct{})
	go func() { wg.Wait(); close(listo) }()
	select {
	case <-listo:
	case <-time.After(3 * time.Second):
	}
	nombres := []string{}
	total := 0
	for k, v := range cuenta {
		nombres = append(nombres, k)
		total += v
	}
	sort.Strings(nombres)
	for _, k := range nombres {
		fmt.Printf("%s: %d tareas\n", k, cuenta[k])
	}
	fmt.Printf("total: %d de 300, cada una a un solo trabajador\n", total)
}
