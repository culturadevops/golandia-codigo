# SALIDA — Estructura de un proyecto Go (golandia)

> Corrida del 2026-10-20 17:31. golang:1.27.2. Fuentes: go.dev/doc/modules/layout (fuente-go.dev-modules-layout.html) y el README de golang-standards/project-layout (fuente-project-layout-README.md).

```
$ go version
go version go1.27.2 linux/amd64

# 1 · Plano: todo en la raíz
$ tree
.
|-- go.mod
|-- main.go
|-- pedidos.go
`-- pedidos_test.go
$ go run .
3 x teclado
$ go test ./...
ok  	github.com/golandia/tienda

# 2 · Servidor: cmd/ + internal/
$ tree
.
|-- cmd
|   |-- api
|   |   `-- main.go
|   `-- migrar
|       `-- main.go
|-- go.mod
`-- internal
    |-- almacen
    |   `-- memoria.go
    `-- pedidos
        |-- pedidos.go
        `-- pedidos_test.go
$ go build ./... && go vet ./...
(sin errores)
$ go test ./...
?   	github.com/golandia/tienda/cmd/api	[no test files]
?   	github.com/golandia/tienda/cmd/migrar	[no test files]
?   	github.com/golandia/tienda/internal/almacen	[no test files]
ok  	github.com/golandia/tienda/internal/pedidos
$ go run ./cmd/api
{1 teclado 3} <nil>
la cantidad debe ser mayor que cero
$ go list ./...
github.com/golandia/tienda/cmd/api
github.com/golandia/tienda/cmd/migrar
github.com/golandia/tienda/internal/almacen
github.com/golandia/tienda/internal/pedidos

# 3 · Otro módulo importa tienda/internal/pedidos
$ go build .
package github.com/otra-empresa/intruso
	main.go:3:8: use of internal package github.com/golandia/tienda/internal/pedidos not allowed

# 3b · internal/ anidado: pedidos/internal/precio, importado desde cmd/api del MISMO módulo
$ tree
.
|-- cmd
|   `-- api
|       `-- main.go
|-- go.mod
`-- pedidos
    |-- internal
    |   `-- precio
    |       `-- precio.go
    `-- pedidos.go
$ go build ./pedidos/
(sin errores)
$ go build ./cmd/api/
package github.com/golandia/anidado/cmd/api
	cmd/api/main.go:6:2: use of internal package github.com/golandia/anidado/pedidos/internal/precio not allowed

# 4 · models importa services y services importa models
$ go build ./...
package github.com/golandia/ciclo/models
	imports github.com/golandia/ciclo/services from usuario.go
	imports github.com/golandia/ciclo/models from usuarios.go: import cycle not allowed
```
