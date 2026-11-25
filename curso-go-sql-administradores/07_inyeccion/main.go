// Lección 7 · Inyección SQL y placeholders.
//
// La trampa: construir la consulta con fmt.Sprintf. Funciona con todos los nombres
// normales. Con el nombre que escribe un atacante, devuelve la tabla entera.
package main

import (
	"database/sql"
	"fmt"
	"slices"

	"curso-sql/comun"

	"github.com/jackc/pgx/v5"
)

func contar(db *sql.DB, q string, args ...any) string {
	rows, err := db.Query(q, args...)
	if err != nil {
		return "ERROR: " + err.Error()
	}
	defer rows.Close()
	n := 0
	for rows.Next() {
		n++
	}
	if n == 1 {
		return "1 fila"
	}
	return fmt.Sprintf("%d filas", n)
}

func main() {
	db := comun.Abrir()
	defer db.Close()
	comun.Sembrar(db)

	entrada := "x' OR '1'='1"

	comun.Titulo("7.1 · Con fmt.Sprintf")
	q := fmt.Sprintf("SELECT nombre, ip FROM servidores WHERE nombre = '%s'", "web-01")
	fmt.Println("  nombre = web-01        ->", contar(db, q))
	q = fmt.Sprintf("SELECT nombre, ip FROM servidores WHERE nombre = '%s'", entrada)
	fmt.Println("  la consulta que sale:  ", q)
	fmt.Printf("  nombre = %-13s ->  %s\n", entrada, contar(db, q))

	comun.Titulo("7.2 · Con placeholder $1")
	q = "SELECT nombre, ip FROM servidores WHERE nombre = $1"
	fmt.Println("  nombre = web-01        ->", contar(db, q, "web-01"))
	fmt.Printf("  nombre = %-13s ->  %s\n", entrada, contar(db, q, entrada))

	comun.Titulo("7.3 · Y el DROP TABLE?")
	q = fmt.Sprintf("SELECT nombre FROM servidores WHERE nombre = '%s'", "x'; DROP TABLE servidores; --")
	fmt.Println("  ", contar(db, q))
	fmt.Println("  Te salva el driver (protocolo extendido: una sentencia por consulta), no tu codigo.")
	fmt.Println("  La tabla sigue:", contar(db, "SELECT 1 FROM servidores"))

	comun.Titulo("7.4 · Lo que NO puede ir en un placeholder: el nombre de una tabla")
	fmt.Println("  SELECT count(*) FROM $1  ->", contar(db, "SELECT count(*) FROM $1", "servidores"))
	permitidas := []string{"servidores", "eventos"}
	for _, t := range []string{"servidores", "pg_authid"} {
		if !slices.Contains(permitidas, t) {
			fmt.Printf("  tabla %-11s -> rechazada: no esta en la lista permitida\n", t)
			continue
		}
		q := "SELECT count(*) FROM " + pgx.Identifier{t}.Sanitize()
		fmt.Printf("  tabla %-11s -> %s   (%s)\n", t, contar(db, q), q)
	}
}
