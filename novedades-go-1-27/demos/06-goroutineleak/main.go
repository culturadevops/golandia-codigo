// Go 1.27: el perfil goroutineleak ya no es experimento.
package main

import (
	"fmt"
	"os"
	"runtime/pprof"
	"time"
)

// buscar lanza una goroutine que manda su resultado por un canal sin búfer.
// Si el que llama se va por el timeout, nadie recibe nunca: la goroutine queda bloqueada para siempre.
func buscar(timeout time.Duration) (string, error) {
	ch := make(chan string)
	go func() {
		time.Sleep(50 * time.Millisecond) // la "consulta" lenta
		ch <- "resultado"
	}()
	select {
	case r := <-ch:
		return r, nil
	case <-time.After(timeout):
		return "", fmt.Errorf("timeout")
	}
}

func main() {
	for i := 0; i < 3; i++ {
		_, err := buscar(10 * time.Millisecond)
		fmt.Println("buscar:", err)
	}
	time.Sleep(200 * time.Millisecond) // las 3 goroutines ya están bloqueadas en ch <-

	p := pprof.Lookup("goroutineleak")
	if p == nil {
		fmt.Println(`pprof.Lookup("goroutineleak") = nil: este Go no tiene el perfil`)
		return
	}
	p.WriteTo(os.Stdout, 1)
}
