# SALIDA — testcontainers-go

Generado por `python ejecutar_todo.py`. **No se edita a mano.**

- fecha: 2026-10-01 22:02 UTC
- go version go1.26.5 windows/amd64 · Docker 29.6.2 · testcontainers-go v0.44.0
- Windows 10
- imágenes ya descargadas antes de medir (postgres:17-alpine, redis:7-alpine)

## 1 · el doble en memoria

`$ go test -count=1 -run ^TestMem$ -v ./usuarios` → código 1 · 3.7 s de reloj

```
=== RUN   TestMem
=== RUN   TestMem/crear_y_contar
=== RUN   TestMem/duplicado
=== RUN   TestMem/correo_sin_arroba
    usuarios_test.go:37: aceptó un correo sin arroba
--- FAIL: TestMem (0.00s)
    --- PASS: TestMem/crear_y_contar (0.00s)
    --- PASS: TestMem/duplicado (0.00s)
    --- FAIL: TestMem/correo_sin_arroba (0.00s)
FAIL
FAIL	github.com/culturadevops/pruebas-testcontainers/usuarios	0.193s
FAIL
```

## 2 · la misma batería contra Postgres de verdad

`$ go test -count=1 -run ^TestPG$ -v ./usuarios` → código 0 · 2.6 s de reloj

```
=== RUN   TestPG
2026/10/01 17:02:07 github.com/testcontainers/testcontainers-go - Connected to docker: 
    usuarios_test.go:74: contenedor listo en 1.62s
=== RUN   TestPG/crear_y_contar
=== RUN   TestPG/duplicado
=== RUN   TestPG/correo_sin_arroba
--- PASS: TestPG (1.88s)
    --- PASS: TestPG/crear_y_contar (0.00s)
    --- PASS: TestPG/duplicado (0.00s)
    --- PASS: TestPG/correo_sin_arroba (0.00s)
PASS
ok  	github.com/culturadevops/pruebas-testcontainers/usuarios	2.077s
```

## 3 · cinco pruebas, un contenedor por prueba

`$ go test -count=1 -run ^TestCincoPorPrueba$ -v ./usuarios` → código 0 · 8.0 s de reloj

```
=== RUN   TestCincoPorPrueba
=== RUN   TestCincoPorPrueba/1
2026/10/01 17:02:10 github.com/testcontainers/testcontainers-go - Connected to docker: 
    usuarios_test.go:84: contenedor listo en 1.46s
=== RUN   TestCincoPorPrueba/2
    usuarios_test.go:84: contenedor listo en 1.11s
=== RUN   TestCincoPorPrueba/3
    usuarios_test.go:84: contenedor listo en 1.1s
=== RUN   TestCincoPorPrueba/4
    usuarios_test.go:84: contenedor listo en 1.11s
=== RUN   TestCincoPorPrueba/5
    usuarios_test.go:84: contenedor listo en 1.13s
--- PASS: TestCincoPorPrueba (7.32s)
    --- PASS: TestCincoPorPrueba/1 (1.77s)
    --- PASS: TestCincoPorPrueba/2 (1.37s)
    --- PASS: TestCincoPorPrueba/3 (1.36s)
    --- PASS: TestCincoPorPrueba/4 (1.43s)
    --- PASS: TestCincoPorPrueba/5 (1.39s)
PASS
ok  	github.com/culturadevops/pruebas-testcontainers/usuarios	7.514s
```

## 4 · cinco pruebas, un contenedor compartido

`$ go test -count=1 -run ^TestCincoCompartido$ -v ./usuarios` → código 0 · 2.5 s de reloj

```
=== RUN   TestCincoCompartido
2026/10/01 17:02:18 github.com/testcontainers/testcontainers-go - Connected to docker: 
    usuarios_test.go:100: contenedor listo en 1.47s
=== RUN   TestCincoCompartido/1
=== RUN   TestCincoCompartido/2
=== RUN   TestCincoCompartido/3
=== RUN   TestCincoCompartido/4
=== RUN   TestCincoCompartido/5
--- PASS: TestCincoCompartido (1.78s)
    --- PASS: TestCincoCompartido/1 (0.00s)
    --- PASS: TestCincoCompartido/2 (0.00s)
    --- PASS: TestCincoCompartido/3 (0.00s)
    --- PASS: TestCincoCompartido/4 (0.00s)
    --- PASS: TestCincoCompartido/5 (0.00s)
PASS
ok  	github.com/culturadevops/pruebas-testcontainers/usuarios	1.963s
```

## 5 · esperar el log de Postgres UNA vez

`$ go test -count=1 -run ^TestEsperaUnaVez$ -v ./usuarios` → código 0 · 56.6 s de reloj

```
=== RUN   TestEsperaUnaVez
2026/10/01 17:02:20 github.com/testcontainers/testcontainers-go - Connected to docker: 
    usuarios_test.go:144: intento 1: failed to connect to `user=postgres database=postgres`:
        	[::1]:32779 (localhost): failed to receive message: unexpected EOF
        	127.0.0.1:32779 (localhost): failed to receive message: unexpected EOF
    usuarios_test.go:144: intento 2: failed to connect to `user=postgres database=postgres`:
        	[::1]:32780 (localhost): failed to receive message: unexpected EOF
        	127.0.0.1:32780 (localhost): failed to receive message: unexpected EOF
    usuarios_test.go:144: intento 3: failed to connect to `user=postgres database=postgres`:
        	[::1]:32781 (localhost): failed to receive message: unexpected EOF
        	127.0.0.1:32781 (localhost): failed to receive message: unexpected EOF
    usuarios_test.go:144: intento 4: failed to connect to `user=postgres database=postgres`:
        	[::1]:32782 (localhost): failed to receive message: unexpected EOF
        	127.0.0.1:32782 (localhost): failed to receive message: unexpected EOF
    usuarios_test.go:144: intento 5: failed to connect to `user=postgres database=postgres`:
        	[::1]:32783 (localhost): failed to receive message: unexpected EOF
        	127.0.0.1:32783 (localhost): failed to receive message: unexpected EOF
    usuarios_test.go:148: esperando el log 1 vez/veces: 5 de 5 conexiones fallaron
--- PASS: TestEsperaUnaVez (55.82s)
PASS
ok  	github.com/culturadevops/pruebas-testcontainers/usuarios	56.010s
```

## 6 · esperar el log de Postgres DOS veces

`$ go test -count=1 -run ^TestEsperaDosVeces$ -v ./usuarios` → código 0 · 7.4 s de reloj

```
=== RUN   TestEsperaDosVeces
2026/10/01 17:03:17 github.com/testcontainers/testcontainers-go - Connected to docker: 
    usuarios_test.go:148: esperando el log 2 vez/veces: 0 de 5 conexiones fallaron
--- PASS: TestEsperaDosVeces (6.71s)
PASS
ok  	github.com/culturadevops/pruebas-testcontainers/usuarios	6.911s
```

## 7 · el proceso muere sin limpiar · con Ryuk (por defecto)

`$ go run ./cmd/huerfano` → código 1 · 0.8 s de reloj

```
contenedor arrancado: 06f6b04a3d7d
2026/10/01 17:03:24 github.com/testcontainers/testcontainers-go - Connected to docker: 
exit status 1
```

```
docker ps -a --filter label=org.testcontainers=true

justo al morir el proceso:
671ec18b8204 testcontainers/ryuk:0.14.0 Up Less than a second
06f6b04a3d7d redis:7-alpine Up Less than a second
84e4fcd59567 testcontainers/ryuk:0.14.0 Up 7 seconds
2b3b9e131bc2 testcontainers/ryuk:0.14.0 Up About a minute

20 segundos después:
(ninguno)
```

## 7 · el proceso muere sin limpiar · sin Ryuk

`$ TESTCONTAINERS_RYUK_DISABLED=true go run ./cmd/huerfano` → código 1 · 0.5 s de reloj

```
contenedor arrancado: 783c39847eaf
2026/10/01 17:03:45 github.com/testcontainers/testcontainers-go - Connected to docker: 
exit status 1
```

```
docker ps -a --filter label=org.testcontainers=true

justo al morir el proceso:
783c39847eaf redis:7-alpine Up Less than a second

20 segundos después:
783c39847eaf redis:7-alpine Up 20 seconds
```

## 8a · DOCKER_HOST apuntando a un sitio que no responde

`$ DOCKER_HOST=tcp://127.0.0.1:1 go test -count=1 -run ^TestPG$ -v ./usuarios` → código 0 · 2.5 s de reloj

```
=== RUN   TestPG
2026/10/01 17:04:06 github.com/testcontainers/testcontainers-go - Connected to docker: 
    usuarios_test.go:74: contenedor listo en 1.47s
=== RUN   TestPG/crear_y_contar
=== RUN   TestPG/duplicado
=== RUN   TestPG/correo_sin_arroba
--- PASS: TestPG (1.74s)
    --- PASS: TestPG/crear_y_contar (0.00s)
    --- PASS: TestPG/duplicado (0.00s)
    --- PASS: TestPG/correo_sin_arroba (0.00s)
PASS
ok  	github.com/culturadevops/pruebas-testcontainers/usuarios	1.938s
```

## 8b · ¿y en una CI sin Docker? (contenedor golang sin socket de Docker)

`$ docker run --rm -v <repo>\pruebas-testcontainers:/app -v C:\Users\tu-usuario\go\pkg\mod:/go/pkg/mod -w /app -e GOFLAGS=-mod=mod golang:1.26 go test -count=1 -run ^TestPG$ -v ./usuarios` → código 1 · 10.1 s de reloj

```
=== RUN   TestPG
    usuarios_test.go:74: run postgres: generic container: get provider: rootless Docker not found, failed to create Docker provider
--- FAIL: TestPG (0.00s)
FAIL
FAIL	github.com/culturadevops/pruebas-testcontainers/usuarios	0.008s
FAIL
```
