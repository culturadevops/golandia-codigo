# SALIDA — ¿Go o TypeScript compilado a nativo? (golandia 880), corrida 1, con el banco de la serie

> Corrida del 2026-11-18 10:20, con el candado `_logs/banco.lock` tomado. Regenerable con `python investigacion/pruebas_go_vs_nativo_2026_11.py ts-nativo 1` (mismo banco que 865). ⛔ Ninguna cifra de blog: todo salió de aquí.

- go: `go version go1.27.2 linux/amd64` (net/http, solo librería estándar)
- ts-nativo: `1.4.4+38afee6c8 · 1.4.5-canary.1+4ce14ba8b · 0.4.14`
- Docker `29.6.2`. Servidores con 2 CPU y 1g; `hey -c 16` con 4 CPU desde otro contenedor.

## 1 · Construir la imagen desde cero (`docker build --no-cache`, dependencias incluidas)

| lenguaje | qué se cronometra | segundos |
|---|---|---|
| go | go build | 5.4 |
| bun --compile | bun build --compile (Bun 1.4.4) | 3.3 |
| bun --compile --aot | bun build --compile --aot (Bun 1.4.5-canary.1) | 2.6 |
| scriptc | apt install clang + descarga de scriptc 0.4.14 + scriptc build | 45.0 |

## 2 · Lo que despliegas

| imagen | base | tamaño |
|---|---|---|
| go | scratch + binario estático | 2.4 MiB |
| bun | debian:trixie-slim + ejecutable de bun --compile | 79.0 MiB |
| bun-aot | debian:trixie-slim + ejecutable de bun --compile --aot | 63.4 MiB |
| scriptc | debian:trixie-slim + binario de scriptc | 28.8 MiB |

El artefacto que se copia a la imagen: Go, 5.6 MiB; bun, 104.07 MiB; bun-aot, 69.05 MiB; scriptc, 0.72 MiB.

## Estado de la máquina antes de medir

- /proc/loadavg (desde un contenedor alpine, VM de Docker con 24 hilos): `2.33 2.46 2.04 4/1137 6`
- contenedores ajenos en marcha (CPU en ese momento): jk-a-fijo 0.51%, jk-a-jenkins 0.59%, jk-b-jenkins-877 0.16%, jk-b-k8s-control-plane 4.79%

## 0 · ¿Hacen lo mismo? (antes de medir)

- go: total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go
- bun --compile: total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go
- bun --compile --aot: total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go
- scriptc: total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go

## 3 · Arranque: contenedor nuevo → primera respuesta 200 (mediana de 8) y memoria en reposo

| variante | 1ª respuesta (docker incluido) | «listo» (su propio reloj) | memoria en reposo |
|---|---|---|---|
| go | 188 ms | 0.0 ms | 2.8 MiB |
| bun --compile | 186 ms | 3.6 ms | 7.4 MiB |
| bun --compile --aot | 199 ms | 1.3 ms | 4.7 MiB |
| scriptc | 196 ms | 0.1 ms | 1.5 MiB |

## 4 · Carga con memoria: /trabajo (objetos, mapa, orden, JSON), a tope

`hey -c 16` desde otro contenedor; servidor con 2 CPU y 1g. «En frío»: los primeros 10 s tras arrancar. «Templado»: los 20 s siguientes.

| variante | peticiones/s en frío | peticiones/s templado | p50 | p99 | memoria máx. bajo carga |
|---|---|---|---|---|---|
| go | 476 | 475 | 26.8 ms | 117.9 ms | 30 MiB |
| bun --compile | 254 | 257 | 58.2 ms | 121.1 ms | 65 MiB |
| bun --compile --aot | 512 | 502 | 29.9 ms | 61.2 ms | 42 MiB |
| scriptc | 548 | 533 | 28.1 ms | 57.3 ms | 6 MiB |

## 5 · Carga de CPU pura: /suma, a tope

`hey -c 16` desde otro contenedor; servidor con 2 CPU y 1g. «En frío»: los primeros 10 s tras arrancar. «Templado»: los 20 s siguientes.

| variante | peticiones/s en frío | peticiones/s templado | p50 | p99 | memoria máx. bajo carga |
|---|---|---|---|---|---|
| go | 14877 | 14893 | 1.0 ms | 2.7 ms | 12 MiB |
| bun --compile | 605 | 609 | 24.6 ms | 49.9 ms | 15 MiB |
| bun --compile --aot | 4701 | 3564 | 3.2 ms | 30.3 ms | 9 MiB |
| scriptc | 2436 | 3279 | 4.6 ms | 9.3 ms | 3 MiB |

## 6 · Latencia de una petición sola: /trabajo

Un solo cliente a 5 peticiones/s (`hey -c 1 -q 5`), 20 s, tras 10 s de calentamiento a tope. Cada petición llega sola: no hay cola, es lo que tarda el servidor en atender una.

| variante | peticiones/s logradas | p50 | p99 |
|---|---|---|---|
| go | 5 | 4.50 ms | 8.00 ms |
| bun --compile | 5 | 5.90 ms | 10.50 ms |
| bun --compile --aot | 5 | 3.10 ms | 6.60 ms |
| scriptc | 5 | 3.00 ms | 5.80 ms |

## 7 · Latencia de una petición sola: /suma

Un solo cliente a 5 peticiones/s (`hey -c 1 -q 5`), 20 s, tras 10 s de calentamiento a tope. Cada petición llega sola: no hay cola, es lo que tarda el servidor en atender una.

| variante | peticiones/s logradas | p50 | p99 |
|---|---|---|---|
| go | 5 | 0.50 ms | 1.40 ms |
| bun --compile | 5 | 2.30 ms | 4.50 ms |
| bun --compile --aot | 5 | 0.70 ms | 1.40 ms |
| scriptc | 5 | 0.70 ms | 1.30 ms |

## 8 · Cómputo puro: el reloj del propio proceso, sin red

50 peticiones de calentamiento y 300 medidas, de una en una; el tiempo es el campo «ns» que mide cada servidor alrededor de su propio cálculo.

| variante | /suma?n=200000 mediana | p99 | /trabajo?n=20000 mediana | p99 |
|---|---|---|---|---|
| go | 0.122 ms | 0.162 ms | 3.376 ms | 3.983 ms |
| bun --compile | 1.623 ms | 2.364 ms | 7.463 ms | 13.798 ms |
| bun --compile --aot | 0.198 ms | 0.327 ms | 2.041 ms | 3.332 ms |
| scriptc | 0.299 ms | 0.354 ms | 2.268 ms | 3.180 ms |

Al terminar: /proc/loadavg `0.64 1.22 1.63 4/1142 6`; ajenos: jk-a-fijo 0.54%, jk-a-jenkins 0.63%, jk-b-jenkins-877 0.18%, jk-b-k8s-control-plane 6.31%

