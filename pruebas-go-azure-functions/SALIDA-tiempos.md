# SALIDA-tiempos — Go en Azure Functions con custom handler, en LOCAL

> Corrida del 2026-11-23 11:28, con el candado banco.lock y sin renders. 2 CPU y 1 GiB por contenedor. ⛔ Es un arranque EN LOCAL, con la imagen ya descargada: no es el arranque en frío de Azure.

## T1 · Contenedor nuevo → primera respuesta 200 de /api/hola (mediana de 8) y memoria en reposo

| variante | 1ª respuesta (docker incluido) | mín | máx | memoria del contenedor en reposo |
|---|---|---|---|---|
| host oficial + handler de Go | 1174 ms | 1146 ms | 1512 ms | 77.9 MiB |
| el handler solo, sin host | 234 ms | 203 ms | 241 ms | 3.0 MiB |

(«1ª respuesta» cuenta desde que se lanza `docker run` hasta el primer 200: incluye crear el contenedor.)

## T2 · Latencia de una petición sola, en caliente: un cliente a 5 pet/s durante 20 s (`hey -c 1 -q 5`)

| variante | p50 | p99 |
|---|---|---|
| host oficial + handler de Go | 2.60 ms | 5.90 ms |
| el handler solo, sin host | 0.70 ms | 2.80 ms |

