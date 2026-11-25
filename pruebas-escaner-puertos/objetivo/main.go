// objetivo — abre una lista de puertos conocidos y se queda escuchando.
// Sirve para tener una VERDAD contra la que comparar: sabemos exactamente
// cuántos puertos hay abiertos, así que se puede medir si el escáner los encuentra.
package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
)

var PUERTOS = []int{22, 80, 443, 3306, 5432, 6379, 8080, 8443, 9090, 9200,
	11211, 15672, 27017, 50000, 61616}

func main() {
	abiertos := 0
	for _, p := range PUERTOS {
		l, err := net.Listen("tcp", ":"+strconv.Itoa(p))
		if err != nil {
			fmt.Fprintf(os.Stderr, "no pude abrir %d: %v\n", p, err)
			continue
		}
		abiertos++
		go func(l net.Listener) {
			for {
				c, err := l.Accept()
				if err != nil {
					return
				}
				c.Close()
			}
		}(l)
	}
	fmt.Printf("OBJETIVO LISTO: %d puertos abiertos de %d pedidos\n", abiertos, len(PUERTOS))
	fmt.Printf("   %v\n", PUERTOS)
	select {}
}
