# -*- coding: utf-8 -*-
"""pruebas_go_vs_x_computo.py — el cómputo PURO de cada lenguaje, medido por el reloj del propio proceso (campo «ns» de la
respuesta), sin red ni servidor HTTP de por medio. Complementa pruebas_go_vs_x.py (2026-10-24): en C++ el servidor HTTP
(cpp-httplib) atendía menos peticiones que Go aunque su bucle era más rápido, y había que separar «lenguaje» de «servidor».

Por imagen: contenedor con 2 CPU y 1 GiB, 50 peticiones de calentamiento y 300 medidas, de una en una (sin concurrencia).
    python investigacion/pruebas_go_vs_x_computo.py   → recursos/pruebas-go-vs-comun/SALIDA-computo.md
"""
import json
import statistics
import sys
import urllib.request
from datetime import datetime
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
import pruebas_go_java_revancha as R  # noqa: E402

SALIDA = Path(__file__).resolve().parent.parent / "recursos" / "pruebas-go-vs-comun" / "SALIDA-computo.md"
IMGS = [("go 1.27.2", "golandia-govs-go:1", {}),
        ("python 3.13 (FastAPI)", "golandia-govs-python:1", {"WORKERS": "1"}),
        ("c++ (g++ 14.4 -O2)", "golandia-govs-cpp:1", {"HILOS": "32"}),
        ("php 8.5 worker (FrankenPHP)", "golandia-govs-php:1", {"MODO": "worker", "FRANKENPHP_CONFIG": "worker /app/public/index.php"}),
        ("php 8.5 worker + JIT", "golandia-govs-php:1", {"MODO": "worker", "FRANKENPHP_CONFIG": "worker /app/public/index.php",
                                                         "PHP_INI_SCAN_DIR": ":/app/jit"}),
        ("php 8.5 clásico (FrankenPHP)", "golandia-govs-php:1", {}),
        ("swift 6.4.1 (Vapor)", "golandia-govs-swift:1", {})]
CAL, N = 50, 300


def medir(ruta):
    ns = []
    for i in range(CAL + N):
        j = json.loads(urllib.request.urlopen(f"http://127.0.0.1:9300{ruta}", timeout=30).read())
        if i >= CAL:
            ns.append(j["ns"])
    return statistics.median(ns) / 1e6, sorted(ns)[int(N * 0.99) - 1] / 1e6


def main():
    R.OUT.clear()
    R.sh(["docker", "network", "create", R.RED])
    R.w("# SALIDA — cómputo puro por lenguaje (reloj del propio proceso, sin red)\n")
    R.w(f"> Corrida del {datetime.now():%Y-%m-%d %H:%M}. Regenerable con `python investigacion/pruebas_go_vs_x_computo.py`. "
        f"{CAL} peticiones de calentamiento y {N} medidas, de una en una; el tiempo es el campo «ns» que mide cada servidor alrededor "
        "de su propio cálculo.\n")
    R.w(f"| lenguaje | /suma?n={R.N_SUMA} mediana | p99 | /trabajo?n={R.N_TRABAJO} mediana | p99 |")
    R.w("|---|---|---|---|---|")
    for nombre, img, env in IMGS:
        t0 = R.arrancar(nombre, img, env, puerto=9300)
        R.esperar_200(9300, t0)
        s50, s99 = medir(f"/suma?n={R.N_SUMA}")
        t50, t99 = medir(f"/trabajo?n={R.N_TRABAJO}")
        R.sh(["docker", "rm", "-f", "srv"])
        R.w(f"| {nombre} | {s50:.3f} ms | {s99:.3f} ms | {t50:.3f} ms | {t99:.3f} ms |")
    R.w()
    SALIDA.write_text("\n".join(R.OUT) + "\n", encoding="utf-8")
    print(f"[ok] {SALIDA}")


if __name__ == "__main__":
    main()
