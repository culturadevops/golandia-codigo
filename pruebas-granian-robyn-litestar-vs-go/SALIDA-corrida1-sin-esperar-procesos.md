# SALIDA — Granian, Robyn y Litestar contra Go (golandia; pedido por @mdt4177 en el 770 de Cultura DevOps)

> Corrida del 2026-11-09 19:07. Regenerable con `python correr.py`. ⛔ Ninguna cifra de blog: todo salió de aquí. La misma lógica (app/logica.py) en Python y el mismo servidor Go de la serie «¿Go o X?».

- `Python 3.13.17` (python:3.13.17-slim, la base del 770) · `go version go1.27.2 linux/amd64` (FROM scratch, binario estático, solo net/http). Docker `29.6.2`.
- Servidores con 2 CPU y 1g. Python: 2 procesos (uno por CPU). Go: 1 proceso (GOMAXPROCS lo fija el runtime según el límite del contenedor). `hey -c 16` con 4 CPU desde otro contenedor.
- /trabajo?n=20000: crea 20000 objetos, los mete en un dict/mapa, los ordena y arma un JSON a mano. /suma?n=200000: multiplicar y módulo en un bucle. /suma?n=1: el cálculo es casi cero, queda el coste del servidor.
- Rutas Python síncronas (`def`): FastAPI → threadpool; Litestar `sync_to_thread=True` → threadpool; Robyn `def` (su runtime en Rust la llama). Granian con `--interface asgi`: la MISMA app que con uvicorn.

| imagen | dependencias directas (versiones fijadas en req/<fw>.txt) |
|---|---|
| golandia-rust-py-fastapi:1 | fastapi==0.162.0, starlette==1.9.0, pydantic==2.14.1, uvicorn[standard]==0.54.0, granian[uvloop]==2.8.7 |
| golandia-rust-py-litestar:1 | litestar==2.25.0, uvicorn[standard]==0.54.0, granian[uvloop]==2.8.7 |
| golandia-rust-py-robyn:1 | robyn==0.88.0 |
| golandia-govs-go:1 | ninguna: librería estándar (net/http) |

## 1 · Construir la imagen desde cero (`docker build --no-cache`; la imagen base ya descargada no cuenta)

| imagen | paquetes Python | segundos |
|---|---|---|
| fastapi | 21 | 10.0 |
| litestar | 29 | 13.9 |
| robyn | 12 | 8.5 |
| go | 0 | 4.7 |

## 2 · Lo que despliegas

| imagen | tamaño en disco |
|---|---|
| fastapi | 61.1 MiB |
| litestar | 67.4 MiB |
| robyn | 54.1 MiB |
| go | 2.4 MiB |

## 0 · ¿Hacen lo mismo? (antes de medir)

El cuerpo de /trabajo?n=20000, /suma?n=200000 y /suma?n=1, comparado byte a byte con go (net/http) quitando solo "lang" (go o python) y "ns" (el reloj del cálculo).

- go (net/http): total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591, 2220 + 31 + 16 bytes, application/json → idéntico
- fastapi + uvicorn: total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591, 2220 + 31 + 16 bytes, application/json → idéntico
- fastapi + granian: total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591, 2220 + 31 + 16 bytes, application/json → idéntico
- litestar + uvicorn: total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591, 2220 + 31 + 16 bytes, application/json → idéntico
- litestar + granian: total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591, 2220 + 31 + 16 bytes, application/json → idéntico
- robyn: total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591, 2220 + 31 + 16 bytes, application/json → idéntico

## 3 · Arranque: contenedor nuevo → primera respuesta 200 (mediana de 8) y memoria en reposo

| variante | 1ª respuesta (docker incluido) | «listo» (su propio reloj) | memoria en reposo |
|---|---|---|---|
| go (net/http) | 202 ms | 0.0 ms | 2.8 MiB |
| fastapi + uvicorn | 500 ms | 104.7 ms | 95.6 MiB |
| fastapi + granian | 448 ms | 118.1 ms | 78.2 MiB |
| litestar + uvicorn | 483 ms | 97.4 ms | 87.7 MiB |
| litestar + granian | 452 ms | 110.7 ms | 69.8 MiB |
| robyn | 333 ms | 113.6 ms | 34.8 MiB |

## A · Cómputo de UNA petición (mediana del campo ns, 30 de calentamiento + 200 medidas, de una en una)

| variante | /trabajo (20.000 objetos) | /suma (200.000 vueltas) |
|---|---|---|
| go (net/http) | 3.31 ms | 0.122 ms |
| fastapi + uvicorn | 8.70 ms | 8.612 ms |
| fastapi + granian | 8.68 ms | 8.213 ms |
| litestar + uvicorn | 8.60 ms | 8.236 ms |
| litestar + granian | 8.90 ms | 8.269 ms |
| robyn | 8.58 ms | 8.620 ms |

## 4 · Carga con memoria: /trabajo (objetos, dict, orden, JSON), a tope

`hey -c 16` desde otro contenedor; servidor con 2 CPU y 1g. «En frío»: los primeros 10 s tras arrancar. «Templado»: los 20 s siguientes.

| variante | peticiones/s en frío | peticiones/s templado | p50 | p99 | memoria máx. bajo carga |
|---|---|---|---|---|---|
| go (net/http) | 476 | 478 | 26.3 ms | 124.9 ms | 31 MiB |
| fastapi + uvicorn | 152 | 150 | 99.8 ms | 231.4 ms | 169 MiB |
| fastapi + granian | 154 | 154 | 88.1 ms | 273.7 ms | 141 MiB |
| litestar + uvicorn | 156 | 154 | 99.2 ms | 221.9 ms | 157 MiB |
| litestar + granian | 163 | 163 | 71.0 ms | 277.6 ms | 130 MiB |
| robyn | 204 | 199 | 143.0 ms | 153.5 ms | 60 MiB |

## 5 · Carga de CPU pura: /suma, a tope

`hey -c 16` desde otro contenedor; servidor con 2 CPU y 1g. «En frío»: los primeros 10 s tras arrancar. «Templado»: los 20 s siguientes.

| variante | peticiones/s en frío | peticiones/s templado | p50 | p99 | memoria máx. bajo carga |
|---|---|---|---|---|---|
| go (net/http) | 15013 | 15019 | 1.0 ms | 2.6 ms | 10 MiB |
| fastapi + uvicorn | 232 | 232 | 58.4 ms | 184.7 ms | 101 MiB |
| fastapi + granian | 231 | 231 | 65.9 ms | 145.1 ms | 84 MiB |
| litestar + uvicorn | 222 | 221 | 69.3 ms | 153.8 ms | 92 MiB |
| litestar + granian | 234 | 234 | 65.2 ms | 147.5 ms | 74 MiB |
| robyn | 242 | 243 | 59.2 ms | 77.8 ms | 37 MiB |

## 6 · El coste del servidor: /suma?n=1 (casi sin cálculo), a tope

`hey -c 16` desde otro contenedor; servidor con 2 CPU y 1g. «En frío»: los primeros 10 s tras arrancar. «Templado»: los 20 s siguientes.

| variante | peticiones/s en frío | peticiones/s templado | p50 | p99 | memoria máx. bajo carga |
|---|---|---|---|---|---|
| go (net/http) | 91843 | 92537 | 0.1 ms | 0.4 ms | 12 MiB |
| fastapi + uvicorn | 14271 | 15867 | 0.9 ms | 2.1 ms | 101 MiB |
| fastapi + granian | 13726 | 13956 | 0.9 ms | 13.9 ms | 84 MiB |
| litestar + uvicorn | 17584 | 17521 | 0.8 ms | 2.2 ms | 92 MiB |
| litestar + granian | 16183 | 16335 | 0.8 ms | 5.7 ms | 75 MiB |
| robyn | 40924 | 40562 | 0.4 ms | 0.6 ms | 37 MiB |

## 7 · Latencia de una petición sola: /trabajo

Un solo cliente a 5 peticiones/s (`hey -c 1 -q 5`), 20 s, tras 10 s de calentamiento a tope. Cada petición llega sola: no hay cola, es lo que tarda el servidor en atender una.

| variante | peticiones/s logradas | p50 | p99 |
|---|---|---|---|
| go (net/http) | 5 | 9.60 ms | 12.00 ms |
| fastapi + uvicorn | 5 | 15.70 ms | 28.70 ms |
| fastapi + granian | 5 | 16.30 ms | 28.20 ms |
| litestar + uvicorn | 5 | 15.40 ms | 25.50 ms |
| litestar + granian | 5 | 16.80 ms | 27.00 ms |
| robyn | 5 | 15.10 ms | 21.50 ms |

## 8 · Latencia de una petición sola: /suma?n=1 (el servidor solo)

Un solo cliente a 5 peticiones/s (`hey -c 1 -q 5`), 20 s, tras 10 s de calentamiento a tope. Cada petición llega sola: no hay cola, es lo que tarda el servidor en atender una.

| variante | peticiones/s logradas | p50 | p99 |
|---|---|---|---|
| go (net/http) | 5 | 1.30 ms | 5.00 ms |
| fastapi + uvicorn | 5 | 2.80 ms | 5.30 ms |
| fastapi + granian | 5 | 3.60 ms | 6.20 ms |
| litestar + uvicorn | 5 | 2.10 ms | 4.90 ms |
| litestar + granian | 5 | 2.70 ms | 7.70 ms |
| robyn | 5 | 1.60 ms | 5.10 ms |

## G · ¿Libres de GIL? UN proceso con 1 y con 2 hilos de trabajo, /suma (CPU pura) a tope

`hey -c 16` 20 s tras 10 s de calentamiento; servidor con 2 CPU y 1g. CPU medida en el cgroup cada 0,5 s (muestreo.py): 100 % = un núcleo entero, el techo con 2 CPU es 200 %. Series en series/gil-*.json.

| variante | peticiones/s | p50 | CPU media | CPU máx. | hilos (máx.) |
|---|---|---|---|---|---|
| go, 1 núcleo (GOMAXPROCS=1) | 7577 | 2.1 ms | 100 % | 102 % | 3 |
| go, 2 núcleos (GOMAXPROCS=2) | 15051 | 1.0 ms | 199 % | 202 % | 4 |
| robyn, 1 proceso × 1 worker | 121 | 131.7 ms | 100 % | 102 % | 4 |
| robyn, 1 proceso × 2 workers | 121 | 131.6 ms | 100 % | 101 % | 5 |
| granian, 1 proceso (fastapi def) | 118 | 128.0 ms | 102 % | 103 % | 22 |
| robyn, 2 procesos × 1 worker | 123 | 129.8 ms | 100 % | 102 % | 9 |

