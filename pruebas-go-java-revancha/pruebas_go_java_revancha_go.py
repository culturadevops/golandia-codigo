# -*- coding: utf-8 -*-
"""pruebas_go_java_revancha_go.py — anexo de la revancha (vídeo 506): ¿POR QUÉ Java le gana a Go en /trabajo?
Mismo método que pruebas_go_java_revancha.py (hey -c 16, servidor con 2 CPU y 1 GiB). Hipótesis:
  1. el recolector de Go: GOGC=400, y GOMEMLIMIT=700MiB + GOGC=off (deja crecer la memoria como hace Java)
  2. el orden: sort.Slice (usa reflexión para intercambiar) contra slices.SortFunc
  3. cómputo puro: el campo ns que declara cada servidor, petición a petición y sin concurrencia
Guarda `recursos/pruebas-go-java-revancha/SALIDA-por-que.md`.
    python investigacion/pruebas_go_java_revancha_go.py   (después de pruebas_go_java_revancha.py: usa sus imágenes)
"""
import json
import statistics
import sys
import urllib.request
from datetime import datetime
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
import pruebas_go_java_revancha as R  # noqa: E402

R.OUT.clear()
IMG_SLICES = "golandia-revancha-go-slices:1"


def construir_slices():
    r = R.sh(["docker", "run", "--rm", "-v", f"{R.win(R.BASE / 'go-slices')}:/src", "-v", f"{R.win(R.BIN)}:/bin_out", "-w", "/src",
              "-e", "CGO_ENABLED=0", "golang:1.26", "go", "build", "-ldflags=-s -w", "-o", "/bin_out/servidor-slices", "."])
    assert r.returncode == 0, r.stderr
    f = R.BASE / "Dockerfile.go-slices"
    f.write_text("FROM scratch\nCOPY _bin/servidor-slices /servidor\nEXPOSE 8080\nENTRYPOINT [\"/servidor\"]\n", encoding="utf-8")
    r = R.sh(["docker", "build", "-q", "-f", str(f), "-t", IMG_SLICES, str(R.BASE)])
    assert r.returncode == 0, r.stderr


def computo(imagen, env, n=300):
    """Mediana del campo ns (el reloj del propio servidor) en n peticiones seguidas, tras 200 de calentamiento."""
    t0 = R.arrancar("x", imagen, env, puerto=9300)
    R.esperar_200(9300, t0)
    url = f"http://127.0.0.1:9300/trabajo?n={R.N_TRABAJO}"
    for _ in range(200):
        urllib.request.urlopen(url).read()
    v = [json.loads(urllib.request.urlopen(url).read())["ns"] / 1e6 for _ in range(n)]
    R.sh(["docker", "rm", "-f", "srv"])
    v.sort()
    return statistics.median(v), v[int(0.99 * (len(v) - 1))]


def main():
    R.sh(["docker", "network", "create", R.RED])
    construir_slices()
    R.w("# SALIDA — ¿por qué Java le gana a Go en /trabajo? (anexo del vídeo 506)\n")
    R.w(f"> Corrida del {datetime.now():%Y-%m-%d %H:%M}. Regenerable con `python investigacion/pruebas_go_java_revancha_go.py`. "
        "Mismo método que SALIDA.md.\n")
    variantes = [
        ("go (por defecto: GOGC=100)", R.IMG["go"], {}),
        ("go GOGC=400", R.IMG["go"], {"GOGC": "400"}),
        ("go GOMEMLIMIT=700MiB GOGC=off", R.IMG["go"], {"GOMEMLIMIT": "700MiB", "GOGC": "off"}),
        ("go con slices.SortFunc", IMG_SLICES, {}),
        ("go slices.SortFunc + GOMEMLIMIT=700MiB GOGC=off", IMG_SLICES, {"GOMEMLIMIT": "700MiB", "GOGC": "off"}),
        ("java-jre (virtuales)", R.IMG["java-jre"], {"EJECUTOR": "virtual"}),
    ]
    R.w("## A · Cómputo puro de /trabajo: el reloj del propio servidor, una petición cada vez (sin concurrencia)\n")
    R.w("| variante | mediana | p99 |")
    R.w("|---|---|---|")
    for nombre, img, env in variantes:
        med, p99 = computo(img, env)
        R.w(f"| {nombre} | {med:.2f} ms | {p99:.2f} ms |")
    R.w()
    R.VARIANTES[:] = variantes
    R.carga("trabajo", R.N_TRABAJO, "B · Bajo carga: /trabajo con hey -c 16")
    (R.BASE / "SALIDA-por-que.md").write_text("\n".join(R.OUT) + "\n", encoding="utf-8")
    print("\n[ok] SALIDA-por-que.md escrita")


if __name__ == "__main__":
    main()
