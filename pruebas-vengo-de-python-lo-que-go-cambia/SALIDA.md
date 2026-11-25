# SALIDA — Vengo de Python: lo que Go cambia (ejecutado)

Generado por `python correr.py`. **No se edita a mano.**

- fecha: 2026-10-12 01:36 UTC
- imagen: golang:1.27.2 (Docker, linux/amd64)
- máquina: 24 · Intel(R) Core(TM) Ultra 7 270K Plus núcleos/CPU
- los tiempos son de ESTA máquina, dentro de Docker en Windows.

## Las versiones

`$ go version`

```
go version go1.27.2 linux/amd64
```

`$ python -VV`

```
Python 3.14.8 (main, Oct  6 2026, 01:59:23) [GCC 14.2.0]
```

## 1 · Trabajo de CPU repartido en 8 trozos

`$ python primos.py`

```
en fila      148933 primos · 5.44 s
8 hilos      148933 primos · 5.36 s
8 procesos   148933 primos · 1.01 s
```

`$ go run ./primos`

```
en fila        148933 primos · 0.35 s
8 goroutines   148933 primos · 0.06 s
```

## 2 · La errata en la rama que casi nunca se ejecuta

`$ python -m py_compile errata.py && echo 'compila: sin errores'; python errata.py`

```
compila: sin errores
todo bien
Traceback (most recent call last):
  File "/w/errata.py", line 10, in <module>
    print(avisar(True))
          ~~~~~~^^^^^^
  File "/w/errata.py", line 6, in avisar
    enviar_alrta("disco lleno")   # ← la errata
    ^^^^^^^^^^^^
NameError: name 'enviar_alrta' is not defined
```

`$ go build ./errata`

```
# ejemplo.com/python-a-go/errata
errata/main.go:10:3: undefined: enviarAlrta
```

`$ go vet ./errata`

```
# ejemplo.com/python-a-go/errata
# [ejemplo.com/python-a-go/errata]
vet: errata/main.go:10:3: undefined: enviarAlrta
```

## 2b · Un tipo equivocado

`$ python tipos.py`

```
arranca
Traceback (most recent call last):
  File "/w/tipos.py", line 7, in <module>
    print(siguiente_puerto(config["puerto"]))
          ~~~~~~~~~~~~~~~~^^^^^^^^^^^^^^^^^^
  File "/w/tipos.py", line 3, in siguiente_puerto
    return puerto + 1
           ~~~~~~~^~~
TypeError: can only concatenate str (not "int") to str
```

`$ go build ./tipos`

```
# ejemplo.com/python-a-go/tipos
tipos/main.go:11:30: cannot use config["puerto"] (map index expression of type string) as int value in argument to siguientePuerto
```

## 2c · Esperar a la red: 50 peticiones

`$ python esperar.py`

```
en fila    50 bien · 8.75 s
50 hilos   50 bien · 0.34 s
```

`$ go run ./esperar`

```
50 goroutines  50 bien · 0.32 s
```

## 2d · Lo que tarda en arrancar

`$ for i in $(seq 20); do python -c 'print(1)'; done   # media de 20`

```
python, 20 arranques: 8 ms cada uno
```

`$ for i in $(seq 20); do ./hola; done   # media de 20`

```
binario de Go, 20 arranques: 9 ms cada uno
```

## 3 · El mismo servicio, en una imagen

`$ docker images servicio-python servicio-go`

```
servicio-python  178MB
servicio-go  8.37MB
```

## 4 · Memoria en reposo

`$ docker stats --no-stream servicio-python servicio-go`

```
servicio-python  16.32MiB / 15.36GiB
servicio-go  3.125MiB / 15.36GiB
```
