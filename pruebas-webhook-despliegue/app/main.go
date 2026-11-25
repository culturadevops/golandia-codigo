// La app que se despliega: dice su versión y se apaga limpio con SIGTERM.
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

const version = "VERSION"

func main() {
	srv := &http.Server{Addr: os.Getenv("ADDR"), Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "hola desde", version)
	})}
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGTERM)
		<-c
		srv.Shutdown(context.Background())
	}()
	srv.ListenAndServe()
}
