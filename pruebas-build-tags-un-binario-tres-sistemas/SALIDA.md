# SALIDA — Build tags en Go (ejecutado)

Generado por `python correr.py`. **No se edita a mano.**

- fecha: 2026-10-10 17:56 UTC
- imagen: golang:1.27.2 (Docker, linux/amd64)
- máquina: 24 · Intel(R) Core(TM) Ultra 7 270K Plus núcleos/CPU
- los tiempos son de ESTA máquina, dentro de Docker en Windows.

## La versión de Go

`$ go version`

```
go version go1.27.2 linux/amd64
```

`$ ls *.go && head -4 edicion_pro.go`

```
edicion_libre.go
edicion_pro.go
generar.go
main.go
ruta_darwin.go
ruta_linux.go
ruta_windows.go
//go:build pro

package main
```

## 1 · Qué ficheros entran en cada sistema

`$ for s in linux windows darwin; do GOOS=$s go list -f '{{.GoFiles}}' .; done`

```
linux: [edicion_libre.go main.go ruta_linux.go]
windows: [edicion_libre.go main.go ruta_windows.go]
darwin: [edicion_libre.go main.go ruta_darwin.go]
```

## 2 · Tres binarios desde la misma máquina

`$ for s in linux windows darwin; do GOOS=$s GOARCH=amd64 go build -o app-$s .; done
$ head -c 4 app-* | od -An -tx1`

```
app-linux:  7f 45 4c 46
app-windows:  4d 5a 90 00
app-darwin:  cf fa ed fe
```

`$ ./app-linux`

```
configuración en: /etc/app/config.toml
edición: libre
```

## 3 · Una etiqueta propia: -tags pro

`$ go list -tags pro -f '{{.GoFiles}}' . && go build -tags pro -o app-pro . && ./app-pro`

```
[edicion_pro.go main.go ruta_linux.go]
configuración en: /etc/app/config.toml
edición: pro
```

## 4 · //go:build ignore

`$ head -1 generar.go && go list -f '{{.IgnoredGoFiles}}' . && go run generar.go`

```
//go:build ignore
[edicion_pro.go generar.go ruta_darwin.go ruta_windows.go]
soy una herramienta aparte; solo corro con go run generar.go
```

## 5 · La trampa: el //go:build después del package

`$ cat edicion_pro.go   # el //go:build, debajo del package
$ go build .`

```
package main

//go:build pro

const edicion = "pro"
# ejemplo.com/app
./edicion_pro.go:5:7: edicion redeclared in this block
	./edicion_libre.go:5:7: other declaration of edicion
```

## 6 · La sintaxis vieja // +build

`$ cat viejo.go   # // +build linux
$ gofmt -d viejo.go`

```
diff viejo.go.orig viejo.go
--- viejo.go.orig
+++ viejo.go
@@ -1,3 +1,4 @@
+//go:build linux
 // +build linux
 
 package main
```

## 7 · Cuántas plataformas conoce Go

`$ go tool dist list | wc -l
$ go tool dist list | grep -c '^linux/'
$ go tool dist list | grep -E '^(windows|darwin)/'`

```
47
13
darwin/amd64 darwin/arm64 windows/386 windows/amd64 windows/arm64
```

## 8 · La arquitectura: amd64 contra arm64

`$ GOOS=linux GOARCH=arm64 go build -o app-arm64 .
$ od -An -tx1 -j18 -N2 app-amd64 app-arm64   # el campo «máquina» de ELF`

```
app-amd64: magia  7f 45 4c 46 · máquina  3e 00
app-arm64: magia  7f 45 4c 46 · máquina  b7 00
```

## 9 · Lo que pesa cada binario

`$ ls -l app-*`

```
2500448 app-darwin
2342361 app-linux
2492928 app-windows
```

## 10 · Tests con etiqueta: //go:build integration

`$ head -1 disco_test.go   # //go:build integration
$ go test -v .
$ go test -v -tags integration .`

```
--- go test -v .
=== RUN   TestRuta
--- PASS: TestRuta (0.00s)
PASS
ok  	ejemplo.com/app	(cached)
--- go test -v -tags integration .
=== RUN   TestContraElDisco
    disco_test.go:8: esto solo corre con -tags integration
--- PASS: TestContraElDisco (0.00s)
=== RUN   TestRuta
--- PASS: TestRuta (0.00s)
PASS
ok  	ejemplo.com/app	(cached)
```

## 11 · Dos etiquetas a la vez: -tags pro,debug

`$ head -1 traza.go   # //go:build pro && debug
$ go list -tags pro -f '{{.GoFiles}}' .
$ go list -tags pro,debug -f '{{.GoFiles}}' .
$ go run -tags pro,debug .`

```
-tags pro:       [edicion_pro.go main.go ruta_linux.go]
-tags pro,debug: [edicion_pro.go main.go ruta_linux.go traza.go]
configuración en: /etc/app/config.toml
edición: pro
traza activada
```

## 12 · cgo, unix y la versión de Go como condición

`$ go list con //go:build cgo, //go:build unix, //go:build go1.99 y //go:build go1.27`

```
CGO_ENABLED=1:  [actual.go con_cgo.go edicion_libre.go main.go ruta_linux.go solo_unix.go]
CGO_ENABLED=0:  [actual.go edicion_libre.go main.go ruta_linux.go solo_unix.go]
linux: solo_unix.go
darwin: solo_unix.go
windows: (sin solo_unix.go)
go1.99: futuro.go  ← ignorado
go1.27: actual.go  ← entra
```

`$ go env CGO_ENABLED; GOOS=windows go env CGO_ENABLED; GOARCH=arm64 go env CGO_ENABLED`

```
nativo:        CGO_ENABLED=1
GOOS=windows:  CGO_ENABLED=0
GOARCH=arm64:  CGO_ENABLED=0
```

## 13 · go vet ve la trampa

`$ go vet -tags pro .   # con el //go:build debajo del package`

```
edicion_pro.go:3:1: misplaced //go:build comment
```

## 14 · La librería estándar usa lo mismo

`$ cd $(go env GOROOT)/src/os && ls *_windows.go *_unix.go
$ GOOS=linux go list -f '{{len .GoFiles}}' os
$ GOOS=windows go list -f '{{len .GoFiles}}' os`

```
ficheros en os: 119
_windows.go: 10
_unix.go: 13
linux:   45 ficheros
windows: 34 ficheros
```

## 15 · Si te olvidas de un sistema

`$ mv ruta_darwin.go /tmp/   # me olvido de Mac
$ GOOS=darwin go build .`

```
# ejemplo.com/app
./main.go:6:36: undefined: rutaConfig
```

`$ GOOS=linux go build .   # en tu máquina no se nota`

```
linux: compila sin quejarse
```

## 16 · GOFLAGS: dejar la etiqueta puesta

`$ GOFLAGS=-tags=pro go run .`

```
configuración en: /etc/app/config.toml
edición: pro
```

## 17 · La alternativa: -ldflags -X cambia un valor, no el código

`$ cat main.go   # var version = "dev"
$ go run .
$ go run -ldflags '-X main.version=1.4.2' .`

```
versión: dev
versión: 1.4.2
```

## 18 · runtime.GOOS contra las etiquetas

`$ cat main.go   # if runtime.GOOS == "windows" { ... }
$ go run .
$ GOOS=windows go build .`

```
en linux
windows: compila
```

`$ cat main.go   # syscall.LoadDLL dentro de un if runtime.GOOS
$ GOOS=windows go build .
$ GOOS=linux go build .`

```
windows: compila
# s
./main.go:11:21: undefined: syscall.LoadDLL
```
