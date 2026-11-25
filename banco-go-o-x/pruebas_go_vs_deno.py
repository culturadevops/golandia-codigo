# -*- coding: utf-8 -*-
"""pruebas_go_vs_deno.py — «¿Go o TypeScript con Deno?» (vídeo 699, 2026-10-29) con el MISMO banco que 651-654
(pruebas_go_vs_x.py + pruebas_go_vs_x_computo.py): mismo servidor y contrato, servidores con 2 CPU y 1 GiB, `hey -c 16` con 4 CPU
desde otro contenedor, la comprobación de que todos devuelven lo mismo antes de medir, latencia a ritmo fijo y cómputo puro.

No se copian las funciones: se importan de pruebas_go_vs_x.py. Lo único que cambia son los NOMBRES de Docker (contenedor, red,
puerto e imágenes con prefijo v699-), porque otros agentes usan a la vez el banco con «srv», «revancha» y el puerto 9300.

Variantes de Deno 2.9.8 (imagen oficial denoland/deno:2.9.8, TypeScript con tipos, Deno.serve):
  · deno run (1 hilo)          → `deno run --allow-net=0.0.0.0:8080 servidor.ts`
  · deno serve --parallel (2)  → un hilo de JavaScript por núcleo, mismo código como export por defecto
  · deno compile (binario)     → un ejecutable sobre gcr.io/distroless/cc-debian12 (la comparación con el binario de Go)

    python investigacion/pruebas_go_vs_deno.py   → recursos/pruebas-go-vs-deno/SALIDA.md
"""
import json
import re
import statistics
import sys
import time
import urllib.request
from datetime import datetime
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
import pruebas_go_vs_x as GX  # noqa: E402

R = GX.R
BASE = GX.AQUI / "pruebas-go-vs-deno"
HEY_CTX = GX.AQUI / "pruebas-go-java-revancha"

# ── nombres propios (v699-…) sin tocar el banco compartido ──────────────────────────────────────────────────────────────────
R.RED = "v699-red"
MAPA = {"srv": "v699-srv", "golandia-revancha-hey:1": "v699-hey:1", "9300:8080": "9699:8080"}
_sh, _urlopen = R.sh, urllib.request.urlopen


def sh(args, **kw):
    return _sh([MAPA.get(a, a.replace("http://srv:8080", "http://v699-srv:8080")) for a in args], **kw)


def urlopen(url, *a, **kw):
    return _urlopen(url.replace("127.0.0.1:9300", "127.0.0.1:9699"), *a, **kw)


R.sh = sh
urllib.request.urlopen = urlopen

IMG = {"go": "v699-go:1", "deno": "v699-deno:1", "deno-compile": "v699-deno-compile:1"}
VARIANTES = [("go", IMG["go"], {}),
             ("deno run (1 hilo)", IMG["deno"], {"MODO": "run"}),
             ("deno serve --parallel (2 hilos)", IMG["deno"], {"MODO": "paralelo"}),
             ("deno compile (binario)", IMG["deno-compile"], {})]
CONSTRUIR = [  # (nombre, qué se cronometra, Dockerfile, contexto, etiqueta, etapa)
    ("go", "go build", GX.AQUI / "pruebas-go-vs-comun" / "Dockerfile.go", GX.GO_SRC, IMG["go"], "build"),
    ("deno", "deno check (tipos) sobre denoland/deno:2.9.8", BASE / "Dockerfile.deno", BASE, IMG["deno"], None),
    ("deno compile", "deno compile (tipos + binario)", BASE / "Dockerfile.deno-compile", BASE, IMG["deno-compile"], "build"),
]


def construir():
    """Igual que GX.construir: `docker build --no-cache` cronometrado de la etapa que compila o instala."""
    (GX.AQUI / "pruebas-go-vs-comun" / "Dockerfile.go").write_text(GX.GO_DF, encoding="utf-8")
    r = R.sh(["docker", "build", "-q", "-f", str(HEY_CTX / "Dockerfile.hey"), "-t", "v699-hey:1", str(HEY_CTX)])
    assert r.returncode == 0, r.stderr[-2000:]
    filas = []
    for nombre, que, df, ctx, tag, etapa in CONSTRUIR:
        args = ["docker", "build", "-q", "--no-cache", "-f", str(df), "-t", tag]
        if etapa:
            args = ["docker", "build", "-q", "--no-cache", "--target", etapa, "-f", str(df), "-t", tag + "-etapa"]
        t0 = time.perf_counter()
        r = R.sh(args + [str(ctx)])
        seg = time.perf_counter() - t0
        assert r.returncode == 0, r.stderr[-2000:]
        if etapa:
            r = R.sh(["docker", "build", "-q", "-f", str(df), "-t", tag, str(ctx)])
            assert r.returncode == 0, r.stderr[-2000:]
        filas.append((nombre, que, seg))
    return filas


def tam_mib(tag):
    return int(R.sh(["docker", "image", "inspect", "-f", "{{.Size}}", tag]).stdout.strip()) / 1048576


def binario_mib(etapa_tag, ruta):
    r = R.sh(["docker", "run", "--rm", "--entrypoint", "stat", etapa_tag, "-c", "%s", ruta])
    return int(r.stdout.strip()) / 1048576


def computo():
    """Lo mismo que pruebas_go_vs_x_computo.py (50 de calentamiento, 300 medidas de una en una, campo «ns»), para estas variantes."""
    cal, n = 50, 300

    def medir(ruta):
        ns = []
        for i in range(cal + n):
            j = json.loads(urllib.request.urlopen(f"http://127.0.0.1:9300{ruta}", timeout=30).read())
            if i >= cal:
                ns.append(j["ns"])
        return statistics.median(ns) / 1e6, sorted(ns)[int(n * 0.99) - 1] / 1e6

    R.w("## 8 · Cómputo puro: el reloj del propio proceso, sin red\n")
    R.w(f"El método de `pruebas_go_vs_x_computo.py`: {cal} peticiones de calentamiento y {n} medidas, de una en una; el tiempo es el "
        "campo «ns» que mide cada servidor alrededor de su propio cálculo.\n")
    R.w(f"| variante | /suma?n={R.N_SUMA} mediana | p99 | /trabajo?n={R.N_TRABAJO} mediana | p99 |")
    R.w("|---|---|---|---|---|")
    for nombre, img, env in R.VARIANTES:
        t0 = R.arrancar(nombre, img, env, puerto=9300)
        R.esperar_200(9300, t0)
        s50, s99 = medir(f"/suma?n={R.N_SUMA}")
        t50, t99 = medir(f"/trabajo?n={R.N_TRABAJO}")
        R.sh(["docker", "rm", "-f", "srv"])
        R.w(f"| {nombre} | {s50:.3f} ms | {s99:.3f} ms | {t50:.3f} ms | {t99:.3f} ms |")
    R.w()


def main():
    R.OUT.clear()
    R.sh(["docker", "network", "create", R.RED])
    R.w("# SALIDA — ¿Go o TypeScript con Deno? (golandia 699), con el banco de la revancha\n")
    R.w(f"> Corrida del {datetime.now():%Y-%m-%d %H:%M}. Regenerable con `python investigacion/pruebas_go_vs_deno.py` (mismo banco "
        "que 651-654: `pruebas_go_vs_x.py` + `pruebas_go_vs_x_computo.py`). ⛔ Ninguna cifra de blog: todo salió de aquí.\n")
    filas = construir()
    ver_go = R.sh(["docker", "run", "--rm", "golang:1.27.2", "go", "version"]).stdout.strip()
    ver_deno = " · ".join(R.sh(["docker", "run", "--rm", "--entrypoint", "deno", IMG["deno"], "--version"]).stdout.split("\n")[:3])
    R.w(f"- go: `{ver_go}` (net/http, solo librería estándar)")
    R.w(f"- deno: `{ver_deno}` — denoland/deno:2.9.8, Deno.serve, sin dependencias")
    R.w(f"- Docker `{R.sh(['docker', 'version', '--format', '{{.Server.Version}}']).stdout.strip()}`. Servidores con {R.CPUS} CPU y "
        f"{R.MEM}; `hey -c {R.C}` con 4 CPU desde otro contenedor.\n")
    R.w("## 1 · Construir la imagen desde cero (`docker build --no-cache`, dependencias incluidas)\n")
    R.w("| lenguaje | qué se cronometra | segundos |")
    R.w("|---|---|---|")
    for nombre, que, seg in filas:
        R.w(f"| {nombre} | {que} | {seg:.1f} |")
    R.w()
    R.w("## 2 · Lo que despliegas\n")
    R.w("| imagen | base | tamaño |")
    R.w("|---|---|---|")
    for nombre, tag, b in [("go", IMG["go"], "scratch + binario estático"),
                           ("deno", IMG["deno"], "denoland/deno:2.9.8 + 3 archivos .ts"),
                           ("deno compile", IMG["deno-compile"], "distroless/cc-debian12 + binario de deno compile")]:
        R.w(f"| {nombre} | {b} | {tam_mib(tag):.1f} MiB |")
    R.w(f"\nEl binario de Go pesa {binario_mib('v699-go:1-etapa', '/servidor'):.1f} MiB; el de `deno compile`, "
        f"{binario_mib('v699-deno-compile:1-etapa', '/servidor'):.1f} MiB (lleva dentro el runtime de Deno y V8).\n")
    R.VARIANTES[:] = VARIANTES
    R.mismo_resultado()
    GX.arranque_y_memoria()
    GX.carga_max("trabajo", R.N_TRABAJO, "4 · Carga con memoria: /trabajo (objetos, mapa, orden, JSON), a tope")
    GX.carga_max("suma", R.N_SUMA, "5 · Carga de CPU pura: /suma, a tope")
    GX.ritmo_fijo("trabajo", R.N_TRABAJO, 0, "6 · Latencia de una petición sola: /trabajo")
    GX.ritmo_fijo("suma", R.N_SUMA, 0, "7 · Latencia de una petición sola: /suma")
    computo()
    (BASE / "SALIDA.md").write_text("\n".join(R.OUT) + "\n", encoding="utf-8")
    R.sh(["docker", "network", "rm", R.RED])
    print(f"\n[ok] {BASE / 'SALIDA.md'}")


if __name__ == "__main__":
    main()
