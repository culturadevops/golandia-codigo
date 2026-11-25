# SALIDA — testing/synctest en Go (ejecutado)

Generado por `python correr.py`. **No se edita a mano.**

- fecha: 2026-10-10 18:05 UTC
- imagen: golang:1.27.2 (Docker, linux/amd64)
- máquina: 24 · Intel(R) Core(TM) Ultra 7 270K Plus núcleos/CPU
- los tiempos son de ESTA máquina, dentro de Docker en Windows.

## La versión de Go

`$ go version`

```
go version go1.27.2 linux/amd64
```

## 1 · El test honesto: espera de verdad

`$ go test -count 1 -run CaducaEsperando -v`

```
--- PASS: TestCaducaEsperando (41.00s)
ok  	ejemplo.com/cache	41.005s
```

## 2 · El mismo test dentro de una burbuja

`$ go test -count 1 -run CaducaSynctest -v`

```
cache_test.go:27: hora dentro de la burbuja: 2000-01-01 00:00:00 +0000 UTC
cache_test.go:35: hora al terminar: 2000-01-01 00:00:41 +0000 UTC
--- PASS: TestCaducaSynctest (0.00s)
ok  	ejemplo.com/cache	0.003s
```

## 3 · Un deadlock dentro de la burbuja

`$ go test -count 1 -run Deadlock`

```
--- FAIL: TestDeadlock (0.00s)
panic: deadlock: all goroutines in bubble are blocked [recovered, repanicked]
```

## 4 · Lo que no avanza el reloj: esperar en la red

`$ go test -count 1 -run BloqueadoEnRed -timeout 8s`

```
panic: test timed out after 8s
		TestBloqueadoEnRed (8s)
(tardó 8 s y no terminó solo)
```

## 5 · Go 1.27: synctest.Sleep

`$ go test -count 1 -run CaducaSleep127 -v`

```
extra_test.go:23: con synctest.Sleep: quedan 0 entradas
--- PASS: TestCaducaSleep127 (0.00s)
ok  	ejemplo.com/cache	0.003s
```

## 6 · Un timeout de contexto de 5 s

`$ go test -count 1 -run ContextoTimeout -v`

```
extra_test.go:34: error: context deadline exceeded · pasaron 5s de reloj falso
--- PASS: TestContextoTimeout (0.00s)
ok  	ejemplo.com/cache	0.003s
```

## 7 · Un ticker de un segundo durante un minuto

`$ go test -count 1 -run TickerUnMinuto -v`

```
extra_test.go:50: ticks en un minuto: 59
--- PASS: TestTickerUnMinuto (0.00s)
ok  	ejemplo.com/cache	0.003s
```

## 8 · La solución a la red: net.Pipe

`$ go test -count 1 -run ConPipe -v`

```
extra_test.go:72: recibido "hola" tras 3s de reloj falso
--- PASS: TestConPipe (0.00s)
ok  	ejemplo.com/cache	0.003s
```

## 9 · Un mutex tampoco deja avanzar el reloj

`$ go test -count 1 -run BloqueadoEnMutex -timeout 8s`

```
panic: test timed out after 8s
		TestBloqueadoEnMutex (8s)
(tardó 8 s y no terminó solo)
```

## 10 · Una goroutine que queda colgada al terminar

`$ go test -count 1 -run FugaAlTerminar`

```
--- FAIL: TestFugaAlTerminar (0.00s)
panic: deadlock: main bubble goroutine has exited but blocked goroutines remain [recovered, repanicked]
panic({0x7441c8?, 0x1cd151d30180?})
```

## 12 · Dos relojes: fuera y dentro

`$ go test -count 1 -run DosRelojes -v`

```
extra_test.go:96: fuera de la burbuja: año 2026
extra_test.go:98: dentro de la burbuja: 2000-01-01 00:00:00 +0000 UTC
--- PASS: TestDosRelojes (0.00s)
ok  	ejemplo.com/cache	0.005s
```

## 13 · Un canal creado fuera de la burbuja

`$ go test -count 1 -run CanalDeFuera -timeout 8s`

```
panic: test timed out after 8s
		TestCanalDeFuera (8s)
(tardó 8 s)
```

## 14 · synctest con el detector de carreras

`$ go test -count 1 -race -run 'CaducaSynctest|CaducaSleep127|TickerUnMinuto' -v`

```
--- PASS: TestCaducaSynctest (0.00s)
--- PASS: TestCaducaSleep127 (0.00s)
--- PASS: TestTickerUnMinuto (0.00s)
ok  	ejemplo.com/cache	1.015s
```

## 11 · Toda la batería, sin los que se cuelgan

`$ go test -count 1 -skip 'BloqueadoEn|Deadlock|FugaAlTerminar|CaducaEsperando|CanalDeFuera|DosRelojes' -v`

```
--- PASS: TestCaducaSynctest (0.00s)
--- PASS: TestCaducaSleep127 (0.00s)
--- PASS: TestContextoTimeout (0.00s)
--- PASS: TestTickerUnMinuto (0.00s)
--- PASS: TestConPipe (0.00s)
ok  	ejemplo.com/cache	0.004s
```
