# SALIDA — Granian, Robyn y Litestar contra Go (golandia; pedido por @mdt4177 en el 770 de Cultura DevOps)

> Corrida del 2026-11-09 19:39. Regenerable con `python correr.py`. ⛔ Ninguna cifra de blog: todo salió de aquí. La misma lógica (app/logica.py) en Python y el mismo servidor Go de la serie «¿Go o X?».

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
| fastapi | 21 | 8.7 |
| litestar | 29 | 17.7 |
| robyn | 12 | 9.7 |
| go | 0 | 4.0 |

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
| go (net/http) | 217 ms | 0.0 ms | 3.0 MiB |
| fastapi + uvicorn | 512 ms | 108.0 ms | 95.4 MiB |
| fastapi + granian | 443 ms | 119.1 ms | 78.2 MiB |
| litestar + uvicorn | 490 ms | 98.2 ms | 87.8 MiB |
| litestar + granian | 432 ms | 110.9 ms | 69.8 MiB |
| robyn | 322 ms | 114.1 ms | 34.8 MiB |

## A · Cómputo de UNA petición (mediana del campo ns, 30 de calentamiento + 200 medidas, de una en una)

| variante | /trabajo (20.000 objetos) | /suma (200.000 vueltas) |
|---|---|---|
| go (net/http) | 3.23 ms | 0.122 ms |
| fastapi + uvicorn | 8.64 ms | 8.359 ms |
| fastapi + granian | 8.64 ms | 8.205 ms |
| litestar + uvicorn | 8.63 ms | 8.332 ms |
| litestar + granian | 8.89 ms | 8.210 ms |
| robyn | 8.51 ms | 8.292 ms |

## 4 · Carga con memoria: /trabajo (objetos, dict, orden, JSON), a tope

`hey -c 16` desde otro contenedor; servidor con 2 CPU y 1g. «En frío»: los primeros 10 s tras arrancar. «Templado»: los 20 s siguientes.

| variante | peticiones/s en frío | peticiones/s templado | p50 | p99 | memoria máx. bajo carga |
|---|---|---|---|---|---|
| go (net/http) | 477 | 476 | 26.4 ms | 123.2 ms | 33 MiB |
| fastapi + uvicorn | 153 | 153 | 99.5 ms | 210.5 ms | 164 MiB |
| fastapi + granian | 155 | 154 | 97.6 ms | 223.4 ms | 142 MiB |
| litestar + uvicorn | 155 | 153 | 92.5 ms | 273.8 ms | 168 MiB |
| litestar + granian | 161 | 160 | 85.5 ms | 260.8 ms | 127 MiB |
| robyn | 206 | 206 | 38.0 ms | 132.2 ms | 61 MiB |

## 5 · Carga de CPU pura: /suma, a tope

`hey -c 16` desde otro contenedor; servidor con 2 CPU y 1g. «En frío»: los primeros 10 s tras arrancar. «Templado»: los 20 s siguientes.

| variante | peticiones/s en frío | peticiones/s templado | p50 | p99 | memoria máx. bajo carga |
|---|---|---|---|---|---|
| go (net/http) | 14980 | 14990 | 1.0 ms | 2.6 ms | 11 MiB |
| fastapi + uvicorn | 231 | 232 | 65.8 ms | 150.6 ms | 101 MiB |
| fastapi + granian | 232 | 232 | 65.1 ms | 144.2 ms | 84 MiB |
| litestar + uvicorn | 233 | 232 | 50.3 ms | 209.8 ms | 92 MiB |
| litestar + granian | 232 | 232 | 66.0 ms | 144.9 ms | 74 MiB |
| robyn | 241 | 238 | 121.4 ms | 127.1 ms | 37 MiB |

## 6 · El coste del servidor: /suma?n=1 (casi sin cálculo), a tope

`hey -c 16` desde otro contenedor; servidor con 2 CPU y 1g. «En frío»: los primeros 10 s tras arrancar. «Templado»: los 20 s siguientes.

| variante | peticiones/s en frío | peticiones/s templado | p50 | p99 | memoria máx. bajo carga |
|---|---|---|---|---|---|
| go (net/http) | 90843 | 91005 | 0.1 ms | 0.4 ms | 12 MiB |
| fastapi + uvicorn | 14389 | 15982 | 0.9 ms | 2.1 ms | 101 MiB |
| fastapi + granian | 13987 | 13727 | 0.9 ms | 13.9 ms | 85 MiB |
| litestar + uvicorn | 17369 | 17512 | 0.8 ms | 2.4 ms | 92 MiB |
| litestar + granian | 15817 | 16233 | 0.8 ms | 4.4 ms | 75 MiB |
| robyn | 77084 | 45963 | 0.4 ms | 0.7 ms | 37 MiB |

## 7 · Latencia de una petición sola: /trabajo

Un solo cliente a 5 peticiones/s (`hey -c 1 -q 5`), 20 s, tras 10 s de calentamiento a tope. Cada petición llega sola: no hay cola, es lo que tarda el servidor en atender una.

| variante | peticiones/s logradas | p50 | p99 |
|---|---|---|---|
| go (net/http) | 5 | 4.10 ms | 6.00 ms |
| fastapi + uvicorn | 5 | 11.30 ms | 45.50 ms |
| fastapi + granian | 5 | 12.10 ms | 31.60 ms |
| litestar + uvicorn | 5 | 11.80 ms | 26.80 ms |
| litestar + granian | 5 | 12.30 ms | 24.70 ms |
| robyn | 5 | 11.60 ms | 19.20 ms |

## 8 · Latencia de una petición sola: /suma?n=1 (el servidor solo)

Un solo cliente a 5 peticiones/s (`hey -c 1 -q 5`), 20 s, tras 10 s de calentamiento a tope. Cada petición llega sola: no hay cola, es lo que tarda el servidor en atender una.

| variante | peticiones/s logradas | p50 | p99 |
|---|---|---|---|
| go (net/http) | 5 | 0.40 ms | 1.20 ms |
| fastapi + uvicorn | 5 | 1.00 ms | 1.80 ms |
| fastapi + granian | 5 | 1.10 ms | 2.50 ms |
| litestar + uvicorn | 5 | 0.90 ms | 1.60 ms |
| litestar + granian | 5 | 1.00 ms | 1.80 ms |
| robyn | 5 | 0.50 ms | 1.80 ms |

## G · ¿Libres de GIL? UN proceso con 1 y con 2 hilos de trabajo, /suma (CPU pura) a tope

`hey -c 16` 20 s tras 10 s de calentamiento; servidor con 2 CPU y 1g. CPU medida en el cgroup cada 0,5 s (muestreo.py): 100 % = un núcleo entero, el techo con 2 CPU es 200 %. Series en series/gil-*.json.

| variante | peticiones/s | p50 | CPU media | CPU máx. | hilos (máx.) |
|---|---|---|---|---|---|
| go, 1 núcleo (GOMAXPROCS=1) | 7047 | 2.3 ms | 100 % | 102 % | 3 |
| go, 2 núcleos (GOMAXPROCS=2) | 13806 | 1.1 ms | 200 % | 203 % | 4 |
| robyn, 1 proceso × 1 worker | 117 | 136.1 ms | 100 % | 102 % | 4 |
| robyn, 1 proceso × 2 workers | 116 | 138.5 ms | 101 % | 103 % | 5 |
| granian, 1 proceso (fastapi def) | 114 | 132.7 ms | 103 % | 105 % | 22 |
| robyn, 2 procesos × 1 worker | 235 | 67.9 ms | 200 % | 203 % | 9 |

## R · Robyn: ¿empezar la carga con la primera respuesta? (10 arranques de cada forma, /suma, `hey -c 16` 4 s)

Contenedor nuevo cada vez. «Primera 200»: la carga arranca en cuanto responde una petición. «Dos LISTO»: espera a que los dos procesos escriban su marca. «Un proceso» = la corrida quedó por debajo de 180 pet/s.

| cuándo empieza la carga | pet/s de cada arranque | con un proceso |
|---|---|---|
| primera 200 | 234, 233, 234, 232, 233, 229, 229, 232, 233, 232 | 0 de 10 |
| dos LISTO | 231, 115, 221, 230, 231, 228, 230, 230, 227, 230 | 1 de 10 |

## B · Reparto entre los 2 procesos: 30 arranques de cada uno (/suma, `hey -c 16` 6 s)
Contenedor nuevo cada vez; la carga empieza cuando los DOS procesos escribieron LISTO_EN_MS. A los 4 s se lee el %CPU de cada proceso (`docker top`). «Medio servidor»: el arranque quedó por debajo del 75 % de la mediana de sus propias corridas.
| variante | mediana pet/s | arranques a medio servidor | pet/s de esos arranques | %CPU de los 2 procesos que más usan, en uno de ellos |
|---|---|---|---|---|
| fastapi + uvicorn | 227 | 0 de 30 | — | 94 % · 93 % · 2 % |
| fastapi + granian | 220 | 0 de 30 | — | 93 % · 92 % · 3 % |
| litestar + uvicorn | 226 | 0 de 30 | — | 93 % · 93 % · 2 % |
| litestar + granian | 224 | 0 de 30 | — | 91 % · 90 % · 4 % |
| robyn | 228 | 1 de 30 | 117 | 94 % · 3 % · 0 % |

(Las tres primeras filas, del 2026-11-09 antes de un apagón; las demás, de la reanudación con --b-solo, en la misma máquina y con el mismo método.)

