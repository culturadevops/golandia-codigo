// Una goroutine que provoca un pánico y nadie lo recupera.
package main

import (
	"fmt"
	"time"
)

func main() {
	fmt.Println("servidor atendiendo")
	go func() {
		var m map[string]int
		m["x"] = 1 // escribir en un mapa nil: pánico
	}()
	time.Sleep(100 * time.Millisecond)
	fmt.Println("esto no se imprime")
}
