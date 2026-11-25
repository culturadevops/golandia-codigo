// huerfano arranca un Redis con testcontainers y muere con os.Exit SIN terminarlo, como un test
// que revienta o un Ctrl+C en mitad de la suite. ¿Quién lo limpia? Ryuk, si está activo.
package main

import (
	"context"
	"fmt"
	"os"

	tcredis "github.com/testcontainers/testcontainers-go/modules/redis"
)

func main() {
	c, err := tcredis.Run(context.Background(), "redis:7-alpine")
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(2)
	}
	fmt.Println("contenedor arrancado:", c.GetContainerID()[:12])
	os.Exit(1) // sin Terminate: el proceso muere y el contenedor se queda
}
