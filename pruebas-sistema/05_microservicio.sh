#!/bin/sh
# 5 · El microservicio que pedia el comentario: las cuatro piezas juntas, y lo que
# pesa la imagen cuando el binario no necesita sistema operativo debajo.
# Se ejecuta en el HOST (necesita docker), no dentro del contenedor de Go.
set -e
cd "$(dirname "$0")"

mkdir -p _micro && cd _micro

cat > main.go <<'FIN'
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
FIN

cat > go.mod <<'FIN'
module ejemplo.com/micro

go 1.26
FIN

cat > Dockerfile <<'FIN'
FROM golang:alpine AS build
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /micro .

# scratch = imagen VACIA. Ni distribucion, ni shell, ni gestor de paquetes.
FROM scratch
COPY --from=build /micro /micro
EXPOSE 8080
ENTRYPOINT ["/micro"]
FIN

echo "== construyendo la imagen =="
docker build -q -t golandia/micro:scratch . >/dev/null

echo
echo "== lo que pesa el microservicio entero, en disco =="
docker images golandia/micro:scratch --format "   {{.Repository}}:{{.Tag}}  {{.Size}}"

echo
echo "== con que se compara =="
for i in python:3.13-alpine node:22-alpine; do
    docker pull -q $i >/dev/null 2>&1 || true
    docker images $i --format "   $i (imagen BASE, sin tu codigo)  {{.Size}}" 2>/dev/null
done

echo
echo "== y arranca? =="
docker run -d --rm --name micro-prueba -p 18080:8080 golandia/micro:scratch >/dev/null
sleep 1
echo "   GET /salud -> $(curl -s http://localhost:18080/salud)"
echo "   mandando SIGTERM (lo mismo que hace docker stop / kubernetes)..."
docker stop micro-prueba >/dev/null
echo "   parado limpiamente"
