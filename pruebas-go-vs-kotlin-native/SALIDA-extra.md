# SALIDA-extra — Kotlin/Native: debug contra release (golandia 879)

> Corrida del 2026-11-18 10:18, con el candado tomado. Servidores con 2 CPU y 1 GiB; hey -c 16 con 4 CPU.

## Estado de la máquina antes de medir

- /proc/loadavg (desde un contenedor alpine, VM de Docker con 24 hilos): `1.34 1.40 1.65 5/1136 6`
- contenedores ajenos en marcha (CPU en ese momento): jk-a-fijo 0.52%, jk-a-jenkins 0.57%, jk-b-jenkins-877 0.12%, jk-b-k8s-control-plane 6.37%

## 0 · ¿Hacen lo mismo? (antes de medir)

- go: total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go
- kotlin/native release, HILOS=2: total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go
- kotlin/native debug, HILOS=2: total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go

## 8 · Cómputo puro: el reloj del propio proceso, sin red

50 peticiones de calentamiento y 300 medidas, de una en una; el tiempo es el campo «ns» que mide cada servidor alrededor de su propio cálculo.

| variante | /suma?n=200000 mediana | p99 | /trabajo?n=20000 mediana | p99 |
|---|---|---|---|---|
| kotlin/native release, HILOS=2 | 0.105 ms | 0.140 ms | 9.957 ms | 11.474 ms |
| kotlin/native debug, HILOS=2 | 0.549 ms | 0.627 ms | 41.638 ms | 43.451 ms |

## Carga corta: hey -c 16 durante 10 s

| variante | /trabajo pet/s | /suma pet/s |
|---|---|---|
| kotlin/native release, HILOS=2 | 120.5 | 4399.4 |
| kotlin/native debug, HILOS=2 | 39.0 | 1372.6 |

