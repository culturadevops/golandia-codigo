# SALIDA — ¿Go o Pyronaut? (golandia; pedido en un comentario del 848)

> Corrida del 2026-11-23 11:45. Regenerable con `python correr.py`. ⛔ Ninguna cifra de blog: todo salió de aquí. El banco del 848 (misma lógica, mismo servidor Go, mismas condiciones) con Pyronaut añadido.

- Pyronaut (`pyronaut --version` en la etapa de construcción):

```
Pyronaut: 0.1.2
Micronaut Core: 5.2.22
Micronaut Platform: 5.2.5
GraalPy: 25.4.4.1.1
GraalPy Interpreter: graalpy3.13-25.4.4
Native Image JDK: 25
```
- JVM de la imagen: `java version "25.0.4.1.1" 2026-08-18 LTS`
- Resto, sin tocar desde el 848: `golandia-govs-go:1` (go1.27.2, net/http, FROM scratch), `golandia-rust-py-fastapi:1` (Python 3.13.17, fastapi 0.162.0 + uvicorn 0.54.0, 2 procesos), `golandia-rust-py-robyn:1` (robyn 0.88.0, 2 procesos).
- Servidores con 2 CPU y 1g. Pyronaut: 1 proceso (la JVM), rutas `def` → ejecutor «io» de Micronaut, pool de contextos de GraalPy por defecto. `hey -c 16` con 4 CPU desde otro contenedor.

## 1 · Construir la imagen de Pyronaut desde cero (`docker build --no-cache`)

150 s. Incluye `pip install pyronaut` y `pyronaut setup`, que DESCARGA el GraalVM JDK, GraalPy y las dependencias de Maven (en el 848 las imágenes de Python tardaron 8,7-17,7 s y la de Go 4,0 s, con la imagen base ya descargada).

## 2 · Lo que despliegas

| imagen | tamaño (docker image inspect .Size, el mismo dato que el 848) |
|---|---|
| pyronaut (JVM) | 491.5 MiB |
| fastapi | 61.1 MiB |
| robyn | 54.1 MiB |
| go | 2.4 MiB |

Por dentro (`ls -l /app/banco.jar; du -sh /opt/graalvm`):

```
-rw------- 1 root root 97560169 Nov 23 16:48 /app/banco.jar
901M	/opt/graalvm
```

## 0 · ¿Hacen lo mismo? (antes de medir)

El cuerpo de /trabajo?n=20000, /suma?n=200000 y /suma?n=1, comparado byte a byte con go (net/http) quitando solo "lang" (go o python) y "ns" (el reloj del cálculo).

- go (net/http): total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591, 2220 + 31 + 16 bytes, application/json → idéntico
- fastapi + uvicorn: total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591, 2220 + 31 + 16 bytes, application/json → idéntico
- robyn: total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591, 2220 + 31 + 16 bytes, application/json → idéntico
- pyronaut (JVM): total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591, 2220 + 31 + 16 bytes, application/json → idéntico

## 3 · Arranque: contenedor nuevo → primera respuesta 200 (mediana de 8) y memoria en reposo

| variante | 1ª respuesta (docker incluido) | «listo» (su propio reloj) | memoria en reposo |
|---|---|---|---|
| go (net/http) | 203 ms | 0.0 ms | 2.8 MiB |
| fastapi + uvicorn | 493 ms | 108.3 ms | 95.7 MiB |
| robyn | 318 ms | 120.0 ms | 35.0 MiB |
| pyronaut (JVM) | 8613 ms | 53.8 ms | 597.5 MiB |

## A · Cómputo de UNA petición (mediana del campo ns, 30 de calentamiento + 200 medidas, de una en una)

| variante | /trabajo (20.000 objetos) | /suma (200.000 vueltas) |
|---|---|---|
| go (net/http) | 3.19 ms | 0.122 ms |
| fastapi + uvicorn | 8.47 ms | 8.202 ms |
| robyn | 8.66 ms | 8.320 ms |
| pyronaut (JVM) | 6.38 ms | 0.139 ms |

## J · El calentamiento: lo que tarda el cálculo, petición a petición, desde la primera (mediana del campo ns por tramo)

Contenedor nuevo; peticiones de una en una. El campo ns es el reloj del propio cálculo, sin red ni framework.

**/suma?n=200000** (ms)

| variante | 1ª | 2-10 | 11-50 | 51-200 | 201-500 |
|---|---|---|---|---|---|
| go (net/http) | 0.14 | 0.12 | 0.12 | 0.12 | 0.12 |
| fastapi + uvicorn | 8.18 | 8.11 | 8.08 | 8.21 | 8.14 |
| robyn | 8.22 | 8.28 | 8.20 | 8.28 | 8.26 |
| pyronaut (JVM) | 271.37 | 99.37 | 7.43 | 0.64 | 0.63 |

**/trabajo?n=20000** (ms)

| variante | 1ª | 2-10 | 11-50 | 51-200 | 201-500 |
|---|---|---|---|---|---|
| go (net/http) | 3.27 | 3.36 | 3.24 | 3.25 | 3.22 |
| fastapi + uvicorn | 9.47 | 9.34 | 8.57 | 8.45 | 8.45 |
| robyn | 13.21 | 8.27 | 8.44 | 8.63 | 8.59 |
| pyronaut (JVM) | 1454.95 | 65.97 | 10.69 | 5.83 | 5.57 |

## 4 · Carga con memoria: /trabajo (objetos, dict, orden, JSON), a tope

`hey -c 16` desde otro contenedor; servidor con 2 CPU y 1g. «En frío»: los primeros 10 s tras arrancar. «Templado»: los 20 s siguientes.

| variante | peticiones/s en frío | peticiones/s templado | p50 | p99 | memoria máx. bajo carga |
|---|---|---|---|---|---|
| go (net/http) | 465 | 479 | 26.4 ms | 123.3 ms | 32 MiB |
| fastapi + uvicorn | 156 | 155 | 93.6 ms | 256.8 ms | 168 MiB |
| robyn | 204 | 104 | 152.8 ms | 161.8 ms | 59 MiB |
| pyronaut (JVM) | 28 | 81 | 208.0 ms | 671.7 ms | 639 MiB |

## 5 · Carga de CPU pura: /suma, a tope

`hey -c 16` desde otro contenedor; servidor con 2 CPU y 1g. «En frío»: los primeros 10 s tras arrancar. «Templado»: los 20 s siguientes.

| variante | peticiones/s en frío | peticiones/s templado | p50 | p99 | memoria máx. bajo carga |
|---|---|---|---|---|---|
| go (net/http) | 15041 | 14986 | 1.0 ms | 2.6 ms | 11 MiB |
| fastapi + uvicorn | 233 | 233 | 63.7 ms | 163.7 ms | 101 MiB |
| robyn | 122 | 237 | 51.5 ms | 90.4 ms | 38 MiB |
| pyronaut (JVM) | 2216 | 8133 | 1.0 ms | 29.8 ms | 584 MiB |

## 6 · El coste del servidor: /suma?n=1 (casi sin cálculo), a tope

`hey -c 16` desde otro contenedor; servidor con 2 CPU y 1g. «En frío»: los primeros 10 s tras arrancar. «Templado»: los 20 s siguientes.

| variante | peticiones/s en frío | peticiones/s templado | p50 | p99 | memoria máx. bajo carga |
|---|---|---|---|---|---|
| go (net/http) | 93218 | 93223 | 0.1 ms | 0.4 ms | 12 MiB |
| fastapi + uvicorn | 14368 | 16095 | 0.9 ms | 2.1 ms | 101 MiB |
| robyn | 80531 | 49584 | 0.4 ms | 0.5 ms | 38 MiB |
| pyronaut (JVM) | 6120 | 39610 | 0.2 ms | 1.2 ms | 603 MiB |

## L · Carga sostenida: /trabajo durante 60 s, en tramos de 10 s

Contenedor nuevo; `hey -c 16` 6 veces 10 s seguidas, sin pausa. Peticiones/s de cada tramo, p99 del último y memoria máxima.

| variante | 0-10 s | 10-20 s | 20-30 s | 30-40 s | 40-50 s | 50-60 s | p99 último | memoria máx. |
|---|---|---|---|---|---|---|---|---|
| go (net/http) | 480 | 479 | 479 | 481 | 481 | 481 | 119.3 ms | 32 MiB |
| fastapi + uvicorn | 155 | 156 | 151 | 152 | 152 | 153 | 212.9 ms | 173 MiB |
| robyn | 201 | 201 | 195 | 197 | 197 | 202 | 144.8 ms | 61 MiB |
| pyronaut (JVM) | 18 | 80 | 74 | 83 | 85 | 81 | 741.3 ms | 701 MiB |

## 7 · Latencia de una petición sola: /trabajo

Un solo cliente a 5 peticiones/s (`hey -c 1 -q 5`), 20 s, tras 10 s de calentamiento a tope. Cada petición llega sola: no hay cola, es lo que tarda el servidor en atender una.

| variante | peticiones/s logradas | p50 | p99 |
|---|---|---|---|
| go (net/http) | 5 | 6.70 ms | 10.70 ms |
| fastapi + uvicorn | 5 | 13.50 ms | 23.50 ms |
| robyn | 5 | 12.30 ms | 18.60 ms |
| pyronaut (JVM) | 5 | 9.00 ms | 160.70 ms |

## 8 · Latencia de una petición sola: /suma?n=1 (el servidor solo)

Un solo cliente a 5 peticiones/s (`hey -c 1 -q 5`), 20 s, tras 10 s de calentamiento a tope. Cada petición llega sola: no hay cola, es lo que tarda el servidor en atender una.

| variante | peticiones/s logradas | p50 | p99 |
|---|---|---|---|
| go (net/http) | 5 | 0.40 ms | 1.50 ms |
| fastapi + uvicorn | 5 | 1.20 ms | 3.90 ms |
| robyn | 5 | 0.70 ms | 2.10 ms |
| pyronaut (JVM) | 5 | 1.10 ms | 2.50 ms |

## G · ¿Libre de GIL? UN proceso, /suma (CPU pura) a tope

`hey -c 16` 20 s tras 30 s de calentamiento (el JIT necesita más que los 10 s del 848); servidor con 2 CPU y 1g. CPU del cgroup cada 0,5 s (muestreo.py): 100 % = un núcleo, el techo es 200 %. «Contextos»: cuántas veces se importó la app (una por contexto de GraalPy). Series en series/gil-*.json.

| variante | peticiones/s | p50 | CPU media | CPU máx. | hilos (máx.) | contextos |
|---|---|---|---|---|---|---|
| go, 1 núcleo (GOMAXPROCS=1) | 7555 | 2.1 ms | 100 % | 102 % | 3 | — |
| go, 2 núcleos (GOMAXPROCS=2) | 14954 | 1.0 ms | 200 % | 203 % | 4 | — |
| fastapi + uvicorn, 1 proceso | 116 | 131.7 ms | 102 % | 104 % | 21 | — |
| pyronaut, sin pool (1 contexto) | 4987 | 3.2 ms | 113 % | 137 % | 41 | 1 |
| pyronaut, pool por defecto | 8825 | 0.9 ms | 200 % | 207 % | 53 | 2 |

