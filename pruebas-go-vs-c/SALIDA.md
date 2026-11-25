# SALIDA — ¿Go o C? (golandia 803), con el banco de la revancha

> Corrida del 2026-11-05 08:52, con el candado `_logs/banco.lock` tomado. Regenerable con `python investigacion/pruebas_go_vs_lote_2026_11.py c` (mismo banco que 651-654 y 699). ⛔ Ninguna cifra de blog: todo salió de aquí.

- go: `go version go1.27.2 linux/amd64` (net/http, solo librería estándar)
- c: `gcc (GCC) 16.2.0 · Copyright (C) 2026 Free Software Foundation, Inc. · This is free software; see the source for copying conditions.  There is NO`
- Docker `29.6.2`. Servidores con 2 CPU y 1g; `hey -c 16` con 4 CPU desde otro contenedor.

## 1 · Construir la imagen desde cero (`docker build --no-cache`, dependencias incluidas)

| lenguaje | qué se cronometra | segundos |
|---|---|---|
| go | go build | 5.6 |
| c | gcc -O2 + apt install libmicrohttpd-dev | 4.2 |

## 2 · Lo que despliegas

| imagen | base | tamaño |
|---|---|---|
| go | scratch + binario estático | 2.4 MiB |
| c | debian:trixie-slim + libmicrohttpd12 + binario (gcc 16.2 -O2) | 31.0 MiB |

El artefacto que se copia a la imagen: Go, 5.6 MiB; c, 0.01 MiB.

## 0 · ¿Hacen lo mismo? (antes de medir)

- go: total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go
- c (2 hilos): total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go
- c (16 hilos): total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go

## 3 · Arranque: contenedor nuevo → primera respuesta 200 (mediana de 8) y memoria en reposo

| variante | 1ª respuesta (docker incluido) | «listo» (su propio reloj) | memoria en reposo |
|---|---|---|---|
| go | 220 ms | 0.0 ms | 2.8 MiB |
| c (2 hilos) | 227 ms | 0.1 ms | 2.1 MiB |
| c (16 hilos) | 223 ms | 0.6 ms | 5.4 MiB |

## 4 · Carga con memoria: /trabajo (objetos, mapa, orden, JSON), a tope

`hey -c 16` desde otro contenedor; servidor con 2 CPU y 1g. «En frío»: los primeros 10 s tras arrancar. «Templado»: los 20 s siguientes.

| variante | peticiones/s en frío | peticiones/s templado | p50 | p99 | memoria máx. bajo carga |
|---|---|---|---|---|---|
| go | 449 | 448 | 28.5 ms | 129.3 ms | 32 MiB |
| c (2 hilos) | 645 | 638 | 25.0 ms | 44.3 ms | 11 MiB |
| c (16 hilos) | 524 | 518 | 6.4 ms | 99.0 ms | 42 MiB |

## 5 · Carga de CPU pura: /suma, a tope

`hey -c 16` desde otro contenedor; servidor con 2 CPU y 1g. «En frío»: los primeros 10 s tras arrancar. «Templado»: los 20 s siguientes.

| variante | peticiones/s en frío | peticiones/s templado | p50 | p99 | memoria máx. bajo carga |
|---|---|---|---|---|---|
| go | 14456 | 14475 | 1.0 ms | 2.8 ms | 10 MiB |
| c (2 hilos) | 15001 | 15185 | 1.0 ms | 1.9 ms | 6 MiB |
| c (16 hilos) | 7380 | 7553 | 0.7 ms | 65.4 ms | 11 MiB |

## 6 · Latencia de una petición sola: /trabajo

Un solo cliente a 5 peticiones/s (`hey -c 1 -q 5`), 20 s, tras 10 s de calentamiento a tope. Cada petición llega sola: no hay cola, es lo que tarda el servidor en atender una.

| variante | peticiones/s logradas | p50 | p99 |
|---|---|---|---|
| go | 5 | 4.10 ms | 7.10 ms |
| c (2 hilos) | 5 | 4.10 ms | 5.40 ms |
| c (16 hilos) | 5 | 4.30 ms | 5.70 ms |

## 7 · Latencia de una petición sola: /suma

Un solo cliente a 5 peticiones/s (`hey -c 1 -q 5`), 20 s, tras 10 s de calentamiento a tope. Cada petición llega sola: no hay cola, es lo que tarda el servidor en atender una.

| variante | peticiones/s logradas | p50 | p99 |
|---|---|---|---|
| go | 5 | 0.50 ms | 2.80 ms |
| c (2 hilos) | 5 | 0.90 ms | 2.90 ms |
| c (16 hilos) | 5 | 1.10 ms | 3.50 ms |

## 8 · Cómputo puro: el reloj del propio proceso, sin red

50 peticiones de calentamiento y 300 medidas, de una en una; el tiempo es el campo «ns» que mide cada servidor alrededor de su propio cálculo.

| variante | /suma?n=200000 mediana | p99 | /trabajo?n=20000 mediana | p99 |
|---|---|---|---|---|
| go | 0.122 ms | 0.186 ms | 3.481 ms | 4.263 ms |
| c (2 hilos) | 0.109 ms | 0.163 ms | 2.579 ms | 3.607 ms |
| c (16 hilos) | 0.109 ms | 0.322 ms | 2.569 ms | 4.371 ms |


## 9 · El binario de C y lo que necesita para arrancar (2026-11-05, a mano tras la corrida)

```
$ docker run --rm --entrypoint stat v803-c:1-etapa -c %s /servidor
14456
$ docker run --rm --entrypoint ldd v803-c:1 /servidor
	linux-vdso.so.1
	libmicrohttpd.so.12 => /lib/x86_64-linux-gnu/libmicrohttpd.so.12
	libc.so.6 => /lib/x86_64-linux-gnu/libc.so.6
	libgnutls.so.30 => /lib/x86_64-linux-gnu/libgnutls.so.30
	/lib64/ld-linux-x86-64.so.2
	libp11-kit.so.0 => /lib/x86_64-linux-gnu/libp11-kit.so.0
	libidn2.so.0 => /lib/x86_64-linux-gnu/libidn2.so.0
	libunistring.so.5 => /lib/x86_64-linux-gnu/libunistring.so.5
	libtasn1.so.6 => /lib/x86_64-linux-gnu/libtasn1.so.6
	libhogweed.so.6 => /lib/x86_64-linux-gnu/libhogweed.so.6
	libnettle.so.8 => /lib/x86_64-linux-gnu/libnettle.so.8
	libgmp.so.10 => /lib/x86_64-linux-gnu/libgmp.so.10
	libffi.so.8 => /lib/x86_64-linux-gnu/libffi.so.8
```
