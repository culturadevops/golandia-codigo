# SALIDA — Un proyecto de portafolio, antes y después (golandia)

> Corrida del 2026-10-20 12:45. golang:1.26, golangci-lint v2.15.0, Docker.

```
go version go1.26.9 linux/amd64
### ANTES: el escáner del vídeo 239, tal cual
$ find . -name '*.go' | xargs wc -l
113 ./main.go
$ go test ./...
?   	escaner	[no test files]
$ go vet ./...
(sin avisos)

### DESPUÉS: lógica en un paquete, pruebas, contexto
$ find . -name '*.go' | xargs wc -l
  31 ./cmd/escaner/main.go
  82 ./escaner/escaner.go
  90 ./escaner/escaner_test.go
 203 total
$ go vet ./...
(sin avisos)
$ go test -race -cover ./...
	escaner/cmd/escaner		coverage: 0.0% of statements
ok  	escaner/escaner	1.132s	coverage: 100.0% of statements
$ go test -race -v -run 'Test' ./escaner | grep -E '^(=== RUN|--- )'
--- PASS: TestEncuentraLosPuertosAbiertos (0.09s)
--- PASS: TestValidaLosParametros (0.00s)
    --- PASS: TestValidaLosParametros/concurrencia_cero (0.00s)
    --- PASS: TestValidaLosParametros/rango_al_revés (0.00s)
    --- PASS: TestValidaLosParametros/puerto_imposible (0.00s)
--- PASS: TestSeCancelaConElContexto (0.00s)
--- PASS: TestClasificar (0.00s)

$ go build -o /tmp/escaner ./cmd/escaner && /tmp/escaner -hasta 100
0 abiertos en 2ms: []
errores: map[rechazado:100]

### golangci-lint (v2.15.0), antes y después
$ golangci-lint run ./...   # antes
main.go:61:11: Error return value of `c.Close` is not checked (errcheck)
			c.Close()
			       ^
1 issues:
* errcheck: 1
$ golangci-lint run ./...   # después
0 issues.

### La imagen
$ docker build -t escaner . && docker image ls escaner
1.0 MiB (descarga comprimida)
$ docker run --rm escaner -hasta 10
0 abiertos en 1ms: []
errores: map[rechazado:10]
```
