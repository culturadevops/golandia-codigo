# SALIDA — build tags — ampliación (ejecutado)

Generado por `python correr.py`. **No se edita a mano.**

- fecha: 2026-10-11 02:59 UTC
- imagen: golang:1.27.2 (Docker, linux/amd64)
- máquina: 24 · Intel(R) Core(TM) Ultra 7 270K Plus núcleos/CPU
- los tiempos son de ESTA máquina, dentro de Docker en Windows.

## 1 · El binario recuerda sus etiquetas

`$ go build -o libre . && go build -tags pro -o pro .
$ go version -m libre;  go version -m pro`

```
== libre
	build	GOARCH=amd64
	build	GOOS=linux
== pro
	build	-tags=pro
	build	GOARCH=amd64
	build	GOOS=linux
```

## 2 · Los tests también llevan sufijo

`$ go test -v .     # en Linux
$ GOOS=windows go test -c -o ruta.test.exe .`

```
=== RUN   TestRutaLinux
--- PASS: TestRutaLinux (0.00s)
ok  	ejemplo.com/app	0.002s
4710400 ruta.test.exe
go vet para windows: sin quejas
```

## 3 · La librería estándar

`$ cd $(go env GOROOT)/src && grep -rl '^//go:build' . | wc -l`

```
ficheros .go: 6827 · con //go:build: 1863
en os/: 10 _windows.go · 6 _linux.go · 3 _darwin.go
```

## 4 · Un cuarto destino: WebAssembly

`$ GOOS=wasip1 GOARCH=wasm go build -o app.wasm .`

```
# ejemplo.com/app
./main.go:6:36: undefined: rutaConfig
ficheros que entran: [edicion_libre.go main.go]
```

`$ cat ruta_otros.go && GOOS=wasip1 GOARCH=wasm go build -o app.wasm .`

```
//go:build !linux && !darwin && !windows

package main

func rutaConfig() string { return "config.toml" } // cualquier otro sistema
2475608 app.wasm
ficheros que entran: [edicion_libre.go main.go ruta_otros.go]
y en linux:          [edicion_libre.go main.go ruta_linux.go]
```
