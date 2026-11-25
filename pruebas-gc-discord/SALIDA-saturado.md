# SALIDA — el caso Discord, reproducido (golandia)

> Análisis del 2026-10-19 16:48 (medición del 2026-10-19 16:32-16:52). Regenerable con `python investigacion/pruebas_gc_discord.py` (o `--resumen` desde _crudo/). Servidor con 2 CPU y 4 GiB. 3.000.000 usuarios en caché; cliente con 16 trabajadores durante 300 s.

## Go 1.9.2 · caché con punteros

`LISTO go1.9.2 · modo=punteros · usuarios=3000000 · montón vivo=722 MiB · llenar la caché tardó 3.4 s`

Recolecciones durante la carga (gctrace: inicio · pausa + marcado concurrente + pausa · montón):

```
@   14.9s  pausa 0.26 ms + marcado 1275 ms + pausa 0.48 ms  · 1407->1416->731 MB 
@   24.6s  pausa 61 ms + marcado 1204 ms + pausa 0.44 ms  · 1427->1437->733 MB 
@   34.6s  pausa 0.16 ms + marcado 1171 ms + pausa 0.42 ms  · 1429->1442->735 MB 
@   44.6s  pausa 0.13 ms + marcado 1182 ms + pausa 0.25 ms  · 1434->1446->734 MB 
@   55.0s  pausa 0.14 ms + marcado 1205 ms + pausa 0.42 ms  · 1433->1444->734 MB 
@   65.3s  pausa 0.13 ms + marcado 1387 ms + pausa 0.46 ms  · 1432->1442->733 MB 
@   75.9s  pausa 0.11 ms + marcado 1476 ms + pausa 0.29 ms  · 1430->1443->735 MB 
@   86.4s  pausa 0.10 ms + marcado 1287 ms + pausa 0.37 ms  · 1434->1446->735 MB 
@   96.8s  pausa 0.13 ms + marcado 1286 ms + pausa 0.93 ms  · 1434->1446->735 MB 
@  107.6s  pausa 0.081 ms + marcado 1385 ms + pausa 0.75 ms  · 1433->1446->735 MB 
@  118.6s  pausa 0.13 ms + marcado 1381 ms + pausa 0.41 ms  · 1433->1445->734 MB 
@  129.7s  pausa 0.12 ms + marcado 1388 ms + pausa 0.16 ms  · 1433->1444->734 MB 
@  140.3s  pausa 0.23 ms + marcado 1392 ms + pausa 0.43 ms  · 1431->1444->735 MB 
@  150.7s  pausa 0.20 ms + marcado 1370 ms + pausa 0.45 ms  · 1434->1446->735 MB 
@  161.8s  pausa 0.11 ms + marcado 1387 ms + pausa 0.52 ms  · 1434->1446->735 MB 
@  172.6s  pausa 0.12 ms + marcado 1475 ms + pausa 0.53 ms  · 1434->1446->735 MB 
@  183.6s  pausa 0.17 ms + marcado 1385 ms + pausa 0.40 ms  · 1434->1446->735 MB 
@  194.5s  pausa 0.085 ms + marcado 1478 ms + pausa 0.53 ms  · 1433->1445->735 MB 
@  205.3s  pausa 0.11 ms + marcado 1588 ms + pausa 0.29 ms  · 1433->1446->735 MB 
@  216.1s  pausa 0.24 ms + marcado 1500 ms + pausa 0.63 ms  · 1434->1446->735 MB 
@  227.2s  pausa 0.10 ms + marcado 1565 ms + pausa 0.29 ms  · 1433->1446->735 MB 
@  238.1s  pausa 0.20 ms + marcado 1480 ms + pausa 0.82 ms  · 1434->1446->734 MB 
@  249.2s  pausa 0.14 ms + marcado 1572 ms + pausa 0.51 ms  · 1432->1445->735 MB 
@  259.9s  pausa 0.14 ms + marcado 1584 ms + pausa 0.19 ms  · 1434->1446->735 MB 
@  270.7s  pausa 0.13 ms + marcado 1373 ms + pausa 0.24 ms  · 1434->1446->735 MB 
@  281.4s  pausa 0.18 ms + marcado 1483 ms + pausa 0.42 ms  · 1433->1444->733 MB 
@  292.2s  pausa 0.096 ms + marcado 1972 ms + pausa 0.54 ms  · 1430->1443->735 MB 
@  303.5s  pausa 0.17 ms + marcado 1272 ms + pausa 0.27 ms  · 1434->1444->732 MB 
```

- Segundos medidos: 300 · peticiones: 11.032.636 (36.775 por segundo)
- p99 típico (mediana de los p99 por segundo): 0.81 ms · peor p99 de un segundo: 92.51 ms
- Peor latencia típica de un segundo (mediana de los máx.): 66.91 ms · la peor de todas: 104.28 ms
- Segundos con alguna petición de más de 10 ms: 300 · de ellos, dentro de una recolección (±1,5 s): 121
- Recolecciones: 28 (forzadas por los 2 minutos: 0) · una cada 10.7 s · marcado típico 1387 ms

## Go 1.26 · caché con punteros

`LISTO go1.26.9 · modo=punteros · usuarios=3000000 · montón vivo=713 MiB · llenar la caché tardó 2.0 s`

Recolecciones durante la carga (gctrace: inicio · pausa + marcado concurrente + pausa · montón):

```
@   10.6s  pausa 0.093 ms + marcado 1166 ms + pausa 0.40 ms  · 1347->1358->725 MB 
@   20.8s  pausa 0.19 ms + marcado 1074 ms + pausa 0.085 ms  · 1370->1383->726 MB 
@   31.3s  pausa 0.41 ms + marcado 1088 ms + pausa 0.018 ms  · 1373->1382->723 MB 
@   41.7s  pausa 0.31 ms + marcado 994 ms + pausa 0.25 ms  · 1366->1378->725 MB 
@   52.5s  pausa 0.14 ms + marcado 1173 ms + pausa 0.32 ms  · 1414->1430->730 MB 
@   63.4s  pausa 0.15 ms + marcado 965 ms + pausa 0.14 ms  · 1412->1424->725 MB 
@   74.0s  pausa 0.13 ms + marcado 1175 ms + pausa 0.43 ms  · 1403->1417->727 MB 
@   84.8s  pausa 0.31 ms + marcado 1080 ms + pausa 0.027 ms  · 1407->1419->725 MB 
@   95.5s  pausa 0.23 ms + marcado 979 ms + pausa 0.56 ms  · 1402->1415->726 MB 
@  106.2s  pausa 0.33 ms + marcado 974 ms + pausa 0.10 ms  · 1405->1415->723 MB 
@  117.0s  pausa 0.20 ms + marcado 990 ms + pausa 0.11 ms  · 1410->1422->725 MB 
@  127.8s  pausa 0.28 ms + marcado 1070 ms + pausa 0.38 ms  · 1414->1429->729 MB 
@  139.0s  pausa 0.20 ms + marcado 1092 ms + pausa 0.46 ms  · 1418->1433->728 MB 
@  150.1s  pausa 0.33 ms + marcado 1099 ms + pausa 0.42 ms  · 1418->1430->726 MB 
@  160.8s  pausa 0.18 ms + marcado 1062 ms + pausa 0.008 ms  · 1412->1426->727 MB 
@  171.8s  pausa 0.20 ms + marcado 1183 ms + pausa 0.41 ms  · 1416->1429->727 MB 
@  182.8s  pausa 63 ms + marcado 1003 ms + pausa 0.15 ms  · 1414->1428->727 MB 
@  194.0s  pausa 0.15 ms + marcado 1183 ms + pausa 0.82 ms  · 1415->1430->728 MB 
@  205.2s  pausa 0.36 ms + marcado 1096 ms + pausa 0.004 ms  · 1416->1429->727 MB 
@  215.8s  pausa 56 ms + marcado 800 ms + pausa 0.25 ms  · 1413->1425->725 MB 
@  226.3s  pausa 0.17 ms + marcado 989 ms + pausa 0.11 ms  · 1410->1423->726 MB 
@  237.1s  pausa 0.17 ms + marcado 874 ms + pausa 0.68 ms  · 1411->1426->728 MB 
@  247.8s  pausa 0.17 ms + marcado 1080 ms + pausa 0.33 ms  · 1415->1427->725 MB 
@  258.3s  pausa 0.50 ms + marcado 1005 ms + pausa 0.20 ms  · 1415->1427->725 MB 
@  268.6s  pausa 0.13 ms + marcado 990 ms + pausa 0.005 ms  · 1415->1426->724 MB 
@  279.1s  pausa 0.18 ms + marcado 1184 ms + pausa 0.32 ms  · 1413->1429->729 MB 
@  289.7s  pausa 0.18 ms + marcado 974 ms + pausa 0.20 ms  · 1416->1430->728 MB 
@  300.2s  pausa 0.27 ms + marcado 1189 ms + pausa 0.040 ms  · 1413->1426->727 MB 
```

- Segundos medidos: 300 · peticiones: 10.347.373 (34.491 por segundo)
- p99 típico (mediana de los p99 por segundo): 0.93 ms · peor p99 de un segundo: 92.66 ms
- Peor latencia típica de un segundo (mediana de los máx.): 65.65 ms · la peor de todas: 192.23 ms
- Segundos con alguna petición de más de 10 ms: 300 · de ellos, dentro de una recolección (±1,5 s): 116
- Recolecciones: 28 (forzadas por los 2 minutos: 0) · una cada 10.7 s · marcado típico 1072 ms

## Go 1.26 · caché sin punteros

`LISTO go1.26.9 · modo=plano · usuarios=3000000 · montón vivo=484 MiB · llenar la caché tardó 1.0 s`

Recolecciones durante la carga (gctrace: inicio · pausa + marcado concurrente + pausa · montón):

```
@    6.7s  pausa 0.23 ms + marcado 368 ms + pausa 0.018 ms  · 928->930->487 MB 
@   13.5s  pausa 0.17 ms + marcado 297 ms + pausa 0.12 ms  · 934->937->488 MB 
@   20.0s  pausa 0.20 ms + marcado 369 ms + pausa 0.18 ms  · 936->939->488 MB 
@   27.1s  pausa 0.12 ms + marcado 270 ms + pausa 0.34 ms  · 954->957->488 MB 
@   34.4s  pausa 0.43 ms + marcado 204 ms + pausa 0.38 ms  · 969->973->488 MB 
@   41.5s  pausa 0.10 ms + marcado 291 ms + pausa 0.23 ms  · 971->973->487 MB 
@   48.8s  pausa 0.33 ms + marcado 279 ms + pausa 0.42 ms  · 968->970->488 MB 
@   56.1s  pausa 0.16 ms + marcado 368 ms + pausa 0.35 ms  · 969->974->490 MB 
@   63.4s  pausa 0.27 ms + marcado 306 ms + pausa 1.3 ms  · 971->973->488 MB 
@   70.5s  pausa 0.30 ms + marcado 383 ms + pausa 0.20 ms  · 967->970->488 MB 
@   77.8s  pausa 0.17 ms + marcado 379 ms + pausa 0.004 ms  · 968->971->487 MB 
@   85.0s  pausa 41 ms + marcado 306 ms + pausa 0.40 ms  · 966->968->487 MB 
@   92.2s  pausa 0.14 ms + marcado 364 ms + pausa 0.57 ms  · 965->968->488 MB 
@   99.3s  pausa 0.14 ms + marcado 275 ms + pausa 0.47 ms  · 970->974->489 MB 
@  106.6s  pausa 0.34 ms + marcado 274 ms + pausa 0.28 ms  · 971->974->488 MB 
@  113.7s  pausa 0.17 ms + marcado 273 ms + pausa 0.15 ms  · 970->973->488 MB 
@  120.9s  pausa 0.32 ms + marcado 399 ms + pausa 0.48 ms  · 970->977->492 MB 
@  128.2s  pausa 0.38 ms + marcado 282 ms + pausa 0.23 ms  · 974->976->487 MB 
@  135.5s  pausa 0.23 ms + marcado 291 ms + pausa 0.21 ms  · 963->965->487 MB 
@  142.5s  pausa 0.32 ms + marcado 297 ms + pausa 0.007 ms  · 964->968->489 MB 
@  149.8s  pausa 58 ms + marcado 206 ms + pausa 0.003 ms  · 968->970->487 MB 
@  156.9s  pausa 0.17 ms + marcado 371 ms + pausa 0.26 ms  · 963->966->488 MB 
@  163.6s  pausa 0.42 ms + marcado 296 ms + pausa 0.17 ms  · 968->972->489 MB 
@  170.7s  pausa 0.14 ms + marcado 374 ms + pausa 1.0 ms  · 970->973->488 MB 
@  177.9s  pausa 0.16 ms + marcado 281 ms + pausa 0.006 ms  · 968->971->487 MB 
@  185.1s  pausa 0.18 ms + marcado 280 ms + pausa 0.035 ms  · 968->973->489 MB 
@  192.3s  pausa 0.21 ms + marcado 281 ms + pausa 0.21 ms  · 971->975->489 MB 
@  199.5s  pausa 0.13 ms + marcado 280 ms + pausa 0.016 ms  · 971->975->489 MB 
@  206.4s  pausa 0.23 ms + marcado 289 ms + pausa 0.006 ms  · 970->974->488 MB 
@  213.1s  pausa 0.20 ms + marcado 281 ms + pausa 0.52 ms  · 969->972->488 MB 
@  220.4s  pausa 0.23 ms + marcado 278 ms + pausa 0.013 ms  · 968->971->488 MB 
@  227.6s  pausa 0.26 ms + marcado 205 ms + pausa 0.043 ms  · 969->972->488 MB 
@  234.7s  pausa 0.32 ms + marcado 200 ms + pausa 0.16 ms  · 970->973->488 MB 
@  241.8s  pausa 0.14 ms + marcado 283 ms + pausa 0.22 ms  · 968->975->491 MB 
@  248.9s  pausa 0.41 ms + marcado 287 ms + pausa 0.12 ms  · 972->976->489 MB 
@  256.0s  pausa 0.43 ms + marcado 286 ms + pausa 0.21 ms  · 967->969->487 MB 
@  262.6s  pausa 0.14 ms + marcado 366 ms + pausa 0.055 ms  · 963->967->488 MB 
@  269.6s  pausa 0.18 ms + marcado 295 ms + pausa 0.23 ms  · 967->969->487 MB 
@  276.7s  pausa 0.14 ms + marcado 284 ms + pausa 0.18 ms  · 965->967->487 MB 
@  283.6s  pausa 0.21 ms + marcado 373 ms + pausa 0.52 ms  · 968->975->491 MB 
@  291.0s  pausa 0.37 ms + marcado 303 ms + pausa 0.004 ms  · 972->974->487 MB 
@  298.2s  pausa 0.14 ms + marcado 382 ms + pausa 0.48 ms  · 964->966->487 MB 
```

- Segundos medidos: 300 · peticiones: 10.823.802 (36.079 por segundo)
- p99 típico (mediana de los p99 por segundo): 0.95 ms · peor p99 de un segundo: 2.54 ms
- Peor latencia típica de un segundo (mediana de los máx.): 65.31 ms · la peor de todas: 96.67 ms
- Segundos con alguna petición de más de 10 ms: 300 · de ellos, dentro de una recolección (±1,5 s): 134
- Recolecciones: 42 (forzadas por los 2 minutos: 0) · una cada 7.1 s · marcado típico 290 ms

## Resumen

| variante | pet/s | p99 típico | peor p99 (1 s) | peor petición | segundos con >10 ms (en recolección) | recolecciones · marcado típico |
|---|---|---|---|---|---|---|
| Go 1.9.2 · caché con punteros | 36.775 | 0.81 ms | 92.51 ms | 104.28 ms | 300 (121) | 28 · 1387 ms |
| Go 1.26 · caché con punteros | 34.491 | 0.93 ms | 92.66 ms | 192.23 ms | 300 (116) | 28 · 1072 ms |
| Go 1.26 · caché sin punteros | 36.079 | 0.95 ms | 2.54 ms | 96.67 ms | 300 (134) | 42 · 290 ms |
