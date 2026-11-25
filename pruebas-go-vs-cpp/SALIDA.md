# SALIDA — ¿Go o C++? (golandia), con el banco de la revancha

> Corrida del 2026-10-24 08:55. Regenerable con `python investigacion/pruebas_go_vs_x.py cpp`. ⛔ Ninguna cifra de blog: todo salió de aquí.

- go: `go version go1.27.2 linux/amd64` (net/http, solo librería estándar)
- cpp: `g++ (GCC) 14.4.0` — scratch + binario estático (g++ 14.4, cpp-httplib 0.63.0)
- Docker `29.6.2`. Servidores con 2 CPU y 1g; `hey -c 16` con 4 CPU desde otro contenedor.

## 1 · Construir la imagen desde cero (`docker build --no-cache`, dependencias incluidas)

| lenguaje | qué se cronometra | segundos |
|---|---|---|
| go | go build | 5.4 |
| cpp | g++ -O2 -static | 8.7 |

## 2 · Lo que despliegas

| imagen | base | tamaño |
|---|---|---|
| go | scratch + binario estático | 2.4 MiB |
| cpp | scratch + binario estático (g++ 14.4, cpp-httplib 0.63.0) | 1.0 MiB |

## 0 · ¿Hacen lo mismo? (antes de medir)

- go: total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go
- c++ (2 hilos): total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go
- c++ (32 hilos): total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go

## 3 · Arranque: contenedor nuevo → primera respuesta 200 (mediana de 8) y memoria en reposo

| variante | 1ª respuesta (docker incluido) | «listo» (su propio reloj) | memoria en reposo |
|---|---|---|---|
| go | 209 ms | 0.0 ms | 3.2 MiB |
| c++ (2 hilos) | 215 ms | 0.0 ms | 1.8 MiB |
| c++ (32 hilos) | 220 ms | 0.0 ms | 3.3 MiB |

## 4 · Carga con memoria: /trabajo (objetos, mapa, orden, JSON), a tope

`hey -c 16` desde otro contenedor; servidor con 2 CPU y 1g. «En frío»: los primeros 10 s tras arrancar. «Templado»: los 20 s siguientes.

| variante | peticiones/s en frío | peticiones/s templado | p50 | p99 | memoria máx. bajo carga |
|---|---|---|---|---|---|
| go | 448 | 457 | 28.2 ms | 127.5 ms | 41 MiB |
| c++ (2 hilos) | 187 | 286 | 3.4 ms | 1072.4 ms | 9 MiB |
| c++ (32 hilos) | 348 | 354 | 6.3 ms | 194.5 ms | 97 MiB |

## 5 · Carga de CPU pura: /suma, a tope

`hey -c 16` desde otro contenedor; servidor con 2 CPU y 1g. «En frío»: los primeros 10 s tras arrancar. «Templado»: los 20 s siguientes.

| variante | peticiones/s en frío | peticiones/s templado | p50 | p99 | memoria máx. bajo carga |
|---|---|---|---|---|---|
| go | 14888 | 14600 | 1.0 ms | 2.8 ms | 11 MiB |
| c++ (2 hilos) | 1564 | 3568 | 0.2 ms | 80.5 ms | 4 MiB |
| c++ (32 hilos) | 10242 | 10355 | 0.3 ms | 74.1 ms | 7 MiB |

## 6 · Latencia de una petición sola: /trabajo

Un solo cliente a 5 peticiones/s (`hey -c 1 -q 5`), 20 s, tras 10 s de calentamiento a tope. Cada petición llega sola: no hay cola, es lo que tarda el servidor en atender una.

| variante | peticiones/s logradas | p50 | p99 |
|---|---|---|---|
| go | 5 | 4.00 ms | 4.80 ms |
| c++ (2 hilos) | 5 | 3.70 ms | 4.70 ms |
| c++ (32 hilos) | 5 | 3.80 ms | 4.70 ms |

## 7 · Latencia de una petición sola: /suma

Un solo cliente a 5 peticiones/s (`hey -c 1 -q 5`), 20 s, tras 10 s de calentamiento a tope. Cada petición llega sola: no hay cola, es lo que tarda el servidor en atender una.

| variante | peticiones/s logradas | p50 | p99 |
|---|---|---|---|
| go | 5 | 0.50 ms | 1.20 ms |
| c++ (2 hilos) | 5 | 0.50 ms | 1.80 ms |
| c++ (32 hilos) | 5 | 0.40 ms | 1.50 ms |


## Anexo · los núcleos que ve C++ dentro del contenedor

`docker run --cpus 2 golandia-govs-cpp:1` sin HILOS (usa `std::thread::hardware_concurrency()`), 2026-10-24:

```
LISTO_EN_MS 0.036 cpp hilos=24
```

El contenedor tiene 2 CPU; `hardware_concurrency()` devuelve los 24 de la máquina.
