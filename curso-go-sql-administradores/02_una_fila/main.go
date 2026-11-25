// Lección 2 · Una fila: QueryRow, Scan y sql.ErrNoRows.
//
// La trampa: tratar cualquier error como «no encontrado». Una columna mal escrita, la base
// de datos caída o un timeout acaban diciendo «ese servidor no existe», y nadie se entera.
package main

import (
	"database/sql"
	"errors"
	"fmt"

	"curso-sql/comun"
)

// ipMal: el patrón que se ve en mil sitios. Cualquier error = «no existe».
func ipMal(db *sql.DB, q, nombre string) string {
	var ip string
	if err := db.QueryRow(q, nombre).Scan(&ip); err != nil {
		return "no existe"
	}
	return ip
}

// ipBien separa «no hay fila», que es un resultado, de «algo se rompió», que es un error.
func ipBien(db *sql.DB, q, nombre string) (string, error) {
	var ip string
	err := db.QueryRow(q, nombre).Scan(&ip)
	if errors.Is(err, sql.ErrNoRows) {
		return "", fmt.Errorf("%s: no existe", nombre)
	}
	if err != nil {
		return "", fmt.Errorf("buscando %s: %w", nombre, err)
	}
	return ip, nil
}

func main() {
	db := comun.Abrir()
	defer db.Close()
	comun.Sembrar(db)

	bien := "SELECT ip FROM servidores WHERE nombre = $1"
	typo := "SELECT ip_addr FROM servidores WHERE nombre = $1" // la columna se llama ip

	comun.Titulo("2.1 · QueryRow + Scan")
	var ip string
	var cpu int
	db.QueryRow("SELECT ip, cpu FROM servidores WHERE nombre = $1", "db-01").Scan(&ip, &cpu)
	fmt.Printf("  db-01 -> ip %s, cpu %d%%\n", ip, cpu)

	comun.Titulo("2.2 · La trampa: todo error es «no existe»")
	fmt.Printf("  ipMal(\"web-09\")                   -> %s\n", ipMal(db, bien, "web-09"))
	fmt.Printf("  ipMal(\"db-01\") con la columna mal -> %s\n", ipMal(db, typo, "db-01"))
	fmt.Println("  Las dos dicen lo mismo. La segunda es un fallo de codigo, y db-01 SI existe.")

	comun.Titulo("2.3 · Lo correcto: errors.Is(err, sql.ErrNoRows)")
	_, err := ipBien(db, bien, "web-09")
	fmt.Println("  ipBien(\"web-09\")                  ->", err)
	_, err = ipBien(db, typo, "db-01")
	fmt.Println("  ipBien(\"db-01\") con la columna mal ->", err)
}
