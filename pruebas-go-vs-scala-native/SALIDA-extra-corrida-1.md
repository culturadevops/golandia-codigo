# SALIDA-extra — Scala Native: el cuelgue en release y el modo por defecto (golandia 865)

> Corrida del 2026-11-17 09:21, con el candado tomado. Servidores con 2 CPU y 1 GiB; hey -c 16 con 4 CPU.

## 1 · /suma a tope durante 30 s: release-full tal cual contra release-full con el arreglo

| binario | peticiones/s | respuestas 200 en 30 s | «Unhandled signal 11» en el log | ¿responde después? |
|---|---|---|---|---|
| release-full tal cual | 279.1 | 460 | 3 | sin respuesta en 5 s |
| release-full + SCALANATIVE_GC_TRAP_BASED_YIELDPOINTS=0 | 5364.6 | 161001 | 0 | 200 |

## 8 · Cómputo puro: el reloj del propio proceso, sin red

50 peticiones de calentamiento y 300 medidas, de una en una; el tiempo es el campo «ns» que mide cada servidor alrededor de su propio cálculo.

| variante | /suma?n=200000 mediana | p99 | /trabajo?n=20000 mediana | p99 |
|---|---|---|---|---|
| scala native release-full (arreglado), HILOS=2 | 0.149 ms | 0.203 ms | 18.897 ms | 20.789 ms |
| scala native modo por defecto (debug), HILOS=2 | 0.445 ms | 0.704 ms | 34.562 ms | 37.872 ms |

## Carga corta: /trabajo con hey -c 16 durante 10 s

- scala native release-full (arreglado), HILOS=2: 94.5 peticiones/s
- scala native modo por defecto (debug), HILOS=2: 7.4 peticiones/s

