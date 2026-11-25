# PRIMERA CORRIDA (guardada para ver si se repite; la que vale es SALIDA.md, corrida de 2026-10-29 07:51). Las dos corrieron mientras otro agente medía en la misma máquina (contenedores v698-*): Go dio 401-440 pet/s en /trabajo, menos que los 451-475 de 651-654.

# SALIDA — ¿Go o TypeScript con Deno? (golandia 699), con el banco de la revancha

> Corrida del 2026-10-29 07:39. Regenerable con `python investigacion/pruebas_go_vs_deno.py` (mismo banco que 651-654: `pruebas_go_vs_x.py` + `pruebas_go_vs_x_computo.py`). ⛔ Ninguna cifra de blog: todo salió de aquí.

- go: `go version go1.27.2 linux/amd64` (net/http, solo librería estándar)
- deno: `deno 2.9.8 (stable, release, x86_64-unknown-linux-gnu) · v8 15.0.245.2-rusty · typescript 6.0.3` — denoland/deno:2.9.8, Deno.serve, sin dependencias
- Docker `29.6.2`. Servidores con 2 CPU y 1g; `hey -c 16` con 4 CPU desde otro contenedor.

## 1 · Construir la imagen desde cero (`docker build --no-cache`, dependencias incluidas)

| lenguaje | qué se cronometra | segundos |
|---|---|---|
| go | go build | 5.4 |
| deno | deno check (tipos) sobre denoland/deno:2.9.8 | 1.2 |
| deno compile | deno compile (tipos + binario) | 6.5 |

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
| go | 200 ms | 0.0 ms | 2.7 MiB |
| deno run (1 hilo) | 250 ms | 5.7 ms | 20.1 MiB |
| deno serve --parallel (2 hilos) | 279 ms | — | 30.6 MiB |
| deno compile (binario) | 254 ms | 5.7 ms | 13.0 MiB |

## 4 · Carga con memoria: /trabajo (objetos, mapa, orden, JSON), a tope

`hey -c 16` desde otro contenedor; servidor con 2 CPU y 1g. «En frío»: los primeros 10 s tras arrancar. «Templado»: los 20 s siguientes.

| variante | peticiones/s en frío | peticiones/s templado | p50 | p99 | memoria máx. bajo carga |
|---|---|---|---|---|---|
| go | 412 | 440 | 29.5 ms | 125.1 ms | 34 MiB |
| deno run (1 hilo) | 248 | 238 | 66.8 ms | 77.6 ms | 97 MiB |
| deno serve --parallel (2 hilos) | 427 | 412 | 38.5 ms | 71.7 ms | 146 MiB |
| deno compile (binario) | 303 | 295 | 53.5 ms | 74.4 ms | 257 MiB |

## 5 · Carga de CPU pura: /suma, a tope

`hey -c 16` desde otro contenedor; servidor con 2 CPU y 1g. «En frío»: los primeros 10 s tras arrancar. «Templado»: los 20 s siguientes.

| variante | peticiones/s en frío | peticiones/s templado | p50 | p99 | memoria máx. bajo carga |
|---|---|---|---|---|---|
| go | 13299 | 13004 | 1.1 ms | 3.3 ms | 9 MiB |
| deno run (1 hilo) | 778 | 820 | 19.3 ms | 22.2 ms | 25 MiB |
| deno serve --parallel (2 hilos) | 1574 | 1568 | 9.8 ms | 13.7 ms | 40 MiB |
| deno compile (binario) | 791 | 813 | 19.4 ms | 22.7 ms | 20 MiB |

## 6 · Latencia de una petición sola: /trabajo

Un solo cliente a 5 peticiones/s (`hey -c 1 -q 5`), 20 s, tras 10 s de calentamiento a tope. Cada petición llega sola: no hay cola, es lo que tarda el servidor en atender una.

| variante | peticiones/s logradas | p50 | p99 |
|---|---|---|---|
| go | 5 | 3.80 ms | 5.60 ms |
| deno run (1 hilo) | 5 | 4.30 ms | 7.60 ms |
| deno serve --parallel (2 hilos) | 5 | 4.20 ms | 6.80 ms |
| deno compile (binario) | 5 | 4.20 ms | 5.20 ms |

## 7 · Latencia de una petición sola: /suma

Un solo cliente a 5 peticiones/s (`hey -c 1 -q 5`), 20 s, tras 10 s de calentamiento a tope. Cada petición llega sola: no hay cola, es lo que tarda el servidor en atender una.

| variante | peticiones/s logradas | p50 | p99 |
|---|---|---|---|
| go | 5 | 0.50 ms | 1.10 ms |
| deno run (1 hilo) | 5 | 1.50 ms | 2.60 ms |
| deno serve --parallel (2 hilos) | 5 | 1.50 ms | 3.10 ms |
| deno compile (binario) | 5 | 1.50 ms | 3.50 ms |

## 8 · Cómputo puro: el reloj del propio proceso, sin red

El método de `pruebas_go_vs_x_computo.py`: 50 peticiones de calentamiento y 300 medidas, de una en una; el tiempo es el campo «ns» que mide cada servidor alrededor de su propio cálculo.

| variante | /suma?n=200000 mediana | p99 | /trabajo?n=20000 mediana | p99 |
|---|---|---|---|---|
| go | 0.122 ms | 0.202 ms | 3.423 ms | 3.942 ms |
| deno run (1 hilo) | 1.171 ms | 1.695 ms | 3.774 ms | 5.430 ms |
| deno serve --parallel (2 hilos) | 1.178 ms | 1.889 ms | 3.786 ms | 5.271 ms |
| deno compile (binario) | 1.170 ms | 2.312 ms | 3.373 ms | 4.691 ms |

