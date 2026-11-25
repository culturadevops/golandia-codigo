// Lección 8 · Exec y RowsAffected.
//
// La trampa: un UPDATE o un DELETE que no encuentra ninguna fila NO es un error.
// Tu herramienta dice «servidor puesto en mantenimiento» y no ha tocado nada.
package main

import (
	"database/sql"
	"fmt"

	"curso-sql/comun"
)

func mantenimientoMal(db *sql.DB, nombre string) error {
	_, err := db.Exec("UPDATE servidores SET notas = 'MANTENIMIENTO' WHERE nombre = $1", nombre)
	return err
}

func mantenimientoBien(db *sql.DB, nombre string) error {
	res, err := db.Exec("UPDATE servidores SET notas = 'MANTENIMIENTO' WHERE nombre = $1", nombre)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		return fmt.Errorf("%s: esperaba cambiar 1 fila y cambie %d", nombre, n)
	}
	return nil
}

func main() {
	db := comun.Abrir()
	defer db.Close()
	comun.Sembrar(db)

	comun.Titulo("8.1 · Sin mirar RowsAffected")
	for _, n := range []string{"web-01", "Web-02"} {
		if err := mantenimientoMal(db, n); err == nil {
			fmt.Printf("  %-7s -> puesto en mantenimiento  (err = %v)\n", n, err)
		}
	}
	var marcados int
	db.QueryRow("SELECT count(*) FROM servidores WHERE notas = 'MANTENIMIENTO'").Scan(&marcados)
	fmt.Printf("  en la tabla hay %d servidor(es) en mantenimiento, no 2\n", marcados)

	comun.Titulo("8.2 · Mirando RowsAffected")
	for _, n := range []string{"web-03", "Web-02"} {
		if err := mantenimientoBien(db, n); err != nil {
			fmt.Printf("  %-7s -> ERROR: %v\n", n, err)
		} else {
			fmt.Printf("  %-7s -> puesto en mantenimiento\n", n)
		}
	}

	comun.Titulo("8.3 · El otro lado: el filtro vacio que lo cambia TODO")
	filtro := "" // venia de un flag que nadie relleno
	res, _ := db.Exec("UPDATE servidores SET cpu = 0 WHERE nombre LIKE '%' || $1 || '%'", filtro)
	n, _ := res.RowsAffected()
	fmt.Printf("  filtro = %q -> %d filas cambiadas de 8\n", filtro, n)
	fmt.Println("  RowsAffected lo dice DESPUES. Para abortar a tiempo: transaccion + comprobar + Rollback (leccion 9).")
}
