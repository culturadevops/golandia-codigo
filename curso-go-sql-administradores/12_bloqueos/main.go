// Lección 12 · Bloqueos: la migración que tumba las lecturas.
//
// El escenario de las tres de la mañana: alguien dejó una transacción abierta (una consola,
// un script que se colgó). Llega una migración inocente, un ALTER TABLE que tarda un
// milisegundo. No puede coger su cerrojo y se pone a la cola. Y TODAS las consultas que
// llegan después, hasta un simple SELECT, se ponen a la cola detrás de ella.
package main

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"curso-sql/comun"
)

func mostrarActividad(db *sql.DB) {
	rows, err := db.Query(`
		SELECT pid, state, coalesce(wait_event_type, '-'), pg_blocking_pids(pid)::text,
		       left(regexp_replace(query, '\s+', ' ', 'g'), 44)
		FROM pg_stat_activity
		WHERE datname = current_database() AND pid <> pg_backend_pid()
		  AND query NOT LIKE '%pg_stat_activity%' AND query <> ''
		ORDER BY backend_start`)
	if err != nil {
		fmt.Println("  ", err)
		return
	}
	defer rows.Close()
	fmt.Printf("  %-6s %-20s %-6s %-9s %s\n", "pid", "state", "espera", "bloqueado", "consulta")
	for rows.Next() {
		var pid int
		var st, w, b, q string
		rows.Scan(&pid, &st, &w, &b, &q)
		fmt.Printf("  %-6s %-20s %-6s %-9s %s\n", fmt.Sprint(pid), st, w, strings.Trim(b, "{}"), q)
	}
}

func main() {
	db := comun.Abrir()
	defer db.Close()
	comun.Sembrar(db)
	ctx := context.Background()

	comun.Titulo("12.1 · Una transaccion abierta + un ALTER TABLE + un SELECT")
	a, _ := db.Conn(ctx)
	a.ExecContext(ctx, "BEGIN")
	a.ExecContext(ctx, "UPDATE servidores SET cpu = cpu WHERE nombre = 'web-01'")
	fmt.Println("  A: BEGIN; UPDATE ... web-01;   y se queda ahi (idle in transaction)")

	migro := make(chan error, 1)
	go func() {
		_, err := db.ExecContext(ctx, "ALTER TABLE servidores ADD COLUMN zona text")
		migro <- err
	}()
	time.Sleep(500 * time.Millisecond)
	fmt.Println("  B: ALTER TABLE servidores ADD COLUMN zona text   (la migracion)")

	c, cancel := context.WithTimeout(ctx, 2*time.Second)
	lectura := make(chan error, 1)
	go func() {
		var n int
		lectura <- db.QueryRowContext(c, "SELECT count(*) FROM servidores").Scan(&n)
	}()
	time.Sleep(500 * time.Millisecond)
	fmt.Println("  C: SELECT count(*) FROM servidores                (una lectura cualquiera)")
	fmt.Println()
	mostrarActividad(db)
	fmt.Println()
	fmt.Println("  C ->", <-lectura)
	cancel()

	a.ExecContext(ctx, "ROLLBACK")
	fmt.Println("  A hace ROLLBACK ->  B termina:", <-migro)
	a.Close()

	comun.Titulo("12.2 · La migracion con lock_timeout: falla rapido y no arrastra a nadie")
	a, _ = db.Conn(ctx)
	a.ExecContext(ctx, "BEGIN")
	a.ExecContext(ctx, "UPDATE servidores SET cpu = cpu WHERE nombre = 'web-01'")
	m, _ := db.Conn(ctx)
	m.ExecContext(ctx, "SET lock_timeout = '1s'")
	t0 := time.Now()
	_, err := m.ExecContext(ctx, "ALTER TABLE servidores ADD COLUMN region text")
	fmt.Printf("  B con lock_timeout=1s -> %.1f s: %v\n", time.Since(t0).Seconds(), err)
	var n int
	c, cancel = context.WithTimeout(ctx, 2*time.Second)
	err = db.QueryRowContext(c, "SELECT count(*) FROM servidores").Scan(&n)
	cancel()
	fmt.Printf("  C -> %d filas, err=%v   (la lectura ya no espera)\n", n, err)
	a.ExecContext(ctx, "ROLLBACK")
	a.Close()
	m.Close()

	comun.Titulo("12.3 · Y la transaccion olvidada: idle_in_transaction_session_timeout")
	a, _ = db.Conn(ctx)
	a.ExecContext(ctx, "SET idle_in_transaction_session_timeout = '1s'")
	a.ExecContext(ctx, "BEGIN")
	a.ExecContext(ctx, "UPDATE servidores SET cpu = cpu WHERE nombre = 'web-01'")
	time.Sleep(1500 * time.Millisecond)
	_, err = a.ExecContext(ctx, "UPDATE servidores SET cpu = cpu WHERE nombre = 'web-02'")
	fmt.Println("  A, 1,5 s despues ->", err)
	a.Close()
}
