// escenario le fabrica a pgsalud los problemas de verdad, durante unos segundos:
// una transacción olvidada, una consulta que no acaba y una sesión bloqueada por la primera.
//
//	go run ./16_proyecto/escenario -durante 12s
package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"curso-sql/comun"
)

func main() {
	durante := flag.Duration("durante", 12*time.Second, "cuanto tiempo mantener el escenario")
	flag.Parse()
	db := comun.Abrir()
	defer db.Close()
	comun.Sembrar(db)
	ctx := context.Background()

	olvidada, _ := db.Conn(ctx)
	olvidada.ExecContext(ctx, "BEGIN")
	olvidada.ExecContext(ctx, "UPDATE servidores SET cpu = cpu + 1 WHERE nombre = 'db-01'")

	lenta, _ := db.Conn(ctx)
	go lenta.ExecContext(ctx, fmt.Sprintf("SELECT pg_sleep(%d)", int(durante.Seconds())))

	bloqueada, _ := db.Conn(ctx)
	go bloqueada.ExecContext(ctx, "UPDATE servidores SET notas = 'reinicio' WHERE nombre = 'db-01'")

	fmt.Println("escenario listo")
	time.Sleep(*durante)
	olvidada.ExecContext(ctx, "ROLLBACK")
	olvidada.Close()
}
