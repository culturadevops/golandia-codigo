#!/bin/bash
# Segunda parte de la demo de Goravel: migración, modelo, controlador y rutas como en Laravel, y lo mismo con net/http + pgx.
# Corre DENTRO de golang:1.26 en la red «goravel», con Postgres 17 en el host «pg». Va después de correr.sh.
set -u
cd /w/app

echo "### 4 · Migración, modelo y controlador (lo que escribes tú)"
MIG=$(ls database/migrations/*_create_productos_table.go)
sed -i 's/table.ID()/table.ID()\n\t\t\ttable.String("nombre")\n\t\t\ttable.Integer("precio")/' "$MIG"
cat > app/models/producto.go <<'EOF'
package models

import (
	"github.com/goravel/framework/database/orm"
)

type Producto struct {
	orm.Model
	Nombre string `json:"nombre"`
	Precio int    `json:"precio"`
}
EOF
cat > app/http/controllers/producto_controller.go <<'EOF'
package controllers

import (
	"github.com/goravel/framework/contracts/http"

	"goravel/app/facades"
	"goravel/app/models"
)

type ProductoController struct{}

func NewProductoController() *ProductoController {
	return &ProductoController{}
}

func (r *ProductoController) Index(ctx http.Context) http.Response {
	var productos []models.Producto
	if err := facades.Orm().Query().OrderByDesc("precio").Find(&productos); err != nil {
		return ctx.Response().Json(http.StatusInternalServerError, http.Json{"error": err.Error()})
	}
	return ctx.Response().Success().Json(productos)
}

func (r *ProductoController) Store(ctx http.Context) http.Response {
	p := models.Producto{Nombre: ctx.Request().Input("nombre"), Precio: ctx.Request().InputInt("precio")}
	if err := facades.Orm().Query().Create(&p); err != nil {
		return ctx.Response().Json(http.StatusInternalServerError, http.Json{"error": err.Error()})
	}
	return ctx.Response().Success().Json(p)
}
EOF
sed -i 's#facades.Route().Get("/users", userController.Index)#facades.Route().Get("/users", userController.Index)\n\n\tproductos := controllers.NewProductoController()\n\tfacades.Route().Get("/productos", productos.Index)\n\tfacades.Route().Post("/productos", productos.Store)#' routes/web.go
echo "\$ cat database/migrations/*_create_productos_table.go | sed -n '/Up Run/,/^}/p'"
sed -n '/Up Run/,/^}/p' "$MIG"
echo "\$ grep -A3 'productos :=' routes/web.go"
grep -A3 'productos :=' routes/web.go

echo
echo "### 5 · Migrar y usar el ORM contra Postgres 17"
echo "\$ ./artisan migrate"
go run . artisan migrate 2>&1 | tail -4
go build -o /tmp/app . || exit 1
/tmp/app > /tmp/app.log 2>&1 &
PID=$!
for i in $(seq 1 100); do curl -s -o /dev/null http://127.0.0.1:3000/users && break; sleep 0.1; done
for p in "Teclado:450" "Monitor:3200" "Mouse:180"; do
  echo "\$ curl -s -X POST localhost:3000/productos -d nombre=${p%%:*} -d precio=${p##*:}"
  curl -s -X POST http://127.0.0.1:3000/productos -d "nombre=${p%%:*}" -d "precio=${p##*:}"; echo
done
echo "\$ curl -s localhost:3000/productos"
curl -s http://127.0.0.1:3000/productos; echo
sleep 1
echo "\$ grep VmRSS /proc/\$PID/status   # memoria del servidor Goravel en marcha"
grep VmRSS /proc/$PID/status
kill $PID

echo
echo "### 6 · Lo mismo sin framework: net/http + pgx"
mkdir -p /w/sin-framework && cd /w/sin-framework
cat > main.go <<'EOF'
package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type Producto struct {
	ID     int64  `json:"id"`
	Nombre string `json:"nombre"`
	Precio int    `json:"precio"`
}

func main() {
	db, err := sql.Open("pgx", "postgres://goravel:goravel@pg:5432/tienda")
	if err != nil {
		panic(err)
	}
	http.HandleFunc("GET /productos", func(w http.ResponseWriter, r *http.Request) {
		filas, err := db.Query("SELECT id, nombre, precio FROM productos ORDER BY precio DESC")
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		defer filas.Close()
		lista := []Producto{}
		for filas.Next() {
			var p Producto
			filas.Scan(&p.ID, &p.Nombre, &p.Precio)
			lista = append(lista, p)
		}
		json.NewEncoder(w).Encode(lista)
	})
	http.HandleFunc("POST /productos", func(w http.ResponseWriter, r *http.Request) {
		p := Producto{Nombre: r.FormValue("nombre")}
		p.Precio, _ = strconv.Atoi(r.FormValue("precio"))
		err := db.QueryRow("INSERT INTO productos (nombre, precio, created_at, updated_at) VALUES ($1, $2, now(), now()) RETURNING id",
			p.Nombre, p.Precio).Scan(&p.ID)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		json.NewEncoder(w).Encode(p)
	})
	http.ListenAndServe(":3001", nil)
}
EOF
go mod init sinframework >/dev/null 2>&1
go get github.com/jackc/pgx/v5@v5.7.6 >/dev/null 2>&1
go mod tidy >/dev/null 2>&1
echo "\$ wc -l main.go"
wc -l main.go
echo "\$ go list -m all | wc -l"
go list -m all | wc -l
echo "\$ time go build -o /tmp/sin ."
{ time -p go build -o /tmp/sin . ; } 2>&1 | grep real
echo "\$ ls -l /tmp/sin"
ls -l /tmp/sin | awk '{print $5, "bytes"}'
/tmp/sin > /tmp/sin.log 2>&1 &
PID=$!
for i in $(seq 1 100); do curl -s -o /dev/null http://127.0.0.1:3001/productos && break; sleep 0.1; done
echo "\$ curl -s localhost:3001/productos"
curl -s http://127.0.0.1:3001/productos
sleep 1
echo "\$ grep VmRSS /proc/\$PID/status   # memoria del servidor net/http en marcha"
grep VmRSS /proc/$PID/status
kill $PID
