# SALIDA — ¿Go o Ruby? (golandia 804), con el banco de la revancha

> Corrida del 2026-11-05 09:04, con el candado `_logs/banco.lock` tomado. Regenerable con `python investigacion/pruebas_go_vs_lote_2026_11.py ruby` (mismo banco que 651-654 y 699). ⛔ Ninguna cifra de blog: todo salió de aquí.

- go: `go version go1.27.2 linux/amd64` (net/http, solo librería estándar)
- ruby: `ruby 4.0.9 (2026-10-23 revision 3e93b23afc) +PRISM [x86_64-linux]`
- Docker `29.6.2`. Servidores con 2 CPU y 1g; `hey -c 16` con 4 CPU desde otro contenedor.

## 1 · Construir la imagen desde cero (`docker build --no-cache`, dependencias incluidas)

| lenguaje | qué se cronometra | segundos |
|---|---|---|
| go | go build | 5.7 |
| ruby | bundle install de puma + rack | 27.5 |

## 2 · Lo que despliegas

| imagen | base | tamaño |
|---|---|---|
| go | scratch + binario estático | 2.4 MiB |
| ruby | ruby:4.0.9-slim + Puma 7 + Rack 3 (sin Rails) | 187.2 MiB |

El artefacto que se copia a la imagen: Go, 5.6 MiB.

## 0 · ¿Hacen lo mismo? (antes de medir)

- go: total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go
- ruby (1 proceso): total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go
- ruby + YJIT (1 proceso): total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go
- ruby + YJIT (2 procesos): total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go
- ruby + ZJIT (1 proceso): total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go

## 3 · Arranque: contenedor nuevo → primera respuesta 200 (mediana de 8) y memoria en reposo

| variante | 1ª respuesta (docker incluido) | «listo» (su propio reloj) | memoria en reposo |
|---|---|---|---|
| go | 214 ms | 0.0 ms | 2.8 MiB |
| ruby (1 proceso) | 384 ms | — | 25.4 MiB |
| ruby + YJIT (1 proceso) | 476 ms | — | 28.0 MiB |
| ruby + YJIT (2 procesos) | 485 ms | — | 41.1 MiB |
| ruby + ZJIT (1 proceso) | 449 ms | — | 28.2 MiB |

## 4 · Carga con memoria: /trabajo (objetos, mapa, orden, JSON), a tope

`hey -c 16` desde otro contenedor; servidor con 2 CPU y 1g. «En frío»: los primeros 10 s tras arrancar. «Templado»: los 20 s siguientes.

| variante | peticiones/s en frío | peticiones/s templado | p50 | p99 | memoria máx. bajo carga |
|---|---|---|---|---|---|
| go | 424 | 445 | 29.1 ms | 121.6 ms | 35 MiB |
| ruby (1 proceso) | 36 | 36 | 460.7 ms | 908.0 ms | 63 MiB |
| ruby + YJIT (1 proceso) | 36 | 36 | 449.8 ms | 912.2 ms | 75 MiB |
| ruby + YJIT (2 procesos) | 73 | 73 | 238.3 ms | 481.9 ms | 122 MiB |
| ruby + ZJIT (1 proceso) | 36 | 36 | 446.4 ms | 893.1 ms | 85 MiB |

## 5 · Carga de CPU pura: /suma, a tope

`hey -c 16` desde otro contenedor; servidor con 2 CPU y 1g. «En frío»: los primeros 10 s tras arrancar. «Templado»: los 20 s siguientes.

| variante | peticiones/s en frío | peticiones/s templado | p50 | p99 | memoria máx. bajo carga |
|---|---|---|---|---|---|
| go | 14458 | 14404 | 1.0 ms | 2.8 ms | 11 MiB |
| ruby (1 proceso) | 361 | 360 | 44.4 ms | 90.1 ms | 71 MiB |
| ruby + YJIT (1 proceso) | 1765 | 1777 | 9.0 ms | 18.7 ms | 88 MiB |
| ruby + YJIT (2 procesos) | 3378 | 3482 | 4.4 ms | 11.3 ms | 176 MiB |
| ruby + ZJIT (1 proceso) | 1689 | 1708 | 9.3 ms | 19.4 ms | 90 MiB |

## 6 · Latencia de una petición sola: /trabajo

Un solo cliente a 5 peticiones/s (`hey -c 1 -q 5`), 20 s, tras 10 s de calentamiento a tope. Cada petición llega sola: no hay cola, es lo que tarda el servidor en atender una.

| variante | peticiones/s logradas | p50 | p99 |
|---|---|---|---|
| go | 5 | 4.10 ms | 6.50 ms |
| ruby (1 proceso) | 5 | 26.50 ms | 30.70 ms |
| ruby + YJIT (1 proceso) | 5 | 25.50 ms | 27.70 ms |
| ruby + YJIT (2 procesos) | 5 | 26.10 ms | 30.50 ms |
| ruby + ZJIT (1 proceso) | 5 | 26.70 ms | 39.20 ms |

## 7 · Latencia de una petición sola: /suma

Un solo cliente a 5 peticiones/s (`hey -c 1 -q 5`), 20 s, tras 10 s de calentamiento a tope. Cada petición llega sola: no hay cola, es lo que tarda el servidor en atender una.

| variante | peticiones/s logradas | p50 | p99 |
|---|---|---|---|
| go | 5 | 0.50 ms | 1.30 ms |
| ruby (1 proceso) | 5 | 3.20 ms | 4.80 ms |
| ruby + YJIT (1 proceso) | 5 | 1.00 ms | 2.10 ms |
| ruby + YJIT (2 procesos) | 5 | 1.00 ms | 1.60 ms |
| ruby + ZJIT (1 proceso) | 5 | 1.10 ms | 2.80 ms |

## 8 · Cómputo puro: el reloj del propio proceso, sin red

50 peticiones de calentamiento y 300 medidas, de una en una; el tiempo es el campo «ns» que mide cada servidor alrededor de su propio cálculo.

| variante | /suma?n=200000 mediana | p99 | /trabajo?n=20000 mediana | p99 |
|---|---|---|---|---|
| go | 0.122 ms | 0.179 ms | 3.377 ms | 3.903 ms |
| ruby (1 proceso) | 2.700 ms | 4.155 ms | 27.127 ms | 29.811 ms |
| ruby + YJIT (1 proceso) | 0.457 ms | 0.623 ms | 26.272 ms | 30.720 ms |
| ruby + YJIT (2 procesos) | 0.457 ms | 2.645 ms | 25.957 ms | 30.969 ms |
| ruby + ZJIT (1 proceso) | 0.466 ms | 0.744 ms | 24.967 ms | 27.108 ms |

