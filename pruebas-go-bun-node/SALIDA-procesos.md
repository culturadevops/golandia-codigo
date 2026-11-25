# SALIDA — ¿Go, Bun o Node? Anexo: un hilo contra dos procesos, y el cómputo de una petición

> Corrida del 2026-10-19 17:26. Regenerable con `python investigacion/pruebas_go_bun_node_cluster.py`. Mismo método que SALIDA.md.

## A · Cómputo de UNA petición (mediana del reloj del propio servidor, sin concurrencia)

| variante | /trabajo (20.000 objetos) | /suma (200.000 vueltas) |
|---|---|---|
| go | 3.34 ms | 0.124 ms |
| bun | 3.85 ms | 1.625 ms |
| node | 3.34 ms | 1.157 ms |

## B · /trabajo con 2 procesos de JavaScript

`hey -c 16` desde otro contenedor; servidor con 2 CPU y 1g de memoria. «En frío»: los primeros 10 s tras arrancar. «Templado»: los 20 s siguientes.

| variante | peticiones/s en frío | peticiones/s templado | p50 | p99 | memoria máx. bajo carga |
|---|---|---|---|---|---|
| go | 467 | 464 | 27.5 ms | 127.1 ms | 34 MiB |
| bun | 257 | 259 | 57.6 ms | 120.2 ms | 30 MiB |
| bun · 2 procesos | 440 | 444 | 37.1 ms | 84.6 ms | 66 MiB |
| node | 258 | 259 | 57.8 ms | 117.8 ms | 121 MiB |
| node · 2 procesos | 381 | 416 | 33.5 ms | 86.4 ms | 285 MiB |

## C · /suma con 2 procesos de JavaScript

`hey -c 16` desde otro contenedor; servidor con 2 CPU y 1g de memoria. «En frío»: los primeros 10 s tras arrancar. «Templado»: los 20 s siguientes.

| variante | peticiones/s en frío | peticiones/s templado | p50 | p99 | memoria máx. bajo carga |
|---|---|---|---|---|---|
| go | 14637 | 14709 | 1.0 ms | 2.7 ms | 10 MiB |
| bun | 608 | 607 | 24.6 ms | 50.1 ms | 7 MiB |
| bun · 2 procesos | 1209 | 1216 | 13.0 ms | 26.4 ms | 16 MiB |
| node | 829 | 834 | 17.9 ms | 36.4 ms | 23 MiB |
| node · 2 procesos | 1639 | 1657 | 8.4 ms | 17.8 ms | 55 MiB |

