// Lección 5: editar una migración que ya se aplicó. goose no lo nota.
//
//	DEMO_BD_URL=postgres://migrador:migrador@localhost:5432/demo05 go run ./lecciones/05_migraciones
//
// Usa su propia base (demo05): crea y borra la tabla «notas».
package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing/fstest"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

const v1 = "-- +goose Up\nCREATE TABLE notas (id int PRIMARY KEY);\n-- +goose Down\nDROP TABLE notas;\n"

// v1Editada es la MISMA migración 00001, editada después de aplicarla: añade una columna.
const v1Editada = "-- +goose Up\nCREATE TABLE notas (id int PRIMARY KEY, texto text NOT NULL DEFAULT '');\n" +
	"-- +goose Down\nDROP TABLE notas;\n"

const v2 = "-- +goose Up\nALTER TABLE notas ADD COLUMN texto text NOT NULL DEFAULT '';\n" +
	"-- +goose Down\nALTER TABLE notas DROP COLUMN texto;\n"

func subir(ctx context.Context, db *sql.DB, archivos fstest.MapFS) {
	p, err := goose.NewProvider(goose.DialectPostgres, db, archivos)
	if err != nil {
		fmt.Println("  error:", err)
		return
	}
	res, err := p.Up(ctx)
	v, _ := p.GetDBVersion(ctx)
	fmt.Printf("  Up -> %d migraciones aplicadas, versión %d, err = %v\n", len(res), v, err)
}

func columna(ctx context.Context, db *sql.DB) {
	_, err := db.ExecContext(ctx, `INSERT INTO notas (id, texto) VALUES (1, 'hola')`)
	if err != nil {
		fmt.Println("  INSERT con texto ->", err)
		return
	}
	fmt.Println("  INSERT con texto -> ok")
}

func main() {
	ctx := context.Background()
	db, err := sql.Open("pgx", os.Getenv("DEMO_BD_URL"))
	if err != nil {
		panic(err)
	}
	defer db.Close()
	_, _ = db.ExecContext(ctx, `DROP TABLE IF EXISTS notas, goose_db_version`)

	fmt.Println("1) se aplica 00001_notas.sql:")
	subir(ctx, db, fstest.MapFS{"00001_notas.sql": {Data: []byte(v1)}})

	fmt.Println("\n2) alguien EDITA 00001_notas.sql para añadir la columna texto, y vuelve a migrar:")
	subir(ctx, db, fstest.MapFS{"00001_notas.sql": {Data: []byte(v1Editada)}})
	columna(ctx, db)

	fmt.Println("\n3) lo correcto: 00001 se queda como estaba y el cambio va en 00002:")
	subir(ctx, db, fstest.MapFS{"00001_notas.sql": {Data: []byte(v1)}, "00002_texto.sql": {Data: []byte(v2)}})
	columna(ctx, db)

	_, _ = db.ExecContext(ctx, `DROP TABLE IF EXISTS notas, goose_db_version`)
}
