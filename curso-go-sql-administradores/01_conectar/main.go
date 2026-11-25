// Lección 1 · Conectar: sql.Open no conecta.
//
// La trampa: sql.Open con la clave mal, o contra un servidor que no existe, devuelve
// err == nil. No ha hablado con nadie. El primer error llega con la primera consulta,
// que en un servicio puede ser horas después de arrancar.
package main

import (
	"database/sql"
	"fmt"

	"curso-sql/comun"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func probar(etiqueta, dsn string) {
	db, err := sql.Open("pgx", dsn)
	fmt.Printf("  %-22s sql.Open -> err = %v\n", etiqueta, err)
	defer db.Close()
	fmt.Printf("  %-22s conexiones abiertas: %d\n", "", db.Stats().OpenConnections)
	err = db.Ping()
	fmt.Printf("  %-22s db.Ping  -> %v\n", "", err)
}

func main() {
	comun.Titulo("1.1 · sql.Open no conecta")
	probar("clave MAL", "postgres://curso:otra@127.0.0.1:55432/tienda?sslmode=disable")
	fmt.Println()
	probar("servidor que no existe", "postgres://curso:curso@127.0.0.1:1/tienda?sslmode=disable&connect_timeout=2")

	comun.Titulo("1.2 · Lo correcto: abrir Y hacer Ping al arrancar")
	db, _ := sql.Open("pgx", comun.DSN())
	defer db.Close()
	fmt.Printf("  antes del Ping:   conexiones abiertas = %d\n", db.Stats().OpenConnections)
	if err := db.Ping(); err != nil {
		fmt.Println("  no hay base de datos:", err)
		return
	}
	fmt.Printf("  despues del Ping: conexiones abiertas = %d\n", db.Stats().OpenConnections)
	var version string
	db.QueryRow("SHOW server_version").Scan(&version)
	fmt.Println("  servidor: PostgreSQL", version)

	comun.Titulo("1.3 · *sql.DB no es una conexion: es un pool")
	fmt.Println("  Se abre UNA vez al arrancar y se comparte entre todas las goroutines.")
	fmt.Println("  Abrir un sql.DB por peticion es abrir un pool por peticion.")
}
