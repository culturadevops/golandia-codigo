# -*- coding: utf-8 -*-
"""¿Por qué Robyn dio 40.924, 45.963 y 77.084 pet/s en /suma?n=1 en distintas corridas? 10 arranques, mismo método que la sección 6
(esperar a los 2 LISTO, 10 s en frío + 20 s templado), y el %CPU de cada proceso a mitad de la carga templada. → SALIDA-robyn-n1.md"""
import statistics
import sys
import threading
import time

import correr as K

R = K.R
R.OUT.clear()
R.w("## Robyn, /suma?n=1, 10 arranques (sección 6 repetida)\n")
R.w("| arranque | frío pet/s | templado pet/s | %CPU por proceso (templado, a los 10 s) |")
R.w("|---|---|---|---|")
url = "http://srv:8080/suma?n=1"
temp = []
for i in range(10):
    t0 = K.arrancar("robyn", K.img("robyn"), {"SRV": "robyn"}, puerto=9300)
    K.esperar_200(9300, t0)
    frio, _ = R.hey(url, R.SEG_FRIO)
    cpu = []
    th = threading.Thread(target=lambda: (time.sleep(10), cpu.append(K.cpu_por_proceso()[:2])))
    th.start()
    rps, _ = R.hey(url, R.SEG_CALIENTE)
    th.join()
    R.sh(["docker", "rm", "-f", "srv"])
    temp.append(rps)
    R.w(f"| {i + 1} | {frio:.0f} | {rps:.0f} | {' · '.join(f'{c:.0f} %' for c in cpu[0])} |")
R.w(f"\nMediana templado: {statistics.median(temp):.0f} · mín {min(temp):.0f} · máx {max(temp):.0f}")
open(K.AQUI / "SALIDA-robyn-n1.md", "w", encoding="utf-8").write("\n".join(R.OUT) + "\n")
