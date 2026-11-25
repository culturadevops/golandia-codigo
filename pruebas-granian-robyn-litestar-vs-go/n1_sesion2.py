# -*- coding: utf-8 -*-
"""Sección 6 (/suma?n=1 a tope) repetida 3 veces, las seis variantes, orden rotado, en UNA sesión (tras el apagón del 2026-11-09):
así Robyn (que varía con el reparto entre procesos) y los demás se comparan en el mismo estado de la máquina. → SALIDA-n1-sesion2.md"""
import statistics

import correr as K
from pruebas_go_vs_x import carga_max

R = K.R
res = {}
todo = []
for vuelta in range(3):
    R.OUT.clear()
    R.VARIANTES[:] = K.VARIANTES[vuelta * 2:] + K.VARIANTES[:vuelta * 2]
    r = carga_max("suma", 1, f"6 · /suma?n=1 a tope — vuelta {vuelta + 1}")
    todo += R.OUT[:]
    for k, v in r.items():
        res.setdefault(k, []).append(v)
todo += ["## Mediana de las 3 vueltas (templado, pet/s)\n", "| variante | vueltas | mediana |", "|---|---|---|"]
for nombre, _, _ in K.VARIANTES:
    todo.append(f"| {nombre} | {', '.join(f'{v:.0f}' for v in res[nombre])} | {statistics.median(res[nombre]):.0f} |")
open(K.AQUI / "SALIDA-n1-sesion2.md", "w", encoding="utf-8").write("\n".join(todo) + "\n")
print("\n".join(todo[-9:]))
