# SALIDA — go work en Go (ejecutado)

Generado por `python correr.py`. **No se edita a mano.**

- fecha: 2026-10-10 17:59 UTC
- imagen: golang:1.27.2 (Docker, linux/amd64)
- máquina: 24 · Intel(R) Core(TM) Ultra 7 270K Plus núcleos/CPU
- los tiempos son de ESTA máquina, dentro de Docker en Windows.

## La versión de Go

`$ go version`

```
go version go1.27.2 linux/amd64
```

`$ bash preparar.sh   # dos módulos: app importa saludo`

```
repo listo
```

`$ find . -name '*.go' -o -name go.mod && cat app/go.mod`

```
./app/go.mod
./app/main.go
./saludo/go.mod
./saludo/saludo.go
module ejemplo.com/app

go 1.27

require ejemplo.com/saludo v0.1.0
```

## 1 · Sin nada

`$ cd app && go build ./...`

```
main.go:6:2: missing go.sum entry for module providing package ejemplo.com/saludo (imported by ejemplo.com/app); to add:
	go get ejemplo.com/app
```

## 2 · Con replace

`$ go mod edit -replace ejemplo.com/saludo=../saludo && go run .`

```
Hola, Golandia
```

`$ git diff app/go.mod`

```
require ejemplo.com/saludo v0.1.0
+
+replace ejemplo.com/saludo => ../saludo
```

## 3 · La CI, que solo clona app

`$ (en la CI) go build ./...`

```
main.go:6:2: ejemplo.com/saludo@v0.1.0: replacement directory ../saludo does not exist
```

`$ git checkout app/go.mod && git status --short`

```
(limpio)
```

## 4 · Con go work

`$ go work init ./app ./saludo && cat go.work`

```
go 1.27.2

use (
	./app
	./saludo
)
```

`$ go run ./app && go env GOWORK`

```
Hola, Golandia
/w/repo/go.work
```

`$ (cambio el saludo en saludo/saludo.go) && go run ./app`

```
¡Hola, Golandia
```

## 5 · GOWORK=off y lo que queda en git

`$ cd app && GOWORK=off go build ./...`

```
main.go:6:2: missing go.sum entry for module providing package ejemplo.com/saludo (imported by ejemplo.com/app); to add:
```

`$ git status --short`

```
M saludo/saludo.go
?? go.work
```

## 6 · Los módulos del workspace

`$ go list -m`

```
ejemplo.com/app
ejemplo.com/saludo
```

`$ go work edit -json`

```
{
	"Go": "1.27.2",
	"Use": [
		{
			"DiskPath": "./app"
		},
		{
			"DiskPath": "./saludo"
		}
	]
}
```

## 7 · Un tercer módulo: go work use

`$ mkdir util && (go.mod + util.go)
$ go work use ./util && cat go.work && go list -m`

```
go 1.27.2

use (
	./app
	./saludo
	./util
)
ejemplo.com/app
ejemplo.com/saludo
ejemplo.com/util
```

`$ go work use -r .   # busca todos los go.mod por debajo`

```
go 1.27.2

use (
	./app
	./herramientas/cli
	./saludo
	./util
)
```

## 8 · ./... desde la raíz del workspace

`$ go build ./...`

```
pattern ./...: directory prefix . does not contain modules listed in go.work or their selected dependencies
exit=0
```

`$ go vet ./app/... ./saludo/...`

```
vet: sin problemas
```

## 9 · go mod tidy dentro de un workspace

`$ cd app && go mod tidy`

```
go: downloading ejemplo.com/saludo v0.1.0
go: ejemplo.com/app imports
	ejemplo.com/saludo: module lookup disabled by GOPROXY=off
```

## 10 · replace en go.work, sobre un módulo del propio workspace

`$ go work edit -replace ejemplo.com/saludo=./saludo && grep replace go.work
$ git diff --stat app/go.mod`

```
replace ejemplo.com/saludo => ./saludo
(app/go.mod sin cambios)
```

`$ go list -m`

```
go: workspace module ejemplo.com/saludo is replaced at all versions in the go.work file. To fix, remove the replacement from the go.work file or specify the version at which to replace the module.
```

`$ go work edit -dropreplace ejemplo.com/saludo && go list -m`

```
ejemplo.com/app
ejemplo.com/cli
ejemplo.com/saludo
ejemplo.com/util
```

`$ ls go.work*`

```
go.work
```

## 11 · Desde una subcarpeta también lo encuentra

`$ cd app && go env GOWORK
$ cd ../saludo && go env GOWORK`

```
/w/repo/go.work
/w/repo/go.work
```

## 12 · Quitar un módulo: -dropuse

`$ go work edit -dropuse ./herramientas/cli && go work edit -dropuse ./util && go list -m`

```
ejemplo.com/app
ejemplo.com/saludo
```

## 13 · Otro fichero de workspace con GOWORK

`$ cp go.work /tmp/solo.work && GOWORK=/tmp/solo.work go list -m`

```
go: cannot load module /tmp/app listed in go.work file: open /tmp/app/go.mod: no such file or directory
go: cannot load module /tmp/saludo listed in go.work file: open /tmp/saludo/go.mod: no such file or directory
```

`$ cp go.work solo-app.work && GOWORK=$PWD/solo-app.work go work edit -dropuse ./saludo
$ GOWORK=$PWD/solo-app.work go list -m
$ go list -m`

```
ejemplo.com/app
--- y el go.work normal sigue igual:
ejemplo.com/app
ejemplo.com/saludo
```

## 14 · go test en dos módulos a la vez

`$ go test ./app/... ./saludo/...`

```
?   	ejemplo.com/app	[no test files]
ok  	ejemplo.com/saludo	(cached)
```
