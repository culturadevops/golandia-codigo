// Lección 6 · El pool, medido.
//
// 40 consultas que tardan 100 ms cada una, lanzadas a la vez, con pools de distinto tamaño.
// Y la otra cara: Postgres tiene un máximo de conexiones para TODOS los clientes juntos.
// Tu pool por el número de réplicas de tu servicio tiene que caber ahí.
package main

import (
	"database/sql"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"curso-sql/comun"
)

func lanzar(db *sql.DB, n int) (time.Duration, int64) {
	var wg sync.WaitGroup
	var fallos int64
	t0 := time.Now()
	for range n {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := db.Exec("SELECT pg_sleep(0.1)"); err != nil {
				atomic.AddInt64(&fallos, 1)
			}
		}()
	}
	wg.Wait()
	return time.Since(t0), fallos
}

func main() {
	comun.Titulo("6.1 · 40 consultas de 100 ms, a la vez")
	for _, max := range []int{1, 5, 20, 40} {
		db := comun.Abrir()
		db.SetMaxOpenConns(max)
		db.SetMaxIdleConns(max)
		d, _ := lanzar(db, 40)
		st := db.Stats()
		fmt.Printf("  pool de %2d -> %5d ms   (esperaron turno: %2d consultas, %6d ms en total)\n",
			max, d.Milliseconds(), st.WaitCount, st.WaitDuration.Milliseconds())
		db.Close()
	}

	comun.Titulo("6.2 · El limite que no es tuyo: max_connections")
	db := comun.Abrir()
	var maxConn string
	db.QueryRow("SHOW max_connections").Scan(&maxConn)
	fmt.Println("  max_connections del servidor:", maxConn)
	db.Close()

	fmt.Println("  3 replicas de tu servicio, cada una con un pool de 40 = 120 conexiones:")
	var pools []*sql.DB
	var wg sync.WaitGroup
	var fallos int64
	var primero atomic.Value
	for range 3 {
		p := comun.Abrir()
		p.SetMaxOpenConns(40)
		pools = append(pools, p)
	}
	for _, p := range pools {
		for range 40 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				if _, err := p.Exec("SELECT pg_sleep(0.5)"); err != nil {
					atomic.AddInt64(&fallos, 1)
					primero.CompareAndSwap(nil, err.Error())
				}
			}()
		}
	}
	wg.Wait()
	fmt.Printf("  consultas que fallaron: %d de 120\n", fallos)
	if e := primero.Load(); e != nil {
		fmt.Println("  error:", e)
	}
	for _, p := range pools {
		p.Close()
	}

	comun.Titulo("6.3 · Los cuatro ajustes")
	fmt.Println("  db.SetMaxOpenConns(n)        tope de conexiones de ESTE proceso")
	fmt.Println("  db.SetMaxIdleConns(n)        cuantas se quedan abiertas esperando")
	fmt.Println("  db.SetConnMaxLifetime(d)     renovarlas (balanceadores, failover)")
	fmt.Println("  db.SetConnMaxIdleTime(d)     cerrar las que nadie usa")
}
