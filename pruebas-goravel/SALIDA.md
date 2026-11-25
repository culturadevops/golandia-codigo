# SALIDA — Goravel: ¿el Laravel de Go? (golandia)

> Corrida del 2026-10-19 15:59. Salida literal de correr.sh + parte2.sh dentro de golang:1.26, con Postgres 17.

```
go version go1.26.9 linux/amd64
### 1 · Lo que trae el esqueleto (goravel/goravel v1.18.0)
$ find . -name '*.go' | wc -l
56
$ go mod download && go list -m all | wc -l   # módulos en el árbol de dependencias
240

### 2 · Compilar (caché vacía) y el binario
$ time go build -o /tmp/app .
real 15.46
$ ls -l /tmp/app
89346144 bytes

### 3 · artisan: los generadores de Laravel, en Go
$ ./artisan key:generate
 SUCCESS  Application key set successfully
$ ./artisan make:controller ProductoController
 SUCCESS  Controller created successfully
$ ./artisan make:model Producto
 SUCCESS  Model created successfully
$ ./artisan make:migration create_productos_table
 SUCCESS  Created Migration: create_productos_table
 SUCCESS  Migration registered successfully
$ ls app/http/controllers app/models database/migrations
app/http/controllers:
producto_controller.go
user_controller.go

app/models:
producto.go

database/migrations:
20210101000001_create_jobs_table.go
20261019210006_create_productos_table.go
$ cat app/models/producto.go
package models

import (
	"github.com/goravel/framework/database/orm"
)

type Producto struct {
	orm.Model
}
### 4 · Migración, modelo y controlador (lo que escribes tú)
$ cat database/migrations/*_create_productos_table.go | sed -n '/Up Run/,/^}/p'
// Up Run the migrations.
func (r *M20261019210006CreateProductosTable) Up() error {
	if !facades.Schema().HasTable("productos") {
		return facades.Schema().Create("productos", func(table schema.Blueprint) {
			table.ID()
			table.String("nombre")
			table.Integer("precio")
			table.TimestampsTz()
		})
	}

	return nil
}
$ grep -A3 'productos :=' routes/web.go
	productos := controllers.NewProductoController()
	facades.Route().Get("/productos", productos.Index)
	facades.Route().Post("/productos", productos.Store)
}

### 5 · Migrar y usar el ORM contra Postgres 17
$ ./artisan migrate
  INFO    Running migration
  INFO    Running: 20210101000002_create_jobs_table
  INFO    Running: 20261019210006_create_productos_table
 SUCCESS  Migration success
$ curl -s -X POST localhost:3000/productos -d nombre=Teclado -d precio=450
{"created_at":"2026-10-19 21:00:10","updated_at":"2026-10-19 21:00:10","id":1,"nombre":"Teclado","precio":450}
$ curl -s -X POST localhost:3000/productos -d nombre=Monitor -d precio=3200
{"created_at":"2026-10-19 21:00:10","updated_at":"2026-10-19 21:00:10","id":2,"nombre":"Monitor","precio":3200}
$ curl -s -X POST localhost:3000/productos -d nombre=Mouse -d precio=180
{"created_at":"2026-10-19 21:00:10","updated_at":"2026-10-19 21:00:10","id":3,"nombre":"Mouse","precio":180}
$ curl -s localhost:3000/productos
[{"created_at":"2026-10-19 21:00:10","updated_at":"2026-10-19 21:00:10","id":2,"nombre":"Monitor","precio":3200},{"created_at":"2026-10-19 21:00:10","updated_at":"2026-10-19 21:00:10","id":1,"nombre":"Teclado","precio":450},{"created_at":"2026-10-19 21:00:10","updated_at":"2026-10-19 21:00:10","id":3,"nombre":"Mouse","precio":180}]
$ grep VmRSS /proc/$PID/status   # memoria del servidor Goravel en marcha
VmRSS:	   48280 kB

### 6 · Lo mismo sin framework: net/http + pgx
$ wc -l main.go
50 main.go
$ go list -m all | wc -l
20
$ time go build -o /tmp/sin .
real 1.76
$ ls -l /tmp/sin
14979353 bytes
$ curl -s localhost:3001/productos
[{"id":2,"nombre":"Monitor","precio":3200},{"id":1,"nombre":"Teclado","precio":450},{"id":3,"nombre":"Mouse","precio":180}]
$ grep VmRSS /proc/$PID/status   # memoria del servidor net/http en marcha
VmRSS:	   12640 kB
```
