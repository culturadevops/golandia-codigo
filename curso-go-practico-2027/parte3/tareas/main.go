package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	ctx := context.Background()
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		log.Fatal("falta DATABASE_URL")
	}
	pool, err := ConectarPostgres(ctx, url, 4)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	aplicadas, err := Migrar(ctx, pool)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("migraciones aplicadas: %v", aplicadas)

	api := &API{almacen: &Postgres{pool: pool, timeout: 2 * time.Second}}
	log.Println("escuchando en :8080")
	log.Fatal(http.ListenAndServe(":8080", api.Rutas()))
}
