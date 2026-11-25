# SALIDA-extra — Scala Native: el cuelgue en release y el modo por defecto (golandia 865)

> Corrida del 2026-11-17 09:36, con el candado tomado. Servidores con 2 CPU y 1 GiB; hey -c 16 con 4 CPU.

## 1 · /suma a tope durante 30 s: release-full tal cual contra release-full con el arreglo

| binario | peticiones/s | respuestas 200 en 30 s | «Unhandled signal 11» en el log | ¿responde después? |
|---|---|---|---|---|
| release-full tal cual | 257.7 | 399 | 5 | sin respuesta en 5 s |
| release-full + SCALANATIVE_GC_TRAP_BASED_YIELDPOINTS=0 | 5746.9 | 172718 | 0 | 200 |

## 8 · Cómputo puro: el reloj del propio proceso, sin red

50 peticiones de calentamiento y 300 medidas, de una en una; el tiempo es el campo «ns» que mide cada servidor alrededor de su propio cálculo.

| variante | /suma?n=200000 mediana | p99 | /trabajo?n=20000 mediana | p99 |
|---|---|---|---|---|
| scala native release-full (arreglado), HILOS=2 | 0.149 ms | 0.223 ms | 18.908 ms | 21.341 ms |
| scala native modo por defecto (debug), HILOS=2 | 0.446 ms | 0.665 ms | 33.707 ms | 36.968 ms |

## Carga corta: /trabajo con hey -c 16 durante 10 s

- scala native release-full (arreglado), HILOS=2: 91.5 peticiones/s
- scala native modo por defecto (debug), HILOS=2: 8.1 peticiones/s

