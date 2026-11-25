package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/salud", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "vivo")
	})
	srv := &http.Server{Addr: ":8080", Handler: mux}

	// pieza 3: enterarse de que te apagan
	ctx, parar := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer parar()

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			panic(err)
		}
	}()
	fmt.Println("escuchando en el 8080")

	<-ctx.Done()
	fmt.Println("apagando con orden...")
	cierre, cancelar := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelar()
	_ = srv.Shutdown(cierre) // termina lo que estaba en vuelo
	fmt.Println("adios")
}
