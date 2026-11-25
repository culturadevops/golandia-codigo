# -*- coding: utf-8 -*-
"""extra.py — lo de Scala Native que no cabe en la carga de SALIDA.md → SALIDA-extra.md (golandia 865).

1. El cuelgue: el binario release-full TAL CUAL (Scala Native 0.5.12, Immix) bajo /suma a tope, 30 s, contra el mismo binario
   recompilado con SCALANATIVE_GC_TRAP_BASED_YIELDPOINTS=0 (el arreglo que propone el changelog de 0.5.12). Cuenta respuestas
   y las líneas «Unhandled signal 11» del log.
2. El modo por defecto de scala-cli (debug) contra release-full arreglado: cómputo puro con el reloj del propio proceso (mismo
   método que la sección 8 de SALIDA.md) y una carga corta de /trabajo.

⛔ Con el candado de medición tomado, después de pruebas_go_vs_lote_2026_11.py scala-native (usa sus imágenes v865-*).

    python recursos/pruebas-go-vs-scala-native/extra.py
"""
import re
import sys
import urllib.request
from datetime import datetime
from pathlib import Path

AQUI = Path(__file__).resolve().parent
sys.path.insert(0, str(AQUI.parent.parent / "investigacion"))
import pruebas_go_vs_lote_2026_11 as L  # noqa: E402

R = L.R
if not L.CANDADO.is_dir():
    raise SystemExit("⛔ sin candado")
L.preparar(L.X["scala-native"])
R.OUT.clear()
R.sh(["docker", "network", "create", "--subnet", "10.86.5.0/24", R.RED])
R.w("# SALIDA-extra — Scala Native: el cuelgue en release y el modo por defecto (golandia 865)\n")
R.w(f"> Corrida del {datetime.now():%Y-%m-%d %H:%M}, con el candado tomado. Servidores con 2 CPU y 1 GiB; hey -c 16 con 4 CPU.\n")


def hey(ruta, seg):
    r = R.sh(["docker", "run", "--rm", "--network", R.RED, "--cpus", "4", "v865-hey:1", "-z", f"{seg}s", "-c", "16",
              f"http://v865-srv:8080{ruta}"])
    rps = float(re.search(r"Requests/sec:\s+([\d.]+)", r.stdout).group(1))
    ok = re.search(r"\[200\]\s+(\d+) responses", r.stdout)
    return rps, int(ok.group(1)) if ok else 0


R.w("## 1 · /suma a tope durante 30 s: release-full tal cual contra release-full con el arreglo\n")
R.w("| binario | peticiones/s | respuestas 200 en 30 s | «Unhandled signal 11» en el log | ¿responde después? |")
R.w("|---|---|---|---|---|")
for nombre, imagen in (("release-full tal cual", "v865-scala-native:1"),
                       ("release-full + SCALANATIVE_GC_TRAP_BASED_YIELDPOINTS=0", "v865-scala-native-sintrap:1")):
    t0 = R.arrancar(nombre, imagen, {}, puerto=9300)
    R.esperar_200(9300, t0)
    rps, n = hey(f"/suma?n={R.N_SUMA}", 30)
    log = R.sh(["docker", "logs", "srv"])
    señales = (log.stdout + log.stderr).count("Unhandled signal 11")
    try:
        vivo = str(urllib.request.urlopen("http://127.0.0.1:9300/suma?n=10", timeout=5).status)
    except Exception:
        vivo = "sin respuesta en 5 s"
    R.w(f"| {nombre} | {rps:.1f} | {n} | {señales} | {vivo} |")
    R.sh(["docker", "rm", "-f", "srv"])
R.w()

R.VARIANTES[:] = [("scala native release-full (arreglado), HILOS=2", "v865-scala-native-sintrap:1", {"HILOS": "2"}),
                  ("scala native modo por defecto (debug), HILOS=2", "v865-scala-native-defecto:1", {"HILOS": "2"})]
L.computo()
R.w("## Carga corta: /trabajo con hey -c 16 durante 10 s\n")
for nombre, imagen, env in R.VARIANTES:
    t0 = R.arrancar(nombre, imagen, env, puerto=9300)
    R.esperar_200(9300, t0)
    rps, _ = hey(f"/trabajo?n={R.N_TRABAJO}", 10)
    R.w(f"- {nombre}: {rps:.1f} peticiones/s")
    R.sh(["docker", "rm", "-f", "srv"])
R.w()
(AQUI / "SALIDA-extra.md").write_text("\n".join(R.OUT) + "\n", encoding="utf-8")
R.sh(["docker", "network", "rm", R.RED])
print("\n".join(R.OUT))
