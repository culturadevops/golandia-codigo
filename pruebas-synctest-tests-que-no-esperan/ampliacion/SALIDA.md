# SALIDA — synctest — ampliación (ejecutado)

Generado por `python correr.py`. **No se edita a mano.**

- fecha: 2026-10-11 02:54 UTC
- imagen: golang:1.27.2 (Docker, linux/amd64)
- máquina: 24 · Intel(R) Core(TM) Ultra 7 270K Plus núcleos/CPU
- los tiempos son de ESTA máquina, dentro de Docker en Windows.

`$ go get golang.org/x/time`

```
require golang.org/x/time v0.14.0
```

## 1 · Reintentos con espera exponencial, sus tres caminos

`$ go test -v -run 'Acierta|Rinde|Cancelado'`

```
reintentos_test.go:24: intentos 2 · err <nil> · reloj falso 1s
--- PASS: TestAciertaAlSegundo (0.00s)
    reintentos_test.go:35: intentos 5 · err me rindo: 5 intentos · reloj falso 15s
--- PASS: TestSeRinde (0.00s)
    reintentos_test.go:48: intentos 3 · err context deadline exceeded · reloj falso 5s
--- PASS: TestCancelado (0.00s)
ok  	ejemplo.com/reintentos	0.002s
```

## 2 · Una librería que no sabe nada: golang.org/x/time/rate

`$ go test -v -run Limitador`

```
reintentos_test.go:65: 10 peticiones a 2 por segundo · reloj falso 4.5s
ok  	ejemplo.com/reintentos	0.002s
```

## 3 · Una tarea cada hora, durante un año

`$ go test -v -run UnAno`

```
reintentos_test.go:79: 8760 ticks · ahora es 2000-12-31 00:00
--- PASS: TestUnAno (0.00s)
ok  	ejemplo.com/reintentos	0.005s
```

## 4 · Un debounce con time.AfterFunc

`$ go test -v -run Debounce`

```
debounce_test.go:20: tras teclear: 0 búsquedas
    debounce_test.go:22: al final: 1 búsqueda, a los 700ms
--- PASS: TestDebounce (0.00s)
ok  	ejemplo.com/reintentos	0.002s
```

## 5 · Toda la ampliación, mil veces

`$ go test -count=1000 .   # 6 tests × 1.000`

```
ok  	ejemplo.com/reintentos	3.791s
```
