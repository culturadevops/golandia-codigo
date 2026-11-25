# SALIDA — ¿Go o TinyGo? (golandia 808), con el banco de la revancha

> Corrida del 2026-11-05 10:41, con el candado `_logs/banco.lock` tomado. Regenerable con `python investigacion/pruebas_go_vs_lote_2026_11.py tinygo` (mismo banco que 651-654 y 699). ⛔ Ninguna cifra de blog: todo salió de aquí.

- go: `go version go1.27.2 linux/amd64` (net/http, solo librería estándar)
- tinygo: `tinygo version 0.44.0 linux/amd64 (using go version go1.27.3 and LLVM version 22.1.4)`
- Docker `29.6.2`. Servidores con 2 CPU y 1g; `hey -c 16` con 4 CPU desde otro contenedor.

## 1 · Construir la imagen desde cero (`docker build --no-cache`, dependencias incluidas)

| lenguaje | qué se cronometra | segundos |
|---|---|---|
| go | go build | 5.4 |
| tinygo | tinygo build -no-debug (por defecto, -opt=z) | 15.6 |
| tinygo -opt=2 | tinygo build -no-debug -opt=2 | 19.2 |

## 2 · Lo que despliegas

| imagen | base | tamaño |
|---|---|---|
| go | scratch + binario estático | 2.4 MiB |
| tinygo | scratch + binario de tinygo (-opt=z) | 0.3 MiB |
| tinygo-opt2 | scratch + binario de tinygo -opt=2 | 0.4 MiB |

El artefacto que se copia a la imagen: Go, 5.6 MiB; tinygo, 1.21 MiB; tinygo-opt2, 1.36 MiB.

## 0 · ¿Hacen lo mismo? (antes de medir)

- go: total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go
- tinygo (-opt=z): total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go
- tinygo -opt=2: total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go

## 3 · Arranque: contenedor nuevo → primera respuesta 200 (mediana de 8) y memoria en reposo

| variante | 1ª respuesta (docker incluido) | «listo» (su propio reloj) | memoria en reposo |
|---|---|---|---|
| go | 208 ms | 0.0 ms | 3.0 MiB |
| tinygo (-opt=z) | 206 ms | 0.0 ms | 1.8 MiB |
| tinygo -opt=2 | 219 ms | 0.0 ms | 2.2 MiB |

## 4 · Carga con memoria: /trabajo (objetos, mapa, orden, JSON), a tope

`hey -c 16` desde otro contenedor; servidor con 2 CPU y 1g. «En frío»: los primeros 10 s tras arrancar. «Templado»: los 20 s siguientes.

| variante | peticiones/s en frío | peticiones/s templado | p50 | p99 | memoria máx. bajo carga |
|---|---|---|---|---|---|
| go | 475 | 475 | 26.8 ms | 119.7 ms | 35 MiB |
| tinygo (-opt=z) | 19 | 19 | 995.3 ms | 1307.6 ms | 54 MiB |
| tinygo -opt=2 | 21 | 21 | 881.1 ms | 1197.6 ms | 77 MiB |

## 5 · Carga de CPU pura: /suma, a tope

`hey -c 16` desde otro contenedor; servidor con 2 CPU y 1g. «En frío»: los primeros 10 s tras arrancar. «Templado»: los 20 s siguientes.

| variante | peticiones/s en frío | peticiones/s templado | p50 | p99 | memoria máx. bajo carga |
|---|---|---|---|---|---|
| go | 15013 | 15016 | 1.0 ms | 2.6 ms | 10 MiB |
| tinygo (-opt=z) | 3690 | 3672 | 1.6 ms | 66.8 ms | 11 MiB |
| tinygo -opt=2 | 8125 | 8228 | 1.4 ms | 30.1 ms | 11 MiB |

## 6 · Latencia de una petición sola: /trabajo

Un solo cliente a 5 peticiones/s (`hey -c 1 -q 5`), 20 s, tras 10 s de calentamiento a tope. Cada petición llega sola: no hay cola, es lo que tarda el servidor en atender una.

| variante | peticiones/s logradas | p50 | p99 |
|---|---|---|---|
| go | 5 | 5.30 ms | 11.20 ms |
| tinygo (-opt=z) | 5 | 5.80 ms | 12.80 ms |
| tinygo -opt=2 | 5 | 5.70 ms | 12.00 ms |

## 7 · Latencia de una petición sola: /suma

Un solo cliente a 5 peticiones/s (`hey -c 1 -q 5`), 20 s, tras 10 s de calentamiento a tope. Cada petición llega sola: no hay cola, es lo que tarda el servidor en atender una.

| variante | peticiones/s logradas | p50 | p99 |
|---|---|---|---|
| go | 5 | 0.70 ms | 3.10 ms |
| tinygo (-opt=z) | 5 | 1.50 ms | 6.30 ms |
| tinygo -opt=2 | 5 | 1.10 ms | 4.40 ms |

## 8 · Cómputo puro: el reloj del propio proceso, sin red

50 peticiones de calentamiento y 300 medidas, de una en una; el tiempo es el campo «ns» que mide cada servidor alrededor de su propio cálculo.

| variante | /suma?n=200000 mediana | p99 | /trabajo?n=20000 mediana | p99 |
|---|---|---|---|---|
| go | 0.122 ms | 0.202 ms | 3.271 ms | 3.792 ms |
| tinygo (-opt=z) | 0.385 ms | 0.518 ms | 3.850 ms | 4.168 ms |
| tinygo -opt=2 | 0.113 ms | 0.153 ms | 3.510 ms | 4.029 ms |


## 9 · Cómo compila TinyGo para Linux (2026-11-05, a mano tras la corrida)

```
$ tinygo info
LLVM triple:       x86_64-unknown-linux-musleabihf
GOOS:              linux
GOARCH:            amd64
build tags:        linux amd64 tinygo purego osusergo math_big_pure_go gc.boehm scheduler.threads tinygo.unwind.setjmp serial.none
garbage collector: boehm
scheduler:         threads
```
