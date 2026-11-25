# SALIDA-comprobar — ¿de verdad hacen el mismo trabajo? (golandia 880)

> Corrida del 2026-11-18 11:44, con el candado tomado. Servidores con 2 CPU y 1 GiB.

## 1 · La respuesta entera de /trabajo (sin lang ni ns), contra Go

- go: n=1000: 2418 caracteres, igual que Go · n=20000: 2524 caracteres, igual que Go · n=50000: 2569 caracteres, igual que Go
- bun --compile: n=1000: 2418 caracteres, igual que Go · n=20000: 2524 caracteres, igual que Go · n=50000: 2569 caracteres, igual que Go
- bun --compile --aot: n=1000: 2418 caracteres, igual que Go · n=20000: 2524 caracteres, igual que Go · n=50000: 2569 caracteres, igual que Go
- scriptc: n=1000: 2418 caracteres, igual que Go · n=20000: 2524 caracteres, igual que Go · n=50000: 2569 caracteres, igual que Go

## 2 · CPU y memoria con /trabajo a tope (hey -c 16, 20 s; docker stats cada ~2 s)

| variante | pet/s | CPU máx. (200 % = los 2 núcleos) | CPU mediana | memoria al final |
|---|---|---|---|---|
| go | 446 | 198 % | 196 % | 19.34MiB |
| bun --compile | 262 | 112 % | 111 % | 16.27MiB |
| bun --compile --aot | 508 | 102 % | 101 % | 5.934MiB |
| scriptc | 550 | 100 % | 99 % | 3.031MiB |

## 3 · Memoria tras 3.000 peticiones seguidas de /trabajo (de una en una)

- go: en reposo 2.566MiB → tras 3.000 peticiones 8.797MiB
- bun --compile: en reposo 8.008MiB → tras 3.000 peticiones 31.45MiB
- bun --compile --aot: en reposo 4.672MiB → tras 3.000 peticiones 9.289MiB
- scriptc: en reposo 1.508MiB → tras 3.000 peticiones 5.984MiB

