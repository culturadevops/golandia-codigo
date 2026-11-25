// Package comun tiene lo que comparten todas las lecciones: la cadena de conexión,
// abrir la base de datos y la tabla de ejemplo con la que trabaja casi todo el curso.
package comun

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib" // registra el driver "pgx" para database/sql
)

// DSN es la cadena de conexión. ejecutar_todo.py levanta un Postgres en el puerto 55432;
// si usas el tuyo, exporta CURSO_PG.
func DSN() string {
	if v := os.Getenv("CURSO_PG"); v != "" {
		return v
	}
	return "postgres://curso:curso@127.0.0.1:55432/tienda?sslmode=disable"
}

// Abrir abre la base de datos Y comprueba que responde. Lo segundo es la lección 1.
func Abrir() *sql.DB {
	db, err := sql.Open("pgx", DSN())
	if err != nil {
		log.Fatal(err)
	}
	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}
	return db
}

// Titulo imprime la cabecera de cada parte de una lección.
func Titulo(s string) { fmt.Printf("\n== %s ==\n", s) }

// Debe ejecuta una sentencia y aborta si falla. Solo para preparar el escenario.
func Debe(db *sql.DB, q string, args ...any) {
	if _, err := db.Exec(q, args...); err != nil {
		log.Fatalf("preparando %q: %v", q, err)
	}
}

// Sembrar crea la tabla servidores con ocho filas. Una de ellas tiene las notas en NULL,
// como pasa siempre en producción y nunca en tu máquina.
func Sembrar(db *sql.DB) {
	Debe(db, `DROP TABLE IF EXISTS servidores`)
	Debe(db, `CREATE TABLE servidores (
		id     serial PRIMARY KEY,
		nombre text UNIQUE NOT NULL,
		ip     text NOT NULL,
		rol    text NOT NULL,
		cpu    int  NOT NULL,
		notas  text
	)`)
	Debe(db, `INSERT INTO servidores (nombre, ip, rol, cpu, notas) VALUES
		('web-01',   '10.0.1.11', 'web',   91, 'detras del balanceador'),
		('web-02',   '10.0.1.12', 'web',   35, 'detras del balanceador'),
		('web-03',   '10.0.1.13', 'web',   42, 'detras del balanceador'),
		('db-01',    '10.0.2.21', 'db',    64, 'primaria'),
		('db-02',    '10.0.2.22', 'db',    12, NULL),
		('cache-01', '10.0.3.31', 'cache', 88, 'redis'),
		('cola-01',  '10.0.4.41', 'cola',  20, 'rabbitmq'),
		('bastion',  '10.0.0.5',  'admin',  3, 'solo ssh')`)
}
