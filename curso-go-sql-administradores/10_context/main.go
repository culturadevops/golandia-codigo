// Lección 10 · context y timeouts.
//
// La trampa: db.Query sin context espera lo que tarde la base de datos. Una consulta que
// se atasca atasca la petición, la goroutine y la conexión del pool.
package main

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"curso-sql/comun"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func activas(db *sql.DB) int {
	var n int
	db.QueryRow(`SELECT count(*) FROM pg_stat_activity
	             WHERE query LIKE 'SELECT pg_sleep(%' AND state = 'active'`).Scan(&n)
	return n
}

func main() {
	db := comun.Abrir()
	defer db.Close()

	comun.Titulo("10.1 · Sin context: esperas lo que tarde")
	t0 := time.Now()
	db.Exec("SELECT pg_sleep(3)")
	fmt.Printf("  db.Exec(pg_sleep(3))                     -> volvio a los %.1f s\n", time.Since(t0).Seconds())

	comun.Titulo("10.2 · Con context.WithTimeout de 1 s")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	t0 = time.Now()
	_, err := db.ExecContext(ctx, "SELECT pg_sleep(3)")
	cancel()
	fmt.Printf("  db.ExecContext(ctx, pg_sleep(3))         -> volvio a los %.1f s\n", time.Since(t0).Seconds())
	fmt.Println("  error:", err)
	time.Sleep(300 * time.Millisecond)
	fmt.Println("  consultas pg_sleep aun corriendo en el servidor:", activas(db))

	comun.Titulo("10.3 · La red por si alguien olvida el context: statement_timeout")
	red, _ := sql.Open("pgx", comun.DSN()+"&statement_timeout=1000")
	defer red.Close()
	t0 = time.Now()
	_, err = red.Exec("SELECT pg_sleep(3)")
	fmt.Printf("  sin context, con statement_timeout=1000  -> volvio a los %.1f s\n", time.Since(t0).Seconds())
	fmt.Println("  error:", err)
}
