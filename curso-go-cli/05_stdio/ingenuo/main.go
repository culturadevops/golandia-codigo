// Lección 5, versión ingenua: todo a stdout, y con colores siempre.
// En la terminal se ve igual de bien que la buena. En un pipe, no.
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"curso-cli/comun"
)

func main() {
	urls := os.Args[1:]
	fmt.Printf("revisando %d endpoints (timeout 10s)\n", len(urls)) // ⛔ progreso por stdout
	for _, u := range urls {
		r := comun.Revisar(context.Background(), u, 10*time.Second)
		l := comun.Linea(r)
		c := "\x1b[32m"
		if !r.OK() {
			c = "\x1b[31m"
		}
		fmt.Println(c + l[:6] + "\x1b[0m" + l[6:]) // ⛔ color aunque no haya terminal
	}
}
