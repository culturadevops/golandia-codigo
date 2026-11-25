# SALIDA — Go práctico 2027 · parte 3

> Generado por `ejecutar_todo.py`. **No se escribe a mano.** Ids, tiempos y tamaños son los de esta corrida.

- fecha: `2026-11-23 12:03:30Z`
- imagen Go: `golang:1.27.2` · Postgres: `postgres:18.6-alpine` · `Docker 29.6.2`

## tabla

```
$ go test -run TestValidarTitulo -v .
=== RUN   TestValidarTitulo
=== RUN   TestValidarTitulo/normal
=== RUN   TestValidarTitulo/quita_espacios
=== RUN   TestValidarTitulo/vacío
=== RUN   TestValidarTitulo/solo_espacios
=== RUN   TestValidarTitulo/120_letras
=== RUN   TestValidarTitulo/121_letras
=== RUN   TestValidarTitulo/120_eñes
--- PASS: TestValidarTitulo (0.00s)
    --- PASS: TestValidarTitulo/normal (0.00s)
    --- PASS: TestValidarTitulo/quita_espacios (0.00s)
    --- PASS: TestValidarTitulo/vacío (0.00s)
    --- PASS: TestValidarTitulo/solo_espacios (0.00s)
    --- PASS: TestValidarTitulo/120_letras (0.00s)
    --- PASS: TestValidarTitulo/121_letras (0.00s)
    --- PASS: TestValidarTitulo/120_eñes (0.00s)
PASS
ok  	tareas	(cached)
```

## rota

```
$ grep -n "maxTitulo:" tareas.go
29:	case n >= maxTitulo:
$ go test -run TestValidarTitulo .
--- FAIL: TestValidarTitulo (0.00s)
    --- FAIL: TestValidarTitulo/120_letras (0.00s)
        tareas_test.go:28: validarTitulo("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"): error = el título tiene 120 caracteres; el máximo es 120, quería error: false
    --- FAIL: TestValidarTitulo/120_eñes (0.00s)
        tareas_test.go:28: validarTitulo("ññññññññññññññññññññññññññññññññññññññññññññññññññññññññññññññññññññññññññññññññññññññññññññññññññññññññññññññññññññññññ"): error = el título tiene 120 caracteres; el máximo es 120, quería error: false
FAIL
FAIL	tareas	0.002s
FAIL
[código de salida: 1]
```

## httptest

```
$ go test -run "TestCrear|TestRutas|TestFlujo" -v .
=== RUN   TestCrearTarea
=== RUN   TestCrearTarea/válida
=== RUN   TestCrearTarea/título_vacío
=== RUN   TestCrearTarea/JSON_roto
=== RUN   TestCrearTarea/campo_desconocido
=== RUN   TestCrearTarea/título_largo
--- PASS: TestCrearTarea (0.00s)
    --- PASS: TestCrearTarea/válida (0.00s)
    --- PASS: TestCrearTarea/título_vacío (0.00s)
    --- PASS: TestCrearTarea/JSON_roto (0.00s)
    --- PASS: TestCrearTarea/campo_desconocido (0.00s)
    --- PASS: TestCrearTarea/título_largo (0.00s)
=== RUN   TestCrearDevuelveLocation
--- PASS: TestCrearDevuelveLocation (0.00s)
=== RUN   TestRutas
=== RUN   TestRutas/GET_/tareas/99
=== RUN   TestRutas/GET_/tareas/abc
=== RUN   TestRutas/PUT_/tareas
=== RUN   TestRutas/DELETE_/tareas/1
--- PASS: TestRutas (0.00s)
    --- PASS: TestRutas/GET_/tareas/99 (0.00s)
    --- PASS: TestRutas/GET_/tareas/abc (0.00s)
    --- PASS: TestRutas/PUT_/tareas (0.00s)
    --- PASS: TestRutas/DELETE_/tareas/1 (0.00s)
=== RUN   TestFlujoCompleto
--- PASS: TestFlujoCompleto (0.00s)
PASS
ok  	tareas	(cached)
```

## sin-bd

```
$ go test ./...
ok  	tareas	(cached)
?   	tareas/herramientas/bloquear	[no test files]
$ go test -run Postgres -v .
=== RUN   TestPostgresAPI
    postgres_test.go:39: sin TAREAS_TEST_BD: me salto la prueba con Postgres
--- SKIP: TestPostgresAPI (0.00s)
=== RUN   TestPostgresRechazaTituloLargo
    postgres_test.go:47: sin TAREAS_TEST_BD: me salto la prueba con Postgres
--- SKIP: TestPostgresRechazaTituloLargo (0.00s)
=== RUN   TestPostgresNoEncontrada
    postgres_test.go:57: sin TAREAS_TEST_BD: me salto la prueba con Postgres
--- SKIP: TestPostgresNoEncontrada (0.00s)
PASS
ok  	tareas	0.002s
```

## con-bd

```
$ TAREAS_TEST_BD="$BD_PRUEBAS" go test -run Postgres -v .
=== RUN   TestPostgresAPI
--- PASS: TestPostgresAPI (0.02s)
=== RUN   TestPostgresRechazaTituloLargo
    postgres_test.go:53: Postgres dijo: new row for relation "tareas" violates check constraint "tareas_titulo_check" (23514)
--- PASS: TestPostgresRechazaTituloLargo (0.01s)
=== RUN   TestPostgresNoEncontrada
--- PASS: TestPostgresNoEncontrada (0.01s)
PASS
ok  	tareas	0.042s
```

## cobertura

```
$ go test -cover .
ok  	tareas	0.008s	coverage: 42.5% of statements
$ TAREAS_TEST_BD="$BD_PRUEBAS" go test -cover -coverprofile=cover.out .
ok  	tareas	0.034s	coverage: 63.5% of statements
$ go tool cover -func=cover.out
tareas/main.go:11:		main			0.0%
tareas/manejadores.go:37:	Rutas			100.0%
tareas/manejadores.go:48:	salud			0.0%
tareas/manejadores.go:57:	listar			0.0%
tareas/manejadores.go:66:	crear			85.7%
tareas/manejadores.go:88:	obtener			100.0%
tareas/manejadores.go:102:	marcar			50.0%
tareas/manejadores.go:127:	borrar			75.0%
tareas/manejadores.go:141:	leerID			100.0%
tareas/manejadores.go:150:	leerJSON		100.0%
tareas/manejadores.go:160:	responderJSON		100.0%
tareas/manejadores.go:166:	responderError		100.0%
tareas/manejadores.go:171:	errorAlmacen		25.0%
tareas/memoria.go:17:		NuevaMemoria		100.0%
tareas/memoria.go:21:		Listar			100.0%
tareas/memoria.go:32:		Crear			100.0%
tareas/memoria.go:41:		Obtener			100.0%
tareas/memoria.go:51:		Marcar			87.5%
tareas/memoria.go:63:		Borrar			100.0%
tareas/memoria.go:73:		Salud			0.0%
tareas/migraciones.go:18:	Migrar			41.9%
tareas/postgres.go:20:		ConectarPostgres	69.2%
tareas/postgres.go:46:		Listar			0.0%
tareas/postgres.go:57:		Crear			100.0%
tareas/postgres.go:67:		Obtener			100.0%
tareas/postgres.go:77:		Marcar			100.0%
tareas/postgres.go:87:		Borrar			87.5%
tareas/postgres.go:101:		Salud			0.0%
tareas/postgres.go:115:		clasificar		71.4%
tareas/tareas.go:23:		validarTitulo		100.0%
total:				(statements)		63.5%
```

## carrera

```
$ grep -c "mu.Lock" memoria.go
0
[código de salida: 1]
$ go test -run TestMemoriaConcurrente -count=1 . 2>&1 | head -n 8
fatal error: concurrent map writes

goroutine 9 [running]:
internal/runtime/maps.fatal({0x947917?, 0x4f5bba?})
	/usr/local/go/src/runtime/panic.go:1195 +0x18
tareas.(*Memoria).Crear(...)
	/work/tareas/memoria.go:33
tareas.TestMemoriaConcurrente.func1()
$ go test -race -run TestMemoriaConcurrente -count=1 . 2>&1 | head -n 30
==================
WARNING: DATA RACE
Read at 0x00c000016218 by goroutine 12:
  tareas.(*Memoria).Crear()
      /work/tareas/memoria.go:31 +0x124
  tareas.TestMemoriaConcurrente.func1()
      /work/tareas/memoria_test.go:18 +0x58

Previous write at 0x00c000016218 by goroutine 10:
  tareas.(*Memoria).Crear()
      /work/tareas/memoria.go:31 +0x144
  tareas.TestMemoriaConcurrente.func1()
      /work/tareas/memoria_test.go:18 +0x58

Goroutine 12 (running) created at:
  sync.(*WaitGroup).Go()
      /usr/local/go/src/sync/waitgroup.go:238 +0x72
  testing.tRunner()
      /usr/local/go/src/testing/testing.go:2195 +0x21c
  testing.(*T).Run.gowrap1()
      /usr/local/go/src/testing/testing.go:2260 +0x38

Goroutine 10 (finished) created at:
  sync.(*WaitGroup).Go()
      /usr/local/go/src/sync/waitgroup.go:238 +0x72
  testing.tRunner()
      /usr/local/go/src/testing/testing.go:2195 +0x21c
  testing.(*T).Run.gowrap1()
      /usr/local/go/src/testing/testing.go:2260 +0x38
==================
$ go test -race -run TestMemoriaConcurrente -count=1 . 2>&1 | grep -E "^(FAIL|ok|--- )|race detected|DATA RACE" | sort | uniq -c
      1 FAIL
      1 FAIL	tareas	0.020s
      5 WARNING: DATA RACE
$ grep -c "mu.Lock" memoria.go
5
$ TAREAS_TEST_BD="$BD_PRUEBAS" go test -race -count=1 ./...
ok  	tareas	1.068s
?   	tareas/herramientas/bloquear	[no test files]

[demo.sh terminó con código 0]
```
