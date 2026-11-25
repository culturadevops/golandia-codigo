# SALIDA — el caso Discord, reproducido (golandia)

> Análisis del 2026-10-19 17:21 . Regenerable con `python investigacion/pruebas_gc_discord.py` (o `--resumen` desde _crudo/). Servidor con 2 CPU y 4 GiB. 3.000.000 usuarios en caché; cliente con 16 trabajadores a 2000 peticiones/s en total durante 300 s (ritmo fijo: servicio sin saturar. como el de Discord).

## Go 1.9.2 · caché con punteros

`LISTO go1.9.2 · modo=punteros · usuarios=3000000 · montón vivo=722 MiB · llenar la caché tardó 2.8 s`

GOMAXPROCS que usó el proceso (gctrace): 24 P

Recolecciones durante la carga (gctrace: inicio · pausa + marcado concurrente + pausa · montón):

```
@  122.8s  pausa 0.075 ms + marcado 1409 ms + pausa 1.2 ms  · 1153->1158->727 MB 
@  244.2s  pausa 0.092 ms + marcado 1395 ms + pausa 0.91 ms  · 1160->1165->727 MB 
```

- Segundos medidos: 300 · peticiones: 599.999 (2.000 por segundo)
- p99 típico (mediana de los p99 por segundo): 0.43 ms · peor p99 de un segundo: 91.72 ms
- Peor latencia típica de un segundo (mediana de los máx.): 0.69 ms · la peor de todas: 102.73 ms
- Segundos con alguna petición de más de 10 ms: 7 · de ellos, dentro de una recolección (±1,5 s): 3
- Recolecciones: 2 (forzadas por los 2 minutos: 0) · una cada 121.4 s · marcado típico 1402 ms

## Go 1.9.2 · caché con punteros · GOMAXPROCS=2

`LISTO go1.9.2 · modo=punteros · usuarios=3000000 · montón vivo=722 MiB · llenar la caché tardó 3.9 s`

GOMAXPROCS que usó el proceso (gctrace): 2 P

Recolecciones durante la carga (gctrace: inicio · pausa + marcado concurrente + pausa · montón):

```
@  123.9s  pausa 0.21 ms + marcado 1092 ms + pausa 0.062 ms  · 1152->1156->726 MB 
@  245.0s  pausa 0.014 ms + marcado 888 ms + pausa 0.015 ms  · 1157->1161->725 MB 
```

- Segundos medidos: 300 · peticiones: 599.999 (2.000 por segundo)
- p99 típico (mediana de los p99 por segundo): 0.44 ms · peor p99 de un segundo: 2.02 ms
- Peor latencia típica de un segundo (mediana de los máx.): 0.72 ms · la peor de todas: 12.25 ms
- Segundos con alguna petición de más de 10 ms: 1 · de ellos, dentro de una recolección (±1,5 s): 1
- Recolecciones: 2 (forzadas por los 2 minutos: 0) · una cada 121.1 s · marcado típico 990 ms

## Go 1.26 · caché con punteros

`LISTO go1.26.9 · modo=punteros · usuarios=3000000 · montón vivo=713 MiB · llenar la caché tardó 1.9 s`

GOMAXPROCS que usó el proceso (gctrace): 2 P

Recolecciones durante la carga (gctrace: inicio · pausa + marcado concurrente + pausa · montón):

```
@  121.9s  pausa 0.023 ms + marcado 865 ms + pausa 0.003 ms  · 1162->1166->716 MB 
@  242.7s  pausa 0.025 ms + marcado 976 ms + pausa 0.002 ms  · 1167->1171->716 MB 
```

- Segundos medidos: 300 · peticiones: 600.000 (2.000 por segundo)
- p99 típico (mediana de los p99 por segundo): 0.42 ms · peor p99 de un segundo: 11.40 ms
- Peor latencia típica de un segundo (mediana de los máx.): 0.72 ms · la peor de todas: 20.58 ms
- Segundos con alguna petición de más de 10 ms: 3 · de ellos, dentro de una recolección (±1,5 s): 2
- Recolecciones: 2 (forzadas por los 2 minutos: 0) · una cada 120.9 s · marcado típico 920 ms

## Go 1.26 · caché sin punteros

`LISTO go1.26.9 · modo=plano · usuarios=3000000 · montón vivo=484 MiB · llenar la caché tardó 1.0 s`

GOMAXPROCS que usó el proceso (gctrace): 2 P

Recolecciones durante la carga (gctrace: inicio · pausa + marcado concurrente + pausa · montón):

```
@   91.8s  pausa 0.046 ms + marcado 186 ms + pausa 0.002 ms  · 824->825->485 MB 
@  182.8s  pausa 0.080 ms + marcado 178 ms + pausa 0.002 ms  · 826->827->485 MB 
@  273.7s  pausa 0.021 ms + marcado 198 ms + pausa 0.002 ms  · 826->827->485 MB 
```

- Segundos medidos: 300 · peticiones: 599.998 (2.000 por segundo)
- p99 típico (mediana de los p99 por segundo): 0.42 ms · peor p99 de un segundo: 7.13 ms
- Peor latencia típica de un segundo (mediana de los máx.): 0.68 ms · la peor de todas: 12.00 ms
- Segundos con alguna petición de más de 10 ms: 2 · de ellos, dentro de una recolección (±1,5 s): 2
- Recolecciones: 3 (forzadas por los 2 minutos: 0) · una cada 91.0 s · marcado típico 186 ms

## Resumen

| variante | pet/s | p99 típico | peor p99 (1 s) | peor petición | segundos con >10 ms (en recolección) | recolecciones · marcado típico |
|---|---|---|---|---|---|---|
| Go 1.9.2 · caché con punteros | 2.000 | 0.43 ms | 91.72 ms | 102.73 ms | 7 (3) | 2 · 1402 ms |
| Go 1.9.2 · caché con punteros · GOMAXPROCS=2 | 2.000 | 0.44 ms | 2.02 ms | 12.25 ms | 1 (1) | 2 · 990 ms |
| Go 1.26 · caché con punteros | 2.000 | 0.42 ms | 11.40 ms | 20.58 ms | 3 (2) | 2 · 920 ms |
| Go 1.26 · caché sin punteros | 2.000 | 0.42 ms | 7.13 ms | 12.00 ms | 2 (2) | 3 · 186 ms |
