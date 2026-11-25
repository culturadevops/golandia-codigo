# SALIDA — ¿Go o PHP? (golandia), con el banco de la revancha

> Corrida del 2026-10-24 09:29. Regenerable con `python investigacion/pruebas_go_vs_x.py php`. ⛔ Ninguna cifra de blog: todo salió de aquí.

- go: `go version go1.27.2 linux/amd64` (net/http, solo librería estándar)
- php: `FrankenPHP v1.13.3 PHP 8.5.12 Caddy v2.11.7 h1:yj0Y4fYZGPkSvibBJ1sTWE33xC0fxztVyXEW5iIdUT4=` — dunglas/frankenphp:1.13.3-php8.5
- Docker `29.6.2`. Servidores con 2 CPU y 1g; `hey -c 16` con 4 CPU desde otro contenedor.

## 1 · Construir la imagen desde cero (`docker build --no-cache`, dependencias incluidas)

| lenguaje | qué se cronometra | segundos |
|---|---|---|
| go | go build | 5.7 |
| php | copiar el código (PHP no compila) | 1.4 |

## 2 · Lo que despliegas

| imagen | base | tamaño |
|---|---|---|
| go | scratch + binario estático | 2.4 MiB |
| php | dunglas/frankenphp:1.13.3-php8.5 | 201.2 MiB |

## 0 · ¿Hacen lo mismo? (antes de medir)

- go: total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go
- php clásico (FrankenPHP): total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go
- php worker (FrankenPHP): total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go
- php worker + JIT: total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go

## 3 · Arranque: contenedor nuevo → primera respuesta 200 (mediana de 8) y memoria en reposo

| variante | 1ª respuesta (docker incluido) | «listo» (su propio reloj) | memoria en reposo |
|---|---|---|---|
| go | 211 ms | 0.0 ms | 3.1 MiB |
| php clásico (FrankenPHP) | 269 ms | — | 24.9 MiB |
| php worker (FrankenPHP) | 256 ms | 0.0 ms | 25.9 MiB |
| php worker + JIT | 261 ms | 0.0 ms | 25.7 MiB |

## 4 · Carga con memoria: /trabajo (objetos, mapa, orden, JSON), a tope

`hey -c 16` desde otro contenedor; servidor con 2 CPU y 1g. «En frío»: los primeros 10 s tras arrancar. «Templado»: los 20 s siguientes.

| variante | peticiones/s en frío | peticiones/s templado | p50 | p99 | memoria máx. bajo carga |
|---|---|---|---|---|---|
| go | 456 | 453 | 28.8 ms | 119.5 ms | 32 MiB |
| php clásico (FrankenPHP) | 153 | 152 | 102.5 ms | 158.4 ms | 63 MiB |
| php worker (FrankenPHP) | 132 | 131 | 111.1 ms | 170.7 ms | 60 MiB |
| php worker + JIT | 153 | 147 | 104.1 ms | 166.6 ms | 64 MiB |

## 5 · Carga de CPU pura: /suma, a tope

`hey -c 16` desde otro contenedor; servidor con 2 CPU y 1g. «En frío»: los primeros 10 s tras arrancar. «Templado»: los 20 s siguientes.

| variante | peticiones/s en frío | peticiones/s templado | p50 | p99 | memoria máx. bajo carga |
|---|---|---|---|---|---|
| go | 14024 | 14142 | 1.0 ms | 2.9 ms | 11 MiB |
| php clásico (FrankenPHP) | 2444 | 2499 | 6.5 ms | 11.5 ms | 33 MiB |
| php worker (FrankenPHP) | 2592 | 2589 | 6.2 ms | 11.2 ms | 34 MiB |
| php worker + JIT | 4036 | 4019 | 4.0 ms | 7.3 ms | 35 MiB |

## 6 · Latencia de una petición sola: /trabajo

Un solo cliente a 5 peticiones/s (`hey -c 1 -q 5`), 20 s, tras 10 s de calentamiento a tope. Cada petición llega sola: no hay cola, es lo que tarda el servidor en atender una.

| variante | peticiones/s logradas | p50 | p99 |
|---|---|---|---|
| go | 5 | 4.00 ms | 4.70 ms |
| php clásico (FrankenPHP) | 5 | 13.50 ms | 14.60 ms |
| php worker (FrankenPHP) | 5 | 15.80 ms | 17.60 ms |
| php worker + JIT | 5 | 14.20 ms | 16.10 ms |

## 7 · Latencia de una petición sola: /suma

Un solo cliente a 5 peticiones/s (`hey -c 1 -q 5`), 20 s, tras 10 s de calentamiento a tope. Cada petición llega sola: no hay cola, es lo que tarda el servidor en atender una.

| variante | peticiones/s logradas | p50 | p99 |
|---|---|---|---|
| go | 5 | 0.50 ms | 1.50 ms |
| php clásico (FrankenPHP) | 5 | 1.50 ms | 2.40 ms |
| php worker (FrankenPHP) | 5 | 1.40 ms | 2.70 ms |
| php worker + JIT | 5 | 1.10 ms | 1.90 ms |

