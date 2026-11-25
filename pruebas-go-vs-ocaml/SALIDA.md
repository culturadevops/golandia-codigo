# SALIDA — ¿Go o OCaml 5? (golandia 903), corrida 1, con el banco de la serie

> Corrida del 2026-11-23 11:31, con el candado `_logs/banco.lock` tomado. Regenerable con `python investigacion/pruebas_go_vs_x_2026_11_23.py ocaml 1` (mismo banco que 879 y 865). ⛔ Ninguna cifra de blog: todo salió de aquí.

- go: `go version go1.27.2 linux/amd64` (net/http, solo librería estándar)
- ocaml: `5.5.2 · cohttp-eio 6.3.0 · eio        1.7`
- Docker `29.6.2`. Servidores con 2 CPU y 1g; `hey -c 16` con 4 CPU desde otro contenedor.

## 1 · Construir la imagen desde cero (`docker build --no-cache`, dependencias incluidas)

| lenguaje | qué se cronometra | segundos |
|---|---|---|
| go | go build | 5.8 |
| ocaml | opam install dune eio_main cohttp-eio mtime + dune build --profile release | 48.7 |

## 2 · Lo que despliegas

| imagen | base | tamaño |
|---|---|---|
| go | scratch + binario estático | 2.4 MiB |
| ocaml | debian:trixie-slim + binario de ocamlopt (OCaml 5.5.2, cohttp-eio 6.3.0) | 31.6 MiB |

El artefacto que se copia a la imagen: Go, 5.6 MiB; ocaml, 8.98 MiB.

## Estado de la máquina antes de medir

- /proc/loadavg (desde un contenedor alpine, VM de Docker con 24 hilos): `2.10 0.79 0.29 2/533 6`
- contenedores ajenos en marcha: ninguno

## 0 · ¿Hacen lo mismo? (antes de medir)

- go: total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go
- ocaml (1 dominio, por defecto): total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go
- ocaml, 2 dominios: total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go
- ocaml, 24 dominios: total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go
- ocaml, 2 dominios + io_uring: total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go

## 3 · Arranque: contenedor nuevo → primera respuesta 200 (mediana de 8) y memoria en reposo

| variante | 1ª respuesta (docker incluido) | «listo» (su propio reloj) | memoria en reposo |
|---|---|---|---|
| go | 214 ms | 0.0 ms | 2.7 MiB |
| ocaml (1 dominio, por defecto) | 211 ms | 1.7 ms | 7.6 MiB |
| ocaml, 2 dominios | 187 ms | 1.7 ms | 12.0 MiB |
| ocaml, 24 dominios | 212 ms | 1.7 ms | 106.0 MiB |
| ocaml, 2 dominios + io_uring | 202 ms | 0.1 ms | 3.9 MiB |

## 4 · Carga con memoria: /trabajo (objetos, mapa, orden, JSON), a tope

`hey -c 16` desde otro contenedor; servidor con 2 CPU y 1g. «En frío»: los primeros 10 s tras arrancar. «Templado»: los 20 s siguientes.

| variante | peticiones/s en frío | peticiones/s templado | p50 | p99 | memoria máx. bajo carga |
|---|---|---|---|---|---|
| go | 481 | 481 | 25.8 ms | 123.4 ms | 32 MiB |
| ocaml (1 dominio, por defecto) | 162 | 161 | 98.8 ms | 103.1 ms | 16 MiB |
| ocaml, 2 dominios | 315 | 315 | 53.7 ms | 59.9 ms | 40 MiB |
| ocaml, 24 dominios | 150 | 136 | 96.5 ms | 487.9 ms | 261 MiB |
| ocaml, 2 dominios + io_uring | 318 | 318 | 49.9 ms | 81.9 ms | 33 MiB |

## 5 · Carga de CPU pura: /suma, a tope

`hey -c 16` desde otro contenedor; servidor con 2 CPU y 1g. «En frío»: los primeros 10 s tras arrancar. «Templado»: los 20 s siguientes.

| variante | peticiones/s en frío | peticiones/s templado | p50 | p99 | memoria máx. bajo carga |
|---|---|---|---|---|---|
| go | 14905 | 15034 | 1.0 ms | 2.6 ms | 11 MiB |
| ocaml (1 dominio, por defecto) | 5108 | 5112 | 3.1 ms | 3.5 ms | 12 MiB |
| ocaml, 2 dominios | 10053 | 9883 | 1.7 ms | 2.1 ms | 20 MiB |
| ocaml, 24 dominios | 7051 | 9309 | 1.0 ms | 46.4 ms | 132 MiB |
| ocaml, 2 dominios + io_uring | 10371 | 10229 | 1.5 ms | 2.9 ms | 13 MiB |

## 6 · Latencia de una petición sola: /trabajo

Un solo cliente a 5 peticiones/s (`hey -c 1 -q 5`), 20 s, tras 10 s de calentamiento a tope. Cada petición llega sola: no hay cola, es lo que tarda el servidor en atender una.

| variante | peticiones/s logradas | p50 | p99 |
|---|---|---|---|
| go | 5 | 5.80 ms | 10.60 ms |
| ocaml (1 dominio, por defecto) | 5 | 9.10 ms | 12.10 ms |
| ocaml, 2 dominios | 5 | 9.40 ms | 11.90 ms |
| ocaml, 24 dominios | 5 | 11.10 ms | 20.10 ms |
| ocaml, 2 dominios + io_uring | 5 | 8.20 ms | 13.20 ms |

## 7 · Latencia de una petición sola: /suma

Un solo cliente a 5 peticiones/s (`hey -c 1 -q 5`), 20 s, tras 10 s de calentamiento a tope. Cada petición llega sola: no hay cola, es lo que tarda el servidor en atender una.

| variante | peticiones/s logradas | p50 | p99 |
|---|---|---|---|
| go | 5 | 0.80 ms | 2.70 ms |
| ocaml (1 dominio, por defecto) | 5 | 0.80 ms | 3.30 ms |
| ocaml, 2 dominios | 5 | 0.70 ms | 4.00 ms |
| ocaml, 24 dominios | 5 | 0.70 ms | 2.40 ms |
| ocaml, 2 dominios + io_uring | 5 | 0.60 ms | 2.90 ms |

## 8 · Cómputo puro: el reloj del propio proceso, sin red

50 peticiones de calentamiento y 300 medidas, de una en una; el tiempo es el campo «ns» que mide cada servidor alrededor de su propio cálculo.

| variante | /suma?n=200000 mediana | p99 | /trabajo?n=20000 mediana | p99 |
|---|---|---|---|---|
| go | 0.122 ms | 0.147 ms | 3.328 ms | 3.826 ms |
| ocaml (1 dominio, por defecto) | 0.180 ms | 0.327 ms | 6.201 ms | 6.971 ms |
| ocaml, 2 dominios | 0.180 ms | 0.234 ms | 5.775 ms | 6.658 ms |
| ocaml, 24 dominios | 0.180 ms | 0.461 ms | 5.906 ms | 10.654 ms |
| ocaml, 2 dominios + io_uring | 0.180 ms | 0.270 ms | 5.223 ms | 6.619 ms |

Al terminar: /proc/loadavg `0.50 1.08 0.88 5/520 6`; ajenos: ninguno

Vigilancia de renders (remotion / fab.py producir) cada 10 s durante toda la corrida: ninguno vivo (orquestar_902_905.py, 2026-11-23 11:45).
