# -*- coding: utf-8 -*-
"""extra.py — lo de Kotlin/Native que no cabe en la carga de SALIDA.md → SALIDA-extra.md (golandia 879).

El binario debug (gradle linkDebugExecutableLinuxX64, el que lanza runDebugExecutableLinuxX64 mientras programas) contra el
release: cómputo puro con el reloj del propio proceso (mismo método que la sección 8 de SALIDA.md) y una carga corta de
/trabajo y de /suma.

⛔ Con el candado de medición tomado, después de pruebas_go_vs_nativo_2026_11.py kotlin-native (usa sus imágenes gl-879-*).

    python recursos/pruebas-go-vs-kotlin-native/extra.py
"""
import re
import sys
from datetime import datetime
from pathlib import Path

AQUI = Path(__file__).resolve().parent
sys.path.insert(0, str(AQUI.parent.parent / "investigacion"))
import pruebas_go_vs_nativo_2026_11 as N  # noqa: E402

L, R = N.L, N.R
if not L.CANDADO.is_dir():
    raise SystemExit("⛔ sin candado")
cfg = L.X["kotlin-native"]
N.preparar(cfg)
R.OUT.clear()
R.sh(["docker", "network", "rm", R.RED])
R.sh(["docker", "network", "create", "--subnet", cfg["subred"], R.RED])
R.w("# SALIDA-extra — Kotlin/Native: debug contra release (golandia 879)\n")
R.w(f"> Corrida del {datetime.now():%Y-%m-%d %H:%M}, con el candado tomado. Servidores con 2 CPU y 1 GiB; hey -c 16 con 4 CPU.\n")
N.esperar_reposo(cfg)


def hey(ruta, seg):
    r = R.sh(["docker", "run", "--rm", "--network", R.RED, "--cpus", "4", "gl-879-hey:1", "-z", f"{seg}s", "-c", "16",
              f"http://gl-879-srv:8080{ruta}"])
    rps = float(re.search(r"Requests/sec:\s+([\d.]+)", r.stdout).group(1))
    return rps


VAR = [("kotlin/native release, HILOS=2", "gl-879-kotlin-native:1", {"HILOS": "2"}),
       ("kotlin/native debug, HILOS=2", "gl-879-kotlin-native-debug:1", {"HILOS": "2"})]
R.VARIANTES[:] = [("go", "gl-879-go:1", {})] + VAR
R.mismo_resultado()
R.VARIANTES[:] = VAR
L.computo()
R.w("## Carga corta: hey -c 16 durante 10 s\n")
R.w("| variante | /trabajo pet/s | /suma pet/s |")
R.w("|---|---|---|")
for nombre, imagen, env in R.VARIANTES:
    t0 = R.arrancar(nombre, imagen, env, puerto=9300)
    R.esperar_200(9300, t0)
    t = hey(f"/trabajo?n={R.N_TRABAJO}", 10)
    s = hey(f"/suma?n={R.N_SUMA}", 10)
    R.w(f"| {nombre} | {t:.1f} | {s:.1f} |")
    R.sh(["docker", "rm", "-f", "srv"])
R.w()
(AQUI / "SALIDA-extra.md").write_text("\n".join(R.OUT) + "\n", encoding="utf-8")
R.sh(["docker", "network", "rm", R.RED])
print("\n".join(R.OUT))
