# SALIDA — cómputo puro por lenguaje (reloj del propio proceso, sin red)

> Corrida del 2026-10-24 09:28. Regenerable con `python investigacion/pruebas_go_vs_x_computo.py`. 50 peticiones de calentamiento y 300 medidas, de una en una; el tiempo es el campo «ns» que mide cada servidor alrededor de su propio cálculo.

| lenguaje | /suma?n=200000 mediana | p99 | /trabajo?n=20000 mediana | p99 |
|---|---|---|---|---|
| go 1.27.2 | 0.122 ms | 0.179 ms | 3.377 ms | 3.818 ms |
| python 3.13 (FastAPI) | 8.191 ms | 8.889 ms | 9.008 ms | 20.741 ms |
| c++ (g++ 14.4 -O2) | 0.109 ms | 0.156 ms | 2.710 ms | 3.155 ms |
| php 8.5 worker (FrankenPHP) | 0.674 ms | 0.761 ms | 14.019 ms | 15.349 ms |
| php 8.5 worker + JIT | 0.402 ms | 0.450 ms | 12.328 ms | 14.393 ms |
| php 8.5 clásico (FrankenPHP) | 0.674 ms | 0.775 ms | 11.961 ms | 12.927 ms |
| swift 6.4.1 (Vapor) | 0.110 ms | 0.161 ms | 2.769 ms | 3.122 ms |

