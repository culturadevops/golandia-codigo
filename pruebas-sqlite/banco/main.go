// banco — la MISMA carga contra SQLite y contra Postgres, para poder comparar.
//
// No mide "SQLite es rapido": mide DONDE se rompe. Por eso los escenarios cambian
// una sola cosa cada vez (el modo de diario, el tope de conexiones, el busy_timeout)
// y todo lo demas se queda igual.
package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	_ "modernc.org/sqlite"
)

func main() {
	motor := flag.String("motor", "sqlite", "sqlite | postgres")
	dsn := flag.String("dsn", "", "cadena de conexion")
	escritores := flag.Int("escritores", 8, "gorrutinas que escriben a la vez")
	lectores := flag.Int("lectores", 0, "gorrutinas que leen a la vez")
	segundos := flag.Float64("segundos", 3, "cuanto dura la corrida")
	maxAbiertas := flag.Int("maxabiertas", 0, "SetMaxOpenConns (0 = por defecto)")
	etiqueta := flag.String("etiqueta", "", "como se llama este escenario")
	limpiar := flag.Bool("limpiar", true, "borrar y recrear la tabla al empezar")
	flag.Parse()

	driver := "sqlite"
	if *motor == "postgres" {
		driver = "pgx"
	}
	db, err := sql.Open(driver, *dsn)
	if err != nil {
		fmt.Println("ERROR abriendo:", err)
		os.Exit(1)
	}
	defer db.Close()
	if *maxAbiertas > 0 {
		db.SetMaxOpenConns(*maxAbiertas)
	}

	if *limpiar {
		if _, err := db.Exec(`DROP TABLE IF EXISTS eventos`); err != nil {
			fmt.Println("ERROR limpiando:", err)
			os.Exit(1)
		}
		crear := `CREATE TABLE eventos (id INTEGER PRIMARY KEY AUTOINCREMENT, quien INTEGER, texto TEXT)`
		if *motor == "postgres" {
			crear = `CREATE TABLE eventos (id SERIAL PRIMARY KEY, quien INTEGER, texto TEXT)`
		}
		if _, err := db.Exec(crear); err != nil {
			fmt.Println("ERROR creando:", err)
			os.Exit(1)
		}
	}
	inserta := `INSERT INTO eventos (quien, texto) VALUES (?, ?)`
	lee := `SELECT COUNT(*) FROM eventos`
	if *motor == "postgres" {
		inserta = `INSERT INTO eventos (quien, texto) VALUES ($1, $2)`
	}

	var okEscritas, okLeidas, bloqueos, otros atomic.Int64
	ctx, parar := context.WithTimeout(context.Background(),
		time.Duration(*segundos*float64(time.Second)))
	defer parar()

	var wg sync.WaitGroup
	empezo := time.Now()

	for i := 0; i < *escritores; i++ {
		wg.Add(1)
		go func(quien int) {
			defer wg.Done()
			for ctx.Err() == nil {
				_, err := db.Exec(inserta, quien, "una linea de registro cualquiera")
				switch {
				case err == nil:
					okEscritas.Add(1)
				case esBloqueo(err):
					bloqueos.Add(1)
				default:
					otros.Add(1)
				}
			}
		}(i)
	}
	for i := 0; i < *lectores; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for ctx.Err() == nil {
				var n int
				if err := db.QueryRow(lee).Scan(&n); err == nil {
					okLeidas.Add(1)
				} else if esBloqueo(err) {
					bloqueos.Add(1)
				} else {
					otros.Add(1)
				}
			}
		}()
	}
	wg.Wait()
	dur := time.Since(empezo).Seconds()

	nombre := *etiqueta
	if nombre == "" {
		nombre = *motor
	}
	fmt.Printf("%-46s escrituras/s %8.0f   lecturas/s %8.0f   bloqueos %6d   otros %4d\n",
		nombre, float64(okEscritas.Load())/dur, float64(okLeidas.Load())/dur,
		bloqueos.Load(), otros.Load())
}

// esBloqueo: los dos motores dicen lo mismo con palabras distintas.
func esBloqueo(err error) bool {
	m := strings.ToLower(err.Error())
	return strings.Contains(m, "locked") || strings.Contains(m, "busy") ||
		strings.Contains(m, "deadlock")
}
