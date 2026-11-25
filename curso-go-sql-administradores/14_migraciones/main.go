// Lección 14 · Migraciones: un migrador de 60 líneas, y las tres cosas que lo rompen.
//
//  1. Dos réplicas arrancan a la vez y las dos migran.
//  2. Una migración con dos sentencias falla en la segunda y deja la primera aplicada.
//  3. CREATE INDEX CONCURRENTLY no puede ir dentro de una transacción.
package main

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"

	"curso-sql/comun"
)

//go:embed migraciones/*.sql
var ficheros embed.FS

type migracion struct {
	version int
	nombre  string
	sql     string
}

func cargar(hasta int) []migracion {
	ents, _ := ficheros.ReadDir("migraciones")
	var ms []migracion
	for _, e := range ents {
		v, _ := strconv.Atoi(e.Name()[:3])
		if v > hasta {
			continue
		}
		b, _ := ficheros.ReadFile("migraciones/" + e.Name())
		ms = append(ms, migracion{v, e.Name(), string(b)})
	}
	sort.Slice(ms, func(i, j int) bool { return ms[i].version < ms[j].version })
	return ms
}

// migrar aplica lo pendiente. Cada migración, en su transacción, junto con su apunte.
func migrar(db *sql.DB, hasta int, cerrojo bool) (aplicadas int, err error) {
	ctx := context.Background()
	c, err := db.Conn(ctx)
	if err != nil {
		return 0, err
	}
	defer c.Close()
	if cerrojo {
		// Un número cualquiera, el mismo en todas las réplicas. El cerrojo es de la sesión.
		c.ExecContext(ctx, "SELECT pg_advisory_lock(20260928)")
		defer c.ExecContext(ctx, "SELECT pg_advisory_unlock(20260928)")
	}
	c.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS schema_migraciones (
		version int PRIMARY KEY, aplicada timestamptz NOT NULL DEFAULT now())`)
	for _, m := range cargar(hasta) {
		var ya bool
		c.QueryRowContext(ctx, "SELECT EXISTS (SELECT 1 FROM schema_migraciones WHERE version = $1)", m.version).Scan(&ya)
		if ya {
			continue
		}
		tx, err := c.BeginTx(ctx, nil)
		if err != nil {
			return aplicadas, err
		}
		if _, err := tx.Exec(m.sql); err != nil {
			tx.Rollback()
			return aplicadas, fmt.Errorf("%s: %w", m.nombre, err)
		}
		tx.Exec("INSERT INTO schema_migraciones (version) VALUES ($1)", m.version)
		if err := tx.Commit(); err != nil {
			return aplicadas, fmt.Errorf("%s: %w", m.nombre, err)
		}
		aplicadas++
	}
	return aplicadas, nil
}

func limpiar(db *sql.DB) {
	comun.Debe(db, "DROP TABLE IF EXISTS eventos, schema_migraciones")
}

func dosReplicas(db *sql.DB, cerrojo bool) {
	var wg sync.WaitGroup
	res := make([]string, 2)
	for i := range 2 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			n, err := migrar(db, 3, cerrojo)
			if err != nil {
				res[i] = fmt.Sprintf("aplico %d y FALLO: %v", n, err)
			} else {
				res[i] = fmt.Sprintf("aplico %d, ok", n)
			}
		}()
	}
	wg.Wait()
	for i, r := range res {
		fmt.Printf("  replica %d -> %s\n", i+1, r)
	}
}

func columnas(db *sql.DB) string {
	rows, _ := db.Query(`SELECT column_name FROM information_schema.columns
	                     WHERE table_name = 'eventos' ORDER BY ordinal_position`)
	defer rows.Close()
	var cs []string
	for rows.Next() {
		var c string
		rows.Scan(&c)
		cs = append(cs, c)
	}
	return strings.Join(cs, ", ")
}

func main() {
	db := comun.Abrir()
	defer db.Close()

	comun.Titulo("14.1 · Dos replicas arrancan a la vez, SIN cerrojo")
	limpiar(db)
	dosReplicas(db, false)

	comun.Titulo("14.2 · Las mismas dos, con pg_advisory_lock")
	limpiar(db)
	dosReplicas(db, true)

	comun.Titulo("14.3 · La 004 tiene dos sentencias y la segunda esta mal")
	fmt.Print("  004_origen.sql:\n")
	for _, l := range strings.Split(strings.TrimSpace(cargar(4)[3].sql), "\n") {
		fmt.Println("    " + l)
	}
	fmt.Println("  sin transaccion, sentencia a sentencia (como hacen muchas herramientas):")
	for _, s := range strings.Split(cargar(4)[3].sql, ";") {
		if strings.TrimSpace(s) == "" {
			continue
		}
		if _, err := db.Exec(s); err != nil {
			fmt.Println("    ->", err)
		}
	}
	fmt.Println("    columnas de eventos:", columnas(db))
	fmt.Println("    la mitad aplicada, y la 004 sin apuntar: al reintentar, falla la PRIMERA.")
	comun.Debe(db, "ALTER TABLE eventos DROP COLUMN origen")
	fmt.Println("  en su transaccion (el migrador de arriba):")
	_, err := migrar(db, 4, true)
	fmt.Println("    ->", err)
	fmt.Println("    columnas de eventos:", columnas(db))

	comun.Titulo("14.4 · CREATE INDEX CONCURRENTLY dentro de una transaccion")
	tx, _ := db.Begin()
	_, err = tx.Exec("CREATE INDEX CONCURRENTLY eventos_nivel ON eventos (nivel)")
	tx.Rollback()
	fmt.Println("  ->", err)
	_, err = db.Exec("CREATE INDEX CONCURRENTLY eventos_nivel ON eventos (nivel)")
	fmt.Println("  fuera de la transaccion ->", err)
}
