package main

import (
	"context"
	"fmt"
	"time"
)

func trabajo() {
	// El descarte: ni se guarda el cancel ni se llama. Es el caso real.
	ctx, _ := context.WithTimeout(context.Background(), time.Second)
	<-ctx.Done()
	fmt.Println("listo")
}

func main() { trabajo() }
