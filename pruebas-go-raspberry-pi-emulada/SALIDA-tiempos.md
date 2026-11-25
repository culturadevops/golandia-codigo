# SALIDA-tiempos — el mismo cálculo, nativo y emulado

> Corrida del 2026-11-23 11:17, con el candado banco.lock y sin renders. `sensor -calculo 300000000` (multiplicar y módulo en un bucle), 5 veces cada uno, mediana.

| binario | plataforma | mediana | mín | máx |
|---|---|---|---|---|
| sensor-amd64 | linux/amd64 | 0.17 s | 0.17 s | 0.17 s |
| sensor-arm64 | linux/arm64 | 0.46 s | 0.40 s | 0.46 s |
| sensor-armv7 | linux/arm/v7 | 0.45 s | 0.40 s | 0.46 s |

