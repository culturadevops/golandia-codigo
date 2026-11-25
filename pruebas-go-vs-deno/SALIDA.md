# SALIDA — ¿Go o TypeScript con Deno? (golandia 699), con el banco de la revancha

> Corrida del 2026-10-29 07:51. Regenerable con `python investigacion/pruebas_go_vs_deno.py` (mismo banco que 651-654: `pruebas_go_vs_x.py` + `pruebas_go_vs_x_computo.py`). ⛔ Ninguna cifra de blog: todo salió de aquí.

- go: `go version go1.27.2 linux/amd64` (net/http, solo librería estándar)
- deno: `deno 2.9.8 (stable, release, x86_64-unknown-linux-gnu) · v8 15.0.245.2-rusty · typescript 6.0.3` — denoland/deno:2.9.8, Deno.serve, sin dependencias
- Docker `29.6.2`. Servidores con 2 CPU y 1g; `hey -c 16` con 4 CPU desde otro contenedor.

## 1 · Construir la imagen desde cero (`docker build --no-cache`, dependencias incluidas)

| lenguaje | qué se cronometra | segundos |
|---|---|---|
| go | go build | 5.6 |
| deno | deno check (tipos) sobre denoland/deno:2.9.8 | 1.2 |
| deno compile | deno compile (tipos + binario) | 6.8 |

## 2 · Lo que despliegas

| imagen | base | tamaño |
|---|---|---|
| go | scratch + binario estático | 2.4 MiB |
| deno | denoland/deno:2.9.8 + 3 archivos .ts | 70.5 MiB |
| deno compile | distroless/cc-debian12 + binario de deno compile | 42.8 MiB |

El binario de Go pesa 5.6 MiB; el de `deno compile`, 100.1 MiB (lleva dentro el runtime de Deno y V8).

## 0 · ¿Hacen lo mismo? (antes de medir)

- go: total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go
- deno run (1 hilo): total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go
- deno serve --parallel (2 hilos): total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go
- deno compile (binario): total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go

## 3 · Arranque: contenedor nuevo → primera respuesta 200 (mediana de 8) y memoria en reposo

| variante | 1ª respuesta (docker incluido) | «listo» (su propio reloj) | memoria en reposo |
|---|---|---|---|
| go | 190 ms | 0.0 ms | 2.7 MiB |
| deno run (1 hilo) | 248 ms | 5.4 ms | 19.7 MiB |
| deno serve --parallel (2 hilos) | 257 ms | — | 30.3 MiB |
| deno compile (binario) | 255 ms | 5.2 ms | 13.7 MiB |

## 4 · Carga con memoria: /trabajo (objetos, mapa, orden, JSON), a tope

`hey -c 16` desde otro contenedor; servidor con 2 CPU y 1g. «En frío»: los primeros 10 s tras arrancar. «Templado»: los 20 s siguientes.

| variante | peticiones/s en frío | peticiones/s templado | p50 | p99 | memoria máx. bajo carga |
|---|---|---|---|---|---|
| go | 420 | 401 | 32.8 ms | 137.7 ms | 36 MiB |
| deno run (1 hilo) | 233 | 235 | 67.3 ms | 80.2 ms | 80 MiB |
| deno serve --parallel (2 hilos) | 444 | 431 | 37.4 ms | 65.3 ms | 144 MiB |
| deno compile (binario) | 287 | 280 | 56.1 ms | 77.8 ms | 292 MiB |

## 5 · Carga de CPU pura: /suma, a tope

`hey -c 16` desde otro contenedor; servidor con 2 CPU y 1g. «En frío»: los primeros 10 s tras arrancar. «Templado»: los 20 s siguientes.

| variante | peticiones/s en frío | peticiones/s templado | p50 | p99 | memoria máx. bajo carga |
|---|---|---|---|---|---|
| go | 14021 | 13382 | 1.1 ms | 3.4 ms | 9 MiB |
| deno run (1 hilo) | 791 | 790 | 19.4 ms | 32.0 ms | 24 MiB |
| deno serve --parallel (2 hilos) | 1450 | 1446 | 12.8 ms | 25.0 ms | 37 MiB |
| deno compile (binario) | 761 | 789 | 19.4 ms | 32.4 ms | 18 MiB |

## 6 · Latencia de una petición sola: /trabajo

Un solo cliente a 5 peticiones/s (`hey -c 1 -q 5`), 20 s, tras 10 s de calentamiento a tope. Cada petición llega sola: no hay cola, es lo que tarda el servidor en atender una.

| variante | peticiones/s logradas | p50 | p99 |
|---|---|---|---|
| go | 5 | 3.80 ms | 9.40 ms |
| deno run (1 hilo) | 5 | 4.40 ms | 7.40 ms |
| deno serve --parallel (2 hilos) | 5 | 4.20 ms | 5.40 ms |
| deno compile (binario) | 5 | 4.10 ms | 6.40 ms |

## 7 · Latencia de una petición sola: /suma

Un solo cliente a 5 peticiones/s (`hey -c 1 -q 5`), 20 s, tras 10 s de calentamiento a tope. Cada petición llega sola: no hay cola, es lo que tarda el servidor en atender una.

| variante | peticiones/s logradas | p50 | p99 |
|---|---|---|---|
| go | 5 | 0.50 ms | 1.10 ms |
| deno run (1 hilo) | 5 | 1.60 ms | 3.80 ms |
| deno serve --parallel (2 hilos) | 5 | 1.60 ms | 2.90 ms |
| deno compile (binario) | 5 | 1.60 ms | 3.40 ms |

## 8 · Cómputo puro: el reloj del propio proceso, sin red

El método de `pruebas_go_vs_x_computo.py`: 50 peticiones de calentamiento y 300 medidas, de una en una; el tiempo es el campo «ns» que mide cada servidor alrededor de su propio cálculo.

| variante | /suma?n=200000 mediana | p99 | /trabajo?n=20000 mediana | p99 |
|---|---|---|---|---|
| go | 0.122 ms | 0.186 ms | 3.486 ms | 4.453 ms |
| deno run (1 hilo) | 1.161 ms | 2.428 ms | 3.551 ms | 4.948 ms |
| deno serve --parallel (2 hilos) | 1.183 ms | 2.866 ms | 3.847 ms | 5.119 ms |
| deno compile (binario) | 1.164 ms | 2.635 ms | 3.526 ms | 4.566 ms |

