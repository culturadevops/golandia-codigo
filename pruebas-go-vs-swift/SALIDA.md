# SALIDA — ¿Go o Swift? (golandia), con el banco de la revancha

> Corrida del 2026-10-24 09:13. Regenerable con `python investigacion/pruebas_go_vs_x.py swift`. ⛔ Ninguna cifra de blog: todo salió de aquí.

- go: `go version go1.27.2 linux/amd64` (net/http, solo librería estándar)
- swift: `Swift version 6.4.1 (swift-6.4.1-RELEASE)` — ubuntu:noble + binario (Swift 6.4.1, Vapor 4.122.3, stdlib estática)
- Docker `29.6.2`. Servidores con 2 CPU y 1g; `hey -c 16` con 4 CPU desde otro contenedor.

## 1 · Construir la imagen desde cero (`docker build --no-cache`, dependencias incluidas)

| lenguaje | qué se cronometra | segundos |
|---|---|---|
| go | go build | 6.4 |
| swift | swift package resolve + swift build -c release | 154.3 |

## 2 · Lo que despliegas

| imagen | base | tamaño |
|---|---|---|
| go | scratch + binario estático | 2.4 MiB |
| swift | ubuntu:noble + binario (Swift 6.4.1, Vapor 4.122.3, stdlib estática) | 62.3 MiB |

## 0 · ¿Hacen lo mismo? (antes de medir)

- go: total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go
- swift (Vapor): total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go

## 3 · Arranque: contenedor nuevo → primera respuesta 200 (mediana de 8) y memoria en reposo

| variante | 1ª respuesta (docker incluido) | «listo» (su propio reloj) | memoria en reposo |
|---|---|---|---|
| go | 214 ms | 0.0 ms | 3.0 MiB |
| swift (Vapor) | 216 ms | 2.5 ms | 5.6 MiB |

## 4 · Carga con memoria: /trabajo (objetos, mapa, orden, JSON), a tope

`hey -c 16` desde otro contenedor; servidor con 2 CPU y 1g. «En frío»: los primeros 10 s tras arrancar. «Templado»: los 20 s siguientes.

| variante | peticiones/s en frío | peticiones/s templado | p50 | p99 | memoria máx. bajo carga |
|---|---|---|---|---|---|
| go | 455 | 451 | 28.9 ms | 121.2 ms | 36 MiB |
| swift (Vapor) | 599 | 598 | 23.4 ms | 48.9 ms | 11 MiB |

## 5 · Carga de CPU pura: /suma, a tope

`hey -c 16` desde otro contenedor; servidor con 2 CPU y 1g. «En frío»: los primeros 10 s tras arrancar. «Templado»: los 20 s siguientes.

| variante | peticiones/s en frío | peticiones/s templado | p50 | p99 | memoria máx. bajo carga |
|---|---|---|---|---|---|
| go | 14812 | 14767 | 1.0 ms | 2.7 ms | 11 MiB |
| swift (Vapor) | 13743 | 13719 | 1.0 ms | 2.3 ms | 5 MiB |

## 6 · Latencia de una petición sola: /trabajo

Un solo cliente a 5 peticiones/s (`hey -c 1 -q 5`), 20 s, tras 10 s de calentamiento a tope. Cada petición llega sola: no hay cola, es lo que tarda el servidor en atender una.

| variante | peticiones/s logradas | p50 | p99 |
|---|---|---|---|
| go | 5 | 4.00 ms | 5.20 ms |
| swift (Vapor) | 5 | 4.10 ms | 5.10 ms |

## 7 · Latencia de una petición sola: /suma

Un solo cliente a 5 peticiones/s (`hey -c 1 -q 5`), 20 s, tras 10 s de calentamiento a tope. Cada petición llega sola: no hay cola, es lo que tarda el servidor en atender una.

| variante | peticiones/s logradas | p50 | p99 |
|---|---|---|---|
| go | 5 | 0.50 ms | 1.40 ms |
| swift (Vapor) | 5 | 0.80 ms | 1.70 ms |


## Anexo · los núcleos que ve Swift dentro del contenedor

`docker run --cpus 2 golandia-govs-swift:1` (`System.coreCount` de SwiftNIO, con el que Vapor dimensiona sus bucles de eventos), 2026-10-24:

```
LISTO_EN_MS 3.256 swift hilos=2
```

El contenedor tiene 2 CPU y Swift ve 2: SwiftNIO respeta el límite del contenedor (sin `--cpus`, la misma imagen dice 24). En C++, `hardware_concurrency()` NO lo respeta (ver pruebas-go-vs-cpp/SALIDA.md).
