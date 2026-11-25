package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"
)

func main() {
	// Un servidor que tarda 2 segundos a proposito.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-time.After(2 * time.Second):
			fmt.Fprintln(w, "por fin")
		case <-r.Context().Done():
			fmt.Println("   [servidor] el cliente se fue:", r.Context().Err())
		}
	}))
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel() // <-- SIEMPRE. Sin esto se fuga el temporizador.

	req, _ := http.NewRequestWithContext(ctx, "GET", srv.URL, nil)
	inicio := time.Now()
	_, err := http.DefaultClient.Do(req)
	fmt.Printf("volvio en %v\n", time.Since(inicio).Round(10*time.Millisecond))
	fmt.Println("error:", err)
	fmt.Println()
	fmt.Println("ctx.Err():                          ", ctx.Err())
	fmt.Println("errors.Is(ctx.Err(), DeadlineExceeded):", ctx.Err() == context.DeadlineExceeded)
	time.Sleep(100 * time.Millisecond) // que de tiempo a ver el mensaje del servidor
}
