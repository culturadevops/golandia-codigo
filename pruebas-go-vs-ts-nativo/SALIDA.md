# SALIDA — ¿Go o TypeScript compilado a nativo? (golandia 880), corrida 2, con el banco de la serie

> Corrida del 2026-11-18 10:33, con el candado `_logs/banco.lock` tomado. Regenerable con `python investigacion/pruebas_go_vs_nativo_2026_11.py ts-nativo 2` (mismo banco que 865). ⛔ Ninguna cifra de blog: todo salió de aquí.

- go: `go version go1.27.2 linux/amd64` (net/http, solo librería estándar)
- ts-nativo: `1.4.4+38afee6c8 · 1.4.5-canary.1+4ce14ba8b · 0.4.14`
- Docker `29.6.2`. Servidores con 2 CPU y 1g; `hey -c 16` con 4 CPU desde otro contenedor.

## 1 · Construir la imagen desde cero (`docker build --no-cache`, dependencias incluidas)

| lenguaje | qué se cronometra | segundos |
|---|---|---|
| go | go build | 7.5 |
| bun --compile | bun build --compile (Bun 1.4.4) | 4.9 |
| bun --compile --aot | bun build --compile --aot (Bun 1.4.5-canary.1) | 4.9 |
| scriptc | apt install clang + descarga de scriptc 0.4.14 + scriptc build | 49.0 |

## 2 · Lo que despliegas

| imagen | base | tamaño |
|---|---|---|
| go | scratch + binario estático | 2.4 MiB |
| bun | debian:trixie-slim + ejecutable de bun --compile | 79.0 MiB |
| bun-aot | debian:trixie-slim + ejecutable de bun --compile --aot | 63.4 MiB |
| scriptc | debian:trixie-slim + binario de scriptc | 28.8 MiB |

El artefacto que se copia a la imagen: Go, 5.6 MiB; bun, 104.07 MiB; bun-aot, 69.05 MiB; scriptc, 0.72 MiB.

## Estado de la máquina antes de medir

- /proc/loadavg (desde un contenedor alpine, VM de Docker con 24 hilos): `1.74 1.44 1.67 1/1143 6`
- contenedores ajenos en marcha (CPU en ese momento): jk-a-fijo 0.52%, jk-a-jenkins 0.64%, jk-b-jenkins-877 0.19%, jk-b-k8s-control-plane 6.62%

## 0 · ¿Hacen lo mismo? (antes de medir)

- go: total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go
- bun --compile: total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go
- bun --compile --aot: total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go
- scriptc: total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go

## 3 · Arranque: contenedor nuevo → primera respuesta 200 (mediana de 8) y memoria en reposo

| variante | 1ª respuesta (docker incluido) | «listo» (su propio reloj) | memoria en reposo |
|---|---|---|---|
| go | 195 ms | 0.0 ms | 2.6 MiB |
| bun --compile | 211 ms | 4.1 ms | 7.2 MiB |
| bun --compile --aot | 208 ms | 1.3 ms | 4.4 MiB |
| scriptc | 194 ms | 0.1 ms | 1.4 MiB |

## 4 · Carga con memoria: /trabajo (objetos, mapa, orden, JSON), a tope

`hey -c 16` desde otro contenedor; servidor con 2 CPU y 1g. «En frío»: los primeros 10 s tras arrancar. «Templado»: los 20 s siguientes.

| variante | peticiones/s en frío | peticiones/s templado | p50 | p99 | memoria máx. bajo carga |
|---|---|---|---|---|---|
| go | 470 | 469 | 27.5 ms | 120.4 ms | 36 MiB |
| bun --compile | 251 | 261 | 57.3 ms | 118.2 ms | 57 MiB |
| bun --compile --aot | 497 | 451 | 32.9 ms | 71.0 ms | 45 MiB |
| scriptc | 546 | 535 | 28.0 ms | 57.2 ms | 6 MiB |

## 5 · Carga de CPU pura: /suma, a tope

`hey -c 16` desde otro contenedor; servidor con 2 CPU y 1g. «En frío»: los primeros 10 s tras arrancar. «Templado»: los 20 s siguientes.

| variante | peticiones/s en frío | peticiones/s templado | p50 | p99 | memoria máx. bajo carga |
|---|---|---|---|---|---|
| go | 14931 | 14843 | 1.0 ms | 2.7 ms | 10 MiB |
| bun --compile | 606 | 598 | 24.7 ms | 50.6 ms | 15 MiB |
| bun --compile --aot | 4724 | 4737 | 3.1 ms | 6.4 ms | 10 MiB |
| scriptc | 3279 | 3280 | 4.6 ms | 9.3 ms | 3 MiB |

## 6 · Latencia de una petición sola: /trabajo

Un solo cliente a 5 peticiones/s (`hey -c 1 -q 5`), 20 s, tras 10 s de calentamiento a tope. Cada petición llega sola: no hay cola, es lo que tarda el servidor en atender una.

| variante | peticiones/s logradas | p50 | p99 |
|---|---|---|---|
| go | 5 | 4.30 ms | 6.70 ms |
| bun --compile | 5 | 6.10 ms | 9.30 ms |
| bun --compile --aot | 5 | 2.90 ms | 4.60 ms |
| scriptc | 5 | 2.80 ms | 3.90 ms |

## 7 · Latencia de una petición sola: /suma

Un solo cliente a 5 peticiones/s (`hey -c 1 -q 5`), 20 s, tras 10 s de calentamiento a tope. Cada petición llega sola: no hay cola, es lo que tarda el servidor en atender una.

| variante | peticiones/s logradas | p50 | p99 |
|---|---|---|---|
| go | 5 | 0.50 ms | 1.80 ms |
| bun --compile | 5 | 2.10 ms | 4.50 ms |
| bun --compile --aot | 5 | 0.60 ms | 1.60 ms |
| scriptc | 5 | 0.70 ms | 1.60 ms |

## 8 · Cómputo puro: el reloj del propio proceso, sin red

50 peticiones de calentamiento y 300 medidas, de una en una; el tiempo es el campo «ns» que mide cada servidor alrededor de su propio cálculo.

| variante | /suma?n=200000 mediana | p99 | /trabajo?n=20000 mediana | p99 |
|---|---|---|---|---|
| go | 0.122 ms | 0.153 ms | 3.290 ms | 3.650 ms |
| bun --compile | 1.617 ms | 2.067 ms | 4.313 ms | 6.541 ms |
| bun --compile --aot | 0.198 ms | 0.235 ms | 1.811 ms | 2.873 ms |
| scriptc | 0.293 ms | 0.332 ms | 1.724 ms | 2.416 ms |

Al terminar: /proc/loadavg `0.50 0.96 1.36 1/1142 6`; ajenos: jk-a-fijo 0.52%, jk-a-jenkins 0.60%, jk-b-jenkins-877 0.10%, jk-b-k8s-control-plane 7.40%

