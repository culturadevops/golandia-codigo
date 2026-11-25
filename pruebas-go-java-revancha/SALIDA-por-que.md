# SALIDA — ¿por qué Java le gana a Go en /trabajo? (anexo del vídeo 506)

> Corrida del 2026-10-19 16:05. Regenerable con `python investigacion/pruebas_go_java_revancha_go.py`. Mismo método que SALIDA.md.

## A · Cómputo puro de /trabajo: el reloj del propio servidor, una petición cada vez (sin concurrencia)

| variante | mediana | p99 |
|---|---|---|
| go (por defecto: GOGC=100) | 3.39 ms | 3.83 ms |
| go GOGC=400 | 2.89 ms | 3.85 ms |
| go GOMEMLIMIT=700MiB GOGC=off | 3.10 ms | 4.63 ms |
| go con slices.SortFunc | 2.99 ms | 3.35 ms |
| go slices.SortFunc + GOMEMLIMIT=700MiB GOGC=off | 2.82 ms | 3.66 ms |
| java-jre (virtuales) | 1.90 ms | 3.76 ms |

## B · Bajo carga: /trabajo con hey -c 16

`hey -c 16` desde otro contenedor; servidor con 2 CPU y 1g de memoria. «En frío»: los primeros 10 s tras arrancar. «Templado»: los 20 s siguientes.

| variante | peticiones/s en frío | peticiones/s templado | p50 | p99 | memoria máx. bajo carga |
|---|---|---|---|---|---|
| go (por defecto: GOGC=100) | 450 | 447 | 27.8 ms | 134.9 ms | 33 MiB |
| go GOGC=400 | 560 | 558 | 25.8 ms | 81.7 ms | 49 MiB |
| go GOMEMLIMIT=700MiB GOGC=off | 654 | 658 | 23.8 ms | 41.5 ms | 671 MiB |
| go con slices.SortFunc | 480 | 444 | 29.4 ms | 123.0 ms | 37 MiB |
| go slices.SortFunc + GOMEMLIMIT=700MiB GOGC=off | 696 | 699 | 22.5 ms | 38.5 ms | 672 MiB |
| java-jre (virtuales) | 792 | 949 | 16.7 ms | 21.6 ms | 295 MiB |

