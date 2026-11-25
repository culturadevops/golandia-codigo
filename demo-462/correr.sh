#!/usr/bin/env bash
# Demo del vídeo 462 «SQL injection en Go» — se EJECUTA y su salida va a SALIDA.md.
# Una tabla de usuarios en SQLite (driver en Go puro, sin cgo) y el mismo login escrito de dos formas:
#   A) la consulta armada con fmt.Sprintf;
#   B) la consulta con placeholder (?) de database/sql.
# Misma entrada maliciosa en las dos. Trabaja en un directorio temporal: no toca nada fuera.
set -u
T=$(mktemp -d)
cd "$T"
go mod init demo/login >/dev/null 2>&1
go get modernc.org/sqlite@v1.39.0 >/dev/null 2>&1

cat > main.go <<'EOF'
package main

import (
	"database/sql"
	"fmt"
	"os"

	_ "modernc.org/sqlite"
)

func loginSprintf(db *sql.DB, usuario, clave string) (string, error) {
	q := fmt.Sprintf("SELECT nombre FROM usuarios WHERE usuario = '%s' AND clave = '%s'", usuario, clave)
	fmt.Println("  consulta:", q)
	var nombre string
	err := db.QueryRow(q).Scan(&nombre)
	return nombre, err
}

func loginPlaceholder(db *sql.DB, usuario, clave string) (string, error) {
	var nombre string
	err := db.QueryRow("SELECT nombre FROM usuarios WHERE usuario = ? AND clave = ?", usuario, clave).Scan(&nombre)
	return nombre, err
}

func main() {
	db, _ := sql.Open("sqlite", ":memory:")
	db.Exec("CREATE TABLE usuarios (usuario TEXT, clave TEXT, nombre TEXT)")
	db.Exec("INSERT INTO usuarios VALUES ('admin', 'x9$Kq!2', 'Administración'), ('ana', 'gato123', 'Ana')")

	usuario, clave := os.Args[1], os.Args[2]
	fmt.Printf("entrada: usuario=%q clave=%q\n", usuario, clave)

	fmt.Println("A) fmt.Sprintf")
	n, err := loginSprintf(db, usuario, clave)
	if err != nil {
		fmt.Println("  resultado: acceso denegado (", err, ")")
	} else {
		fmt.Println("  resultado: ENTRASTE como", n)
	}

	fmt.Println("B) placeholder ?")
	n, err = loginPlaceholder(db, usuario, clave)
	if err != nil {
		fmt.Println("  resultado: acceso denegado (", err, ")")
	} else {
		fmt.Println("  resultado: ENTRASTE como", n)
	}
}
EOF

go mod tidy >/dev/null 2>&1
go build -o login . 2>&1
echo "\$ grep modernc go.mod"
grep "modernc.org/sqlite " go.mod | sed "s/^	//"
echo
echo "\$ ./login ana gato123"
./login ana gato123
echo
echo "\$ ./login ana nolase"
./login ana nolase
echo
echo "\$ ./login admin \"' OR '1'='1\""
./login admin "' OR '1'='1"
echo
echo "\$ ./login \"admin' --\" cualquiera"
./login "admin' --" cualquiera
cd / && rm -rf "$T"
