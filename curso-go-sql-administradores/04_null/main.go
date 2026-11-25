// Lección 4 · NULL.
//
// La trampa: en tu base de datos de pruebas todas las filas tienen notas. En producción
// una tiene NULL, y el Scan a string revienta justo en esa fila, a mitad del listado.
package main

import (
	"database/sql"
	"fmt"

	"curso-sql/comun"
)

func main() {
	db := comun.Abrir()
	defer db.Close()
	comun.Sembrar(db)

	comun.Titulo("4.1 · Scan de NULL a string")
	rows, _ := db.Query("SELECT nombre, notas FROM servidores ORDER BY id")
	for rows.Next() {
		var nombre, notas string
		if err := rows.Scan(&nombre, &notas); err != nil {
			fmt.Printf("    %-9s ERROR: %v\n", nombre, err)
			break
		}
		fmt.Printf("    %-9s %s\n", nombre, notas)
	}
	rows.Close()

	comun.Titulo("4.2 · Tres formas de recibir un NULL")
	var ns sql.NullString
	db.QueryRow("SELECT notas FROM servidores WHERE nombre = 'db-02'").Scan(&ns)
	fmt.Printf("  sql.NullString -> Valid=%v String=%q\n", ns.Valid, ns.String)

	var p *string
	db.QueryRow("SELECT notas FROM servidores WHERE nombre = 'db-02'").Scan(&p)
	fmt.Printf("  *string        -> %v\n", p)

	var s string
	db.QueryRow("SELECT COALESCE(notas, '') FROM servidores WHERE nombre = 'db-02'").Scan(&s)
	fmt.Printf("  COALESCE       -> %q\n", s)

	comun.Titulo("4.3 · NULL no es la cadena vacia")
	db.Exec("UPDATE servidores SET notas = '' WHERE nombre = 'bastion'")
	var nulos, vacios int
	db.QueryRow("SELECT count(*) FILTER (WHERE notas IS NULL), count(*) FILTER (WHERE notas = '') FROM servidores").Scan(&nulos, &vacios)
	fmt.Printf("  notas IS NULL -> %d   notas = '' -> %d\n", nulos, vacios)
	var cuenta int
	db.QueryRow("SELECT count(*) FROM servidores WHERE notas <> 'primaria'").Scan(&cuenta)
	fmt.Printf("  WHERE notas <> 'primaria' -> %d de 8   (la fila con NULL no entra)\n", cuenta)
}
