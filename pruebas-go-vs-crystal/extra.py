# -*- coding: utf-8 -*-
"""extra.py — ¿por qué Crystal con HILOS=2 se hunde en /trabajo? → SALIDA-extra.md (golandia 902)

En SALIDA.md, con 2 hilos /suma sube y /trabajo (20.000 objetos por petición) se hunde. Antes de contarlo, se aísla:
la misma imagen con 1 y con 2 hilos, `hey` con 1, 2, 4 y 16 clientes durante 10 s, y la CPU que usa el servidor
mientras tanto (docker stats). Mismo servidor, mismas condiciones (2 CPU y 1 GiB) que el banco.

⛔ Con el candado de medición tomado, después de pruebas_go_vs_x_2026_11_23.py crystal (usa sus imágenes gl-902-*).

    python recursos/pruebas-go-vs-crystal/extra.py
"""
import re
import statistics
import sys
import threading
from datetime import datetime
from pathlib import Path

AQUI = Path(__file__).resolve().parent
sys.path.insert(0, str(AQUI.parent.parent / "investigacion"))
import pruebas_go_vs_x_2026_11_23 as X  # noqa: E402

N, L, R = X.N, X.L, X.R
if not L.CANDADO.is_dir():
    raise SystemExit("⛔ sin candado")
X.ACTUAL = 902
N.ajenos = X.ajenos
cfg = L.X["crystal"]
X.preparar(cfg)
R.OUT.clear()
R.sh(["docker", "rm", "-f", "gl-902-srv"])
R.sh(["docker", "network", "rm", R.RED])
R.sh(["docker", "network", "create", "--subnet", cfg["subred"], R.RED])
R.w("# SALIDA-extra — Crystal: 1 hilo contra 2 hilos según el número de clientes (golandia 902)\n")
R.w(f"> Corrida del {datetime.now():%Y-%m-%d %H:%M}, con el candado tomado. Servidor con 2 CPU y 1 GiB; hey con 4 CPU desde "
    "otro contenedor, 10 s por casilla. CPU = media de `docker stats` del servidor durante esos 10 s (200 % = los 2 núcleos).\n")
N.esperar_reposo(cfg)


def cpu():
    s = R.sh(["docker", "stats", "--no-stream", "--format", "{{.CPUPerc}}", "srv"]).stdout.strip().rstrip("%")
    return float(s)


def hey(ruta, c):
    usos, fin = [], threading.Event()

    def muestrear():
        while not fin.is_set():
            try:
                usos.append(cpu())
            except Exception:
                pass

    th = threading.Thread(target=muestrear)
    th.start()
    r = R.sh(["docker", "run", "--rm", "--network", R.RED, "--cpus", "4", "gl-902-hey:1", "-z", "10s", "-c", str(c),
              f"http://gl-902-srv:8080{ruta}"])
    fin.set()
    th.join()
    rps = float(re.search(r"Requests/sec:\s+([\d.]+)", r.stdout).group(1))
    codigos = re.findall(r"\[(\d+)\]\s+(\d+) responses", r.stdout)
    assert codigos and all(k == "200" for k, _ in codigos), codigos
    return rps, statistics.mean(usos[1:] or usos)


for ruta in (f"/trabajo?n={R.N_TRABAJO}", f"/suma?n={R.N_SUMA}"):
    R.w(f"## {ruta.split('?')[0]}\n")
    R.w("| clientes | 1 hilo: pet/s | CPU | HILOS=2: pet/s | CPU |")
    R.w("|---|---|---|---|---|")
    res = {}
    for nombre, env in (("1", {}), ("2", {"HILOS": "2"})):
        t0 = R.arrancar(nombre, "gl-902-crystal:1", env, puerto=9300)
        R.esperar_200(9300, t0)
        hey(ruta, 16)  # calentamiento
        for c in (1, 2, 4, 16):
            res[(nombre, c)] = hey(ruta, c)
        R.sh(["docker", "rm", "-f", "srv"])
    for c in (1, 2, 4, 16):
        a, b = res[("1", c)], res[("2", c)]
        R.w(f"| {c} | {a[0]:.0f} | {a[1]:.0f} % | {b[0]:.0f} | {b[1]:.0f} % |")
    R.w()
c, linea = N.carga_1min()
R.w(f"Al terminar: /proc/loadavg `{linea}`; ajenos: " + (", ".join(f"{n} {c}" for n, c in N.ajenos()) or "ninguno") + "\n")
(AQUI / "SALIDA-extra.md").write_text("\n".join(R.OUT) + "\n", encoding="utf-8")
R.sh(["docker", "network", "rm", R.RED])
print("\n".join(R.OUT))
