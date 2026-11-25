// Lección 15 · Fechas: timestamp contra timestamptz.
//
// La trampa: guardas «las 12:00 en Ciudad de México» en una columna timestamp (sin zona).
// Postgres guarda el reloj de pared y tira la zona. Al leerla, Go no sabe de qué zona era.
package main

import (
	"context"
	"fmt"
	"time"
	_ "time/tzdata" // para que LoadLocation funcione también en Windows y en contenedores mínimos

	"curso-sql/comun"
)

func main() {
	db := comun.Abrir()
	defer db.Close()
	comun.Debe(db, `DROP TABLE IF EXISTS alarmas`)
	comun.Debe(db, `CREATE TABLE alarmas (id int, sin_zona timestamp, con_zona timestamptz)`)

	mx, _ := time.LoadLocation("America/Mexico_City")
	t := time.Date(2026, 9, 28, 12, 0, 0, 0, mx)

	comun.Titulo("15.1 · Guardo las 12:00 de Ciudad de Mexico en las dos columnas")
	fmt.Println("  en Go:      ", t.Format("2006-01-02 15:04 MST (-07:00)"))
	fmt.Println("  en UTC es:  ", t.UTC().Format("2006-01-02 15:04 MST"))
	comun.Debe(db, `INSERT INTO alarmas VALUES (1, $1, $2)`, t, t)

	comun.Titulo("15.2 · Lo que vuelve a Go")
	var a, b time.Time
	db.QueryRow(`SELECT sin_zona, con_zona FROM alarmas WHERE id = 1`).Scan(&a, &b)
	fmt.Println("  timestamp   ->", a.Format("2006-01-02 15:04 MST"))
	fmt.Println("  timestamptz ->", b.UTC().Format("2006-01-02 15:04 MST"), "   =", b.In(mx).Format("15:04 MST"), "en Mexico")
	fmt.Printf("  el mismo instante? %v   diferencia: %v\n", a.Equal(b), b.Sub(a))

	comun.Titulo("15.3 · Lo que ve alguien con psql, segun la zona de SU sesion")
	ctx := context.Background()
	for _, zona := range []string{"UTC", "America/Mexico_City", "America/Bogota"} {
		c, _ := db.Conn(ctx)
		c.ExecContext(ctx, "SET TimeZone = '"+zona+"'")
		var s1, s2 string
		c.QueryRowContext(ctx, `SELECT sin_zona::text, con_zona::text FROM alarmas WHERE id = 1`).Scan(&s1, &s2)
		fmt.Printf("  %-20s timestamp: %s   timestamptz: %s\n", zona, s1, s2)
		c.Close()
	}
}
