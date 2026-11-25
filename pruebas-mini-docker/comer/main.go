// comer: reserva memoria de 8 en 8 MB y lo dice, hasta 512 MB o hasta que alguien lo pare.
package main

import (
	"fmt"
	"time"
)

func main() {
	var bloques [][]byte
	for i := 1; i <= 64; i++ {
		b := make([]byte, 8<<20)
		for j := 0; j < len(b); j += 4096 {
			b[j] = 1 // tocar cada página: si no, el kernel no la entrega de verdad
		}
		bloques = append(bloques, b)
		fmt.Printf("reservados %d MB\n", i*8)
		time.Sleep(20 * time.Millisecond)
	}
	fmt.Println("terminé: nadie me paró")
	_ = bloques
}
