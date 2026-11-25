# SALIDA (solo sección G) — corrida del 2026-11-09 19:04

## G · ¿Libres de GIL? UN proceso con 1 y con 2 hilos de trabajo, /suma (CPU pura) a tope

`hey -c 16` 20 s tras 10 s de calentamiento; servidor con 2 CPU y 1g. CPU medida en el cgroup cada 0,5 s (muestreo.py): 100 % = un núcleo entero, el techo con 2 CPU es 200 %. Series en series/gil-*.json.

| variante | peticiones/s | p50 | CPU media | CPU máx. | hilos (máx.) |
|---|---|---|---|---|---|
| go, 1 núcleo (GOMAXPROCS=1) | 7524 | 2.1 ms | 100 % | 102 % | 3 |
| go, 2 núcleos (GOMAXPROCS=2) | 14861 | 1.0 ms | 199 % | 203 % | 4 |
| robyn, 1 proceso × 1 worker | 122 | 130.5 ms | 100 % | 102 % | 4 |
| robyn, 1 proceso × 2 workers | 118 | 135.2 ms | 100 % | 102 % | 5 |
| granian, 1 proceso (fastapi def) | 117 | 131.7 ms | 102 % | 104 % | 22 |
| robyn, 2 procesos × 1 worker | 238 | 81.1 ms | 200 % | 203 % | 9 |

