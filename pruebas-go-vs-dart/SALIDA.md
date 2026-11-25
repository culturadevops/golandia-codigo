# SALIDA — ¿Go o Dart? (golandia 805), con el banco de la revancha

> Corrida del 2026-11-05 09:26, con el candado `_logs/banco.lock` tomado. Regenerable con `python investigacion/pruebas_go_vs_lote_2026_11.py dart` (mismo banco que 651-654 y 699). ⛔ Ninguna cifra de blog: todo salió de aquí.

- go: `go version go1.27.2 linux/amd64` (net/http, solo librería estándar)
- dart: `Dart SDK version: 3.13.7 (stable) (Thu Oct 22 14:28:57 2026 -0700) on "linux_x64"`
- Docker `29.6.2`. Servidores con 2 CPU y 1g; `hey -c 16` con 4 CPU desde otro contenedor.

## 1 · Construir la imagen desde cero (`docker build --no-cache`, dependencias incluidas)

| lenguaje | qué se cronometra | segundos |
|---|---|---|
| go | go build | 6.1 |
| dart compile exe | dart pub get + dart compile exe (AOT) | 2.3 |
| dart run | dart pub get (JIT, sin compilar) | 0.7 |

## 2 · Lo que despliegas

| imagen | base | tamaño |
|---|---|---|
| go | scratch + binario estático | 2.4 MiB |
| dart | scratch + /runtime de dart + binario de dart compile exe | 4.6 MiB |
| dart-jit | dart:3.13.7 (SDK completo) + código fuente | 302.1 MiB |

El artefacto que se copia a la imagen: Go, 5.6 MiB; dart, 6.69 MiB.

## 0 · ¿Hacen lo mismo? (antes de medir)

- go: total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go
- dart AOT (1 isolate): total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go
- dart AOT (2 isolates): total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go
- dart run JIT (2 isolates): total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go

## 3 · Arranque: contenedor nuevo → primera respuesta 200 (mediana de 8) y memoria en reposo

| variante | 1ª respuesta (docker incluido) | «listo» (su propio reloj) | memoria en reposo |
|---|---|---|---|
| go | 212 ms | 0.0 ms | 3.0 MiB |
| dart AOT (1 isolate) | 217 ms | — | 3.3 MiB |
| dart AOT (2 isolates) | 211 ms | — | 3.5 MiB |
| dart run JIT (2 isolates) | 391 ms | — | 111.0 MiB |

## 4 · Carga con memoria: /trabajo (objetos, mapa, orden, JSON), a tope

`hey -c 16` desde otro contenedor; servidor con 2 CPU y 1g. «En frío»: los primeros 10 s tras arrancar. «Templado»: los 20 s siguientes.

| variante | peticiones/s en frío | peticiones/s templado | p50 | p99 | memoria máx. bajo carga |
|---|---|---|---|---|---|
| go | 440 | 443 | 29.4 ms | 127.5 ms | 39 MiB |
| dart AOT (1 isolate) | 197 | 218 | 73.0 ms | 78.5 ms | 46 MiB |
| dart AOT (2 isolates) | 318 | 321 | 48.7 ms | 72.1 ms | 65 MiB |
| dart run JIT (2 isolates) | 268 | 274 | 60.1 ms | 73.4 ms | 165 MiB |

## 5 · Carga de CPU pura: /suma, a tope

`hey -c 16` desde otro contenedor; servidor con 2 CPU y 1g. «En frío»: los primeros 10 s tras arrancar. «Templado»: los 20 s siguientes.

| variante | peticiones/s en frío | peticiones/s templado | p50 | p99 | memoria máx. bajo carga |
|---|---|---|---|---|---|
| go | 14763 | 14699 | 1.0 ms | 2.7 ms | 10 MiB |
| dart AOT (1 isolate) | 6692 | 6678 | 2.4 ms | 3.0 ms | 11 MiB |
| dart AOT (2 isolates) | 13119 | 13190 | 1.2 ms | 2.1 ms | 11 MiB |
| dart run JIT (2 isolates) | 4662 | 4842 | 3.5 ms | 4.4 ms | 122 MiB |

## 6 · Latencia de una petición sola: /trabajo

Un solo cliente a 5 peticiones/s (`hey -c 1 -q 5`), 20 s, tras 10 s de calentamiento a tope. Cada petición llega sola: no hay cola, es lo que tarda el servidor en atender una.

| variante | peticiones/s logradas | p50 | p99 |
|---|---|---|---|
| go | 5 | 4.00 ms | 6.00 ms |
| dart AOT (1 isolate) | 5 | 5.70 ms | 8.00 ms |
| dart AOT (2 isolates) | 5 | 5.40 ms | 9.60 ms |
| dart run JIT (2 isolates) | 5 | 6.10 ms | 8.20 ms |

## 7 · Latencia de una petición sola: /suma

Un solo cliente a 5 peticiones/s (`hey -c 1 -q 5`), 20 s, tras 10 s de calentamiento a tope. Cada petición llega sola: no hay cola, es lo que tarda el servidor en atender una.

| variante | peticiones/s logradas | p50 | p99 |
|---|---|---|---|
| go | 5 | 0.50 ms | 1.20 ms |
| dart AOT (1 isolate) | 5 | 0.60 ms | 1.20 ms |
| dart AOT (2 isolates) | 5 | 0.60 ms | 1.30 ms |
| dart run JIT (2 isolates) | 5 | 1.00 ms | 2.20 ms |

## 8 · Cómputo puro: el reloj del propio proceso, sin red

50 peticiones de calentamiento y 300 medidas, de una en una; el tiempo es el campo «ns» que mide cada servidor alrededor de su propio cálculo.

| variante | /suma?n=200000 mediana | p99 | /trabajo?n=20000 mediana | p99 |
|---|---|---|---|---|
| go | 0.122 ms | 0.254 ms | 3.454 ms | 5.328 ms |
| dart AOT (1 isolate) | 0.123 ms | 0.153 ms | 5.019 ms | 7.628 ms |
| dart AOT (2 isolates) | 0.120 ms | 0.266 ms | 4.790 ms | 6.743 ms |
| dart run JIT (2 isolates) | 0.369 ms | 0.496 ms | 4.551 ms | 8.825 ms |

