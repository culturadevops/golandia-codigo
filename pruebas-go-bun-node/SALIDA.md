# SALIDA — ¿Go, Bun o Node? (golandia), con el banco de la revancha

> Corrida del 2026-10-19 17:21. Regenerable con `python investigacion/pruebas_go_bun_node.py`. ⛔ Ninguna cifra de blog.

- go: `go version go1.26.9 linux/amd64`
- bun: `1.4.3`
- node: `v24.22.0`
- Servidores con 2 CPU y 1g; `hey -c 16` con 4 CPU desde otro contenedor. Cada runtime con su API nativa.

## 1 · Lo que despliegas

| imagen | base | descarga (comprimida) |
|---|---|---|
| go | scratch | 2.4 MiB |
| bun | oven/bun:1.4.3-slim | 70.6 MiB |
| node | node:24-slim | 77.3 MiB |

## 0 · ¿Hacen lo mismo? (antes de medir)

- go: total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go
- bun: total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go
- node: total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go

## 3 · Arranque: contenedor nuevo → primera respuesta 200 (mediana de 8) y memoria en reposo

| variante | 1ª respuesta (docker incluido) | «listo» (su propio reloj) | memoria en reposo |
|---|---|---|---|
| go | 209 ms | 0.0 ms | 3.0 MiB |
| bun | 209 ms | 0.2 ms | 3.8 MiB |
| node | 211 ms | 2.9 ms | 12.5 MiB |

## 4 · Carga con memoria: /trabajo (objetos, mapa, orden, JSON)

`hey -c 16` desde otro contenedor; servidor con 2 CPU y 1g de memoria. «En frío»: los primeros 10 s tras arrancar. «Templado»: los 20 s siguientes.

| variante | peticiones/s en frío | peticiones/s templado | p50 | p99 | memoria máx. bajo carga |
|---|---|---|---|---|---|
| go | 457 | 450 | 27.6 ms | 131.6 ms | 36 MiB |
| bun | 244 | 246 | 60.3 ms | 126.5 ms | 34 MiB |
| node | 253 | 255 | 58.5 ms | 120.0 ms | 128 MiB |

## 5 · Carga de CPU pura: /suma

`hey -c 16` desde otro contenedor; servidor con 2 CPU y 1g de memoria. «En frío»: los primeros 10 s tras arrancar. «Templado»: los 20 s siguientes.

| variante | peticiones/s en frío | peticiones/s templado | p50 | p99 | memoria máx. bajo carga |
|---|---|---|---|---|---|
| go | 14595 | 14577 | 1.0 ms | 2.7 ms | 11 MiB |
| bun | 603 | 607 | 24.6 ms | 50.1 ms | 8 MiB |
| node | 829 | 834 | 17.9 ms | 36.4 ms | 26 MiB |

