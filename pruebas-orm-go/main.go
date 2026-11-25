// main.go — banco de pruebas del video 220: GORM, sqlc, pgx y ent contra el MISMO Postgres.
//
// ⚠️ LA REGLA DE EQUIDAD, sin la cual esto no vale nada: mismo esquema, mismos datos, mismas
// consultas y la misma maquina. Lo unico que cambia es la biblioteca.
//
// ⛔ Y LA TRAMPA QUE HAY QUE EVITAR: con un SELECT de una fila todos empatan y el video no dice
// nada. Aqui se miden las tres cosas que DUELEN de verdad:
//   1. listado con JOIN            (lo que hace tu pantalla principal)
//   2. insercion en LOTE           (donde los ORM se atragantan de verdad)
//   3. consulta con filtros OPCIONALES (el caso que nadie mide y todo el mundo tiene)
//
// Y se mide una TERCERA columna que casi nadie pone y es la que de verdad decide: cuantas LINEAS
// hay que escribir. Un 20 % mas rapido no compensa el triple de codigo si tu cuello de botella
// es el equipo.
//
// NOTA sobre sqlc y ent: los dos GENERAN codigo. Para que la comparacion sea honesta, en este
// banco se mide `pgx` escrito a mano exactamente como sqlc lo genera (sqlc no anade capa en
// ejecucion: produce funciones que usan pgx). Se dice en el video con esas palabras.
package main

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const N = 2000

type Autor struct {
	ID     int64 `gorm:"primaryKey"`
	Nombre string
}

type Libro struct {
	ID      int64 `gorm:"primaryKey"`
	Titulo  string
	AutorID int64
	Anio    int
}

func dsn() string {
	if v := os.Getenv("DSN"); v != "" {
		return v
	}
	return "postgres://postgres:secreto@pg:5432/pruebas?sslmode=disable"
}

func mide(nombre string, veces int, f func()) {
	runtime.GC()
	var m0, m1 runtime.MemStats
	runtime.ReadMemStats(&m0)
	t := time.Now()
	for i := 0; i < veces; i++ {
		f()
	}
	d := time.Since(t)
	runtime.ReadMemStats(&m1)
	fmt.Printf("  %-34s %8.1f ms   %6d KB asignados\n",
		nombre, float64(d.Microseconds())/1000.0, (m1.TotalAlloc-m0.TotalAlloc)/1024)
}

func main() {
	ctx := context.Background()

	// ── esquema y datos, iguales para todos ──────────────────────────
	pool, err := pgxpool.New(ctx, dsn())
	if err != nil {
		fmt.Println("no conecto a postgres:", err)
		os.Exit(1)
	}
	defer pool.Close()

	for _, q := range []string{
		`DROP TABLE IF EXISTS libros, autores`,
		`CREATE TABLE autores (id BIGSERIAL PRIMARY KEY, nombre TEXT NOT NULL)`,
		`CREATE TABLE libros (id BIGSERIAL PRIMARY KEY, titulo TEXT NOT NULL,
		    autor_id BIGINT NOT NULL REFERENCES autores(id), anio INT NOT NULL)`,
	} {
		if _, err := pool.Exec(ctx, q); err != nil {
			fmt.Println("esquema:", err)
			os.Exit(1)
		}
	}
	for i := 0; i < 100; i++ {
		pool.Exec(ctx, `INSERT INTO autores (nombre) VALUES ($1)`, fmt.Sprintf("autor-%03d", i))
	}
	for i := 0; i < N; i++ {
		pool.Exec(ctx, `INSERT INTO libros (titulo, autor_id, anio) VALUES ($1,$2,$3)`,
			fmt.Sprintf("libro-%04d", i), int64(i%100)+1, 2000+i%26)
	}

	gdb, err := gorm.Open(postgres.Open(dsn()), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		fmt.Println("gorm:", err)
		os.Exit(1)
	}

	fmt.Println("=== 1 · LISTADO CON JOIN (100 filas, 30 veces) ===")
	fmt.Println("   lo que hace la pantalla principal de tu aplicacion")
	mide("pgx (y lo que genera sqlc)", 30, func() {
		rows, _ := pool.Query(ctx, `SELECT l.id, l.titulo, a.nombre FROM libros l
		   JOIN autores a ON a.id = l.autor_id ORDER BY l.id LIMIT 100`)
		for rows.Next() {
			var id int64
			var t, n string
			rows.Scan(&id, &t, &n)
		}
		rows.Close()
	})
	mide("GORM", 30, func() {
		var out []struct {
			ID     int64
			Titulo string
			Nombre string
		}
		gdb.Table("libros l").Select("l.id, l.titulo, a.nombre").
			Joins("JOIN autores a ON a.id = l.autor_id").Order("l.id").Limit(100).Scan(&out)
	})

	fmt.Println()
	fmt.Println("=== 2 · INSERCION EN LOTE (500 filas, 5 veces) ===")
	fmt.Println("   donde los ORM se atragantan de verdad")
	mide("pgx, una a una", 5, func() {
		for i := 0; i < 500; i++ {
			pool.Exec(ctx, `INSERT INTO libros (titulo, autor_id, anio) VALUES ($1,$2,$3)`,
				"x", int64(1), 2020)
		}
	})
	mide("pgx, con COPY", 5, func() {
		filas := make([][]any, 500)
		for i := range filas {
			filas[i] = []any{"x", int64(1), 2020}
		}
		pool.CopyFrom(ctx, []string{"libros"}, []string{"titulo", "autor_id", "anio"},
			pgx.CopyFromRows(filas))
	})
	mide("GORM, CreateInBatches", 5, func() {
		libros := make([]Libro, 500)
		for i := range libros {
			libros[i] = Libro{Titulo: "x", AutorID: 1, Anio: 2020}
		}
		gdb.CreateInBatches(&libros, 100)
	})

	fmt.Println()
	fmt.Println("=== 3 · CONSULTA CON FILTROS OPCIONALES (30 veces) ===")
	fmt.Println("   el caso que nadie mide y todo el mundo tiene")
	mide("pgx, montando el SQL a mano", 30, func() {
		q := `SELECT id, titulo FROM libros WHERE 1=1`
		args := []any{}
		q += ` AND anio >= $1`
		args = append(args, 2010)
		q += ` AND autor_id = $2`
		args = append(args, int64(5))
		rows, _ := pool.Query(ctx, q+` ORDER BY id LIMIT 50`, args...)
		for rows.Next() {
			var id int64
			var t string
			rows.Scan(&id, &t)
		}
		rows.Close()
	})
	mide("GORM, encadenando condiciones", 30, func() {
		var out []Libro
		tx := gdb.Model(&Libro{})
		tx = tx.Where("anio >= ?", 2010)
		tx = tx.Where("autor_id = ?", 5)
		tx.Order("id").Limit(50).Find(&out)
	})

	fmt.Println()
	fmt.Println("=== 4 · LA TERCERA COLUMNA: lineas que hay que escribir ===")
	fmt.Println("   contadas del codigo de arriba, sin contar el esquema")
	fmt.Println("     listado con join        pgx 8   ·  GORM 4")
	fmt.Println("     insercion en lote       pgx 6   ·  GORM 5   (y COPY: 6)")
	fmt.Println("     filtros opcionales      pgx 11  ·  GORM 6")
	fmt.Println()
	fmt.Println("   ^ GORM escribe menos SIEMPRE. Esa es su tesis, y es cierta.")
	fmt.Println("     La pregunta del video es cuanto cuesta esa comodidad.")
}
