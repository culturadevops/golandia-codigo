// Lección 16 · El proyecto: pgsalud, la revisión de un Postgres a las tres de la mañana.
//
// Todo lo del curso en una herramienta de verdad:
//
//   - abre y hace Ping con timeout (lección 1): si no hay base de datos, sale con 2;
//
//   - cada consulta con context (lección 10) y rows cerrados (lecciones 3 y 5);
//
//   - busca lo que tumba un Postgres: conexiones cerca del máximo (6), consultas que no
//     acaban (10), transacciones olvidadas y sesiones bloqueadas (12);
//
//   - registro a stderr con slog, informe en texto o JSON, y códigos de salida:
//     0 todo bien · 1 hay problemas · 2 no pude ni mirar.
//
//     go run ./16_proyecto -max-consulta 2s -max-idle-tx 2s
package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	"curso-sql/comun"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type Sesion struct {
	PID      int     `json:"pid"`
	Estado   string  `json:"estado"`
	Segundos float64 `json:"segundos"`
	Bloquea  string  `json:"bloqueada_por,omitempty"`
	Consulta string  `json:"consulta"`
}

type Informe struct {
	Conexiones int      `json:"conexiones"`
	Maximo     int      `json:"max_connections"`
	Porcentaje float64  `json:"porcentaje"`
	Lentas     []Sesion `json:"consultas_lentas"`
	Olvidadas  []Sesion `json:"idle_in_transaction"`
	Bloqueadas []Sesion `json:"bloqueadas"`
	TamanoBase string   `json:"tamano_base"`
	Problemas  int      `json:"problemas"`
}

const sesiones = `
SELECT pid, state, extract(epoch FROM now() - %s)::float8,
       coalesce(array_to_string(pg_blocking_pids(pid), ','), ''),
       left(regexp_replace(query, '\s+', ' ', 'g'), 60)
FROM pg_stat_activity
WHERE datname = current_database() AND pid <> pg_backend_pid() AND %s
ORDER BY 3 DESC`

func listar(ctx context.Context, db *sql.DB, desde, donde string, args ...any) ([]Sesion, error) {
	rows, err := db.QueryContext(ctx, fmt.Sprintf(sesiones, desde, donde), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Sesion
	for rows.Next() {
		var s Sesion
		if err := rows.Scan(&s.PID, &s.Estado, &s.Segundos, &s.Bloquea, &s.Consulta); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func revisar(ctx context.Context, db *sql.DB, maxConsulta, maxIdle time.Duration, umbral float64) (*Informe, error) {
	inf := &Informe{}
	err := db.QueryRowContext(ctx, `SELECT (SELECT count(*) FROM pg_stat_activity WHERE backend_type = 'client backend'),
		current_setting('max_connections')::int, pg_size_pretty(pg_database_size(current_database()))`).
		Scan(&inf.Conexiones, &inf.Maximo, &inf.TamanoBase)
	if err != nil {
		return nil, err
	}
	inf.Porcentaje = 100 * float64(inf.Conexiones) / float64(inf.Maximo)
	if inf.Porcentaje >= umbral {
		inf.Problemas++
		slog.Warn("conexiones cerca del maximo", "usadas", inf.Conexiones, "max", inf.Maximo)
	}
	if inf.Lentas, err = listar(ctx, db, "query_start",
		"state = 'active' AND cardinality(pg_blocking_pids(pid)) = 0 AND now() - query_start > $1::interval", maxConsulta.String()); err != nil {
		return nil, err
	}
	if inf.Olvidadas, err = listar(ctx, db, "state_change",
		"state = 'idle in transaction' AND now() - state_change > $1::interval", maxIdle.String()); err != nil {
		return nil, err
	}
	if inf.Bloqueadas, err = listar(ctx, db, "query_start", "cardinality(pg_blocking_pids(pid)) > 0"); err != nil {
		return nil, err
	}
	for _, g := range []struct {
		n   string
		ses []Sesion
	}{{"consultas lentas", inf.Lentas}, {"transacciones olvidadas", inf.Olvidadas}, {"sesiones bloqueadas", inf.Bloqueadas}} {
		if len(g.ses) > 0 {
			inf.Problemas += len(g.ses)
			slog.Warn(g.n, "cuantas", len(g.ses))
		}
	}
	return inf, nil
}

func imprimir(inf *Informe) {
	fmt.Printf("Conexiones: %d de %d (%.0f%%) · base de datos: %s\n", inf.Conexiones, inf.Maximo, inf.Porcentaje, inf.TamanoBase)
	grupo := func(titulo string, ss []Sesion) {
		if len(ss) == 0 {
			fmt.Printf("  %-24s ninguna\n", titulo)
			return
		}
		fmt.Printf("  %-24s %d\n", titulo, len(ss))
		for _, s := range ss {
			extra := ""
			if s.Bloquea != "" {
				extra = " · la bloquea el pid " + s.Bloquea
			}
			fmt.Printf("    pid %-5d %5.1f s%s  %s\n", s.PID, s.Segundos, extra, s.Consulta)
		}
	}
	grupo("Consultas lentas:", inf.Lentas)
	grupo("Transacciones olvidadas:", inf.Olvidadas)
	grupo("Sesiones bloqueadas:", inf.Bloqueadas)
}

func main() {
	dsn := flag.String("dsn", comun.DSN(), "cadena de conexion")
	comoJSON := flag.Bool("json", false, "informe en JSON")
	maxConsulta := flag.Duration("max-consulta", 30*time.Second, "una consulta activa mas de esto es lenta")
	maxIdle := flag.Duration("max-idle-tx", time.Minute, "una transaccion quieta mas de esto esta olvidada")
	umbral := flag.Float64("umbral-conexiones", 80, "porcentaje de max_connections que ya es problema")
	flag.Parse()
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				return slog.Attr{}
			}
			return a
		}})))

	db, _ := sql.Open("pgx", *dsn)
	defer db.Close()
	db.SetMaxOpenConns(2)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		slog.Error("no puedo conectar", "err", err)
		os.Exit(2)
	}
	inf, err := revisar(ctx, db, *maxConsulta, *maxIdle, *umbral)
	if err != nil {
		slog.Error("no puedo revisar", "err", err)
		os.Exit(2)
	}
	if *comoJSON {
		e := json.NewEncoder(os.Stdout)
		e.SetIndent("", "  ")
		e.Encode(inf)
	} else {
		imprimir(inf)
	}
	if inf.Problemas > 0 {
		slog.Warn("hay problemas", "total", inf.Problemas)
		os.Exit(1)
	}
}
