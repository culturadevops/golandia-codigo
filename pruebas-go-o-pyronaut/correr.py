# -*- coding: utf-8 -*-
"""correr.py — «¿Go o Pyronaut?» (golandia; pedido en un comentario del vídeo 848, no se cita).

Pyronaut 0.1.2 (micronaut-projects/pyronaut, PyPI `pyronaut`): Python sobre GraalPy (JIT de Truffle) dentro de la JVM, con el
servidor HTTP de Micronaut (Netty). El MISMO banco que el 848 (recursos/pruebas-granian-robyn-litestar-vs-go):
  · la MISMA lógica (proyecto/src/banco/logica.py, copiada tal cual de app/logica.py del 848) y el MISMO servidor Go (net/http);
  · servidores con 2 CPU y 1 GiB; `hey -c 16` con 4 CPU desde otro contenedor; antes de medir, el mismo cuerpo byte a byte;
  · referencias de Python del 848 sin tocar: FastAPI + uvicorn (la de siempre) y Robyn (la mejor de Python en el 848).
  · Pyronaut se empaqueta como FAT JAR (`pyronaut build --jar`) y corre con `java -jar` en el GraalVM JDK de `pyronaut setup`,
    con las opciones por defecto de la JVM (sin -Xmx: lo que haría quien sigue la guía).
Añadido para Pyronaut:
  · J: el calentamiento del JIT, petición a petición (mediana del campo ns por tramos);
  · L: carga sostenida 60 s en tramos de 10 s, para ver si «templado» a los 30 s es de verdad templado;
  · G: un solo proceso con y sin el pool de contextos de GraalPy (micronaut.python.pool.enabled), CPU del cgroup.

    python correr.py                → SALIDA.md (~35 min, la máquina sin otras cargas; con el candado banco.lock)
    python correr.py --comprobar    → solo construir + sección 0, sin carga
⛔ Ninguna cifra de blog: lo que diga el vídeo sale de la SALIDA de aquí.
"""
import importlib.util
import json
import re
import statistics
import sys
import time
import urllib.request
from datetime import datetime
from pathlib import Path

AQUI = Path(__file__).resolve().parent
G = Path(r"<proyecto>\canales\golandia")
B848 = G / "recursos" / "pruebas-granian-robyn-litestar-vs-go"
_spec = importlib.util.spec_from_file_location("banco848", B848 / "correr.py")
c848 = importlib.util.module_from_spec(_spec)
_spec.loader.exec_module(c848)  # deja R.RED = "fwpy" y R.arrancar/esperar_200 esperando a TODOS los LISTO_EN_MS
R, X = c848.R, c848.X
from muestreo import Muestreo  # noqa: E402  (investigacion/, ya en sys.path por el 848)

PYRO_IMG = "gl-prac-pyronaut:1"
SERIES = AQUI / "series"
VARIANTES = [  # (nombre, imagen, entorno). WORKERS=1 en Pyronaut: un solo proceso → se espera UN LISTO_EN_MS
    ("go (net/http)", c848.GO_IMG, {}),
    ("fastapi + uvicorn", c848.img("fastapi"), {"SRV": "uvicorn"}),
    ("robyn", c848.img("robyn"), {"SRV": "robyn"}),
    ("pyronaut (JVM)", PYRO_IMG, {"WORKERS": "1"}),
]


def construir_pyro(sin_cache=True):
    base = ["docker", "build", "-q", "-t", PYRO_IMG]
    r = R.sh(base + [str(AQUI)])
    assert r.returncode == 0, r.stderr[-2000:]
    if not sin_cache:
        return 0.0
    t0 = time.perf_counter()
    r = R.sh(base[:3] + ["--no-cache"] + base[3:] + [str(AQUI)])
    seg = time.perf_counter() - t0
    assert r.returncode == 0, r.stderr[-2000:]
    r = R.sh(["docker", "build", "-q", "--target", "sdk", "-t", "gl-prac-pyronaut-sdk:1", str(AQUI)])  # de la caché: para la versión
    assert r.returncode == 0, r.stderr[-2000:]
    return seg


def pedir(ruta, n, puerto=9300):
    with urllib.request.urlopen(f"http://127.0.0.1:{puerto}/{ruta}?n={n}") as r:
        return json.loads(r.read())


def calentamiento():
    """J: el campo ns (reloj del cálculo, sin red) petición a petición, desde la primera, en un contenedor nuevo."""
    tramos = [(1, 1), (2, 10), (11, 50), (51, 200), (201, 500)]
    R.w("## J · El calentamiento: lo que tarda el cálculo, petición a petición, desde la primera (mediana del campo ns por tramo)\n")
    R.w("Contenedor nuevo; peticiones de una en una. El campo ns es el reloj del propio cálculo, sin red ni framework.\n")
    cab = " | ".join(f"{a}-{b}" if a != b else f"{a}ª" for a, b in tramos)
    for ruta, n in (("suma", R.N_SUMA), ("trabajo", R.N_TRABAJO)):
        R.w(f"**/{ruta}?n={n}** (ms)\n")
        R.w(f"| variante | {cab} |")
        R.w("|---|" + "---|" * len(tramos))
        for nombre, imagen, env in R.VARIANTES:
            t0 = R.arrancar(nombre, imagen, env, puerto=9300)
            R.esperar_200(9300, t0)
            ns = [pedir(ruta, n)["ns"] / 1e6 for _ in range(tramos[-1][1])]
            R.sh(["docker", "rm", "-f", "srv"])
            celdas = [f"{statistics.median(ns[a - 1:b]):.2f}" for a, b in tramos]
            R.w(f"| {nombre} | {' | '.join(celdas)} |")
        R.w()


def sostenida(ruta, n, titulo, tramos=6, seg=10):
    """L: el mismo contenedor bajo carga 60 s seguidos, en tramos de 10 s."""
    R.w(f"## {titulo}\n")
    R.w(f"Contenedor nuevo; `hey -c {R.C}` {tramos} veces {seg} s seguidas, sin pausa. Peticiones/s de cada tramo, p99 del último "
        "y memoria máxima.\n")
    R.w("| variante | " + " | ".join(f"{(i) * seg}-{(i + 1) * seg} s" for i in range(tramos)) + " | p99 último | memoria máx. |")
    R.w("|---|" + "---|" * (tramos + 2))
    import threading
    url = f"http://srv:8080/{ruta}?n={n}"
    for nombre, imagen, env in R.VARIANTES:
        t0 = R.arrancar(nombre, imagen, env, puerto=9300)
        R.esperar_200(9300, t0)
        pico, fin = [0.0], threading.Event()

        def muestrear():
            while not fin.is_set():
                try:
                    pico[0] = max(pico[0], R.mem_mib())
                except Exception:
                    pass

        th = threading.Thread(target=muestrear)
        th.start()
        vals, lat = [], None
        for _ in range(tramos):
            rps, lat = R.hey(url, seg)
            vals.append(rps)
        fin.set()
        th.join()
        R.sh(["docker", "rm", "-f", "srv"])
        R.w(f"| {nombre} | " + " | ".join(f"{v:.0f}" for v in vals) + f" | {lat[99]:.1f} ms | {pico[0]:.0f} MiB |")
    R.w()


GIL = [
    ("go, 1 núcleo (GOMAXPROCS=1)", c848.GO_IMG, {"GOMAXPROCS": "1"}),
    ("go, 2 núcleos (GOMAXPROCS=2)", c848.GO_IMG, {"GOMAXPROCS": "2"}),
    ("fastapi + uvicorn, 1 proceso", c848.img("fastapi"), {"SRV": "uvicorn", "WORKERS": "1"}),
    ("pyronaut, sin pool (1 contexto)", PYRO_IMG, {"WORKERS": "1", "MICRONAUT_PYTHON_POOL_ENABLED": "false"}),
    ("pyronaut, pool por defecto", PYRO_IMG, {"WORKERS": "1"}),
]


def contextos():
    log = R.sh(["docker", "logs", "srv"])
    return len(re.findall(r"LISTO_EN_MS", log.stdout + log.stderr))


def sin_gil():
    R.w("## G · ¿Libre de GIL? UN proceso, /suma (CPU pura) a tope\n")
    R.w(f"`hey -c {R.C}` 20 s tras 30 s de calentamiento (el JIT necesita más que los 10 s del 848); servidor con {R.CPUS} CPU y "
        f"{R.MEM}. CPU del cgroup cada 0,5 s (muestreo.py): 100 % = un núcleo, el techo es 200 %. «Contextos»: cuántas veces se "
        "importó la app (una por contexto de GraalPy). Series en series/gil-*.json.\n")
    R.w("| variante | peticiones/s | p50 | CPU media | CPU máx. | hilos (máx.) | contextos |")
    R.w("|---|---|---|---|---|---|---|")
    SERIES.mkdir(exist_ok=True)
    url = f"http://srv:8080/suma?n={R.N_SUMA}"
    for nombre, imagen, env in GIL:
        t0 = R.arrancar(nombre, imagen, env, puerto=9300)
        R.esperar_200(9300, t0)
        R.hey(url, 30)
        with Muestreo("srv") as m:
            rps, lat = R.hey(url, 20)
        clave = re.sub(r"[^a-z0-9]+", "-", nombre.lower()).strip("-")
        m.guardar(SERIES / f"gil-{clave}.json", csv_tambien=False, extra={"variante": nombre, "rps": rps})
        r = m.resumen(desde=1.0)
        ctx = contextos() if "pyronaut" in nombre else "—"
        R.sh(["docker", "rm", "-f", "srv"])
        R.w(f"| {nombre} | {rps:.0f} | {lat[50]:.1f} ms | {r['cpu_media']:.0f} % | {r['cpu_max']:.0f} % | {r['hilos_max']} | {ctx} |")
    R.w()


TRAMPAS = r"""
set +e
echo '### T1 · lo que instala `pyronaut setup` (du -sh)'
du -sh /root/.local/share/pyronaut /root/.local/share/pyronaut/sdks/* /root/.m2 2>/dev/null
echo
echo '### T2 · el modo «un solo archivo» (pyronaut dev main.py) con la lógica en otro archivo'
mkdir -p /t2 && cd /t2 && cp /p/src/banco/logica.py . && printf 'import logica\nfrom micronaut.http.annotation import Get\n\n@Get("/suma")\ndef suma(n: int = 0) -> str:\n    return logica.suma(n)\n' > main.py
cat main.py
timeout 150 pyronaut dev main.py > /t2/log 2>&1
echo "rc=$?"
grep -m2 -E "ModuleNotFoundError" /t2/log
echo
echo '### T3 · un __init__.py en el paquete, como en cualquier proyecto de Python'
cp -r /p /t3 && touch /t3/src/banco/__init__.py && cd /t3
pyronaut build --jar > /t3/log 2>&1
echo "rc=$?"
grep -m1 -E "Custom __init__" /t3/log
"""


def trampas():
    """`--trampas` → SALIDA-trampas.md: lo que tropecé al montar Pyronaut, reproducido (sin cronómetro)."""
    R.OUT.clear()
    R.w("# SALIDA-trampas — Pyronaut 0.1.2: lo que se tropieza al montarlo\n")
    R.w(f"> Corrida del {datetime.now():%Y-%m-%d %H:%M}. En la etapa `build` de la imagen (gl-prac-pyronaut-build:1: "
        "python:3.13.17-slim + pyronaut 0.1.2 + `pyronaut setup` + el proyecto en /p).\n")
    r = R.sh(["docker", "build", "-q", "--target", "build", "-t", "gl-prac-pyronaut-build:1", str(AQUI)])
    assert r.returncode == 0, r.stderr[-1500:]
    r = R.sh(["docker", "run", "--rm", "--ulimit", "core=1", "--entrypoint", "bash", "gl-prac-pyronaut-build:1", "-c", TRAMPAS])
    R.w("```")
    R.w((r.stdout + r.stderr).strip())
    R.w("```")
    (AQUI / "SALIDA-trampas.md").write_text("\n".join(R.OUT) + "\n", encoding="utf-8")
    print(f"[ok] {AQUI / 'SALIDA-trampas.md'}")


def main():
    if "--trampas" in sys.argv:
        trampas()
        return
    comprobar = "--comprobar" in sys.argv
    R.OUT.clear()
    R.VARIANTES[:] = VARIANTES
    R.w("# SALIDA — ¿Go o Pyronaut? (golandia; pedido en un comentario del 848)\n")
    R.w(f"> Corrida del {datetime.now():%Y-%m-%d %H:%M}. Regenerable con `python correr.py`. ⛔ Ninguna cifra de blog: todo salió "
        "de aquí. El banco del 848 (misma lógica, mismo servidor Go, mismas condiciones) con Pyronaut añadido.\n")
    seg = construir_pyro(sin_cache=not comprobar)
    v = R.sh(["docker", "run", "--rm", "--entrypoint", "sh", "gl-prac-pyronaut-sdk:1", "-c", "pyronaut --version"])
    jv = R.sh(["docker", "run", "--rm", "--entrypoint", "/opt/graalvm/bin/java", PYRO_IMG, "-version"])
    R.w("- Pyronaut (`pyronaut --version` en la etapa de construcción):\n")
    R.w("```")
    R.w((v.stdout or v.stderr).strip())
    R.w("```")
    R.w("- JVM de la imagen: `" + (jv.stderr or jv.stdout).strip().splitlines()[0] + "`")
    R.w("- Resto, sin tocar desde el 848: `golandia-govs-go:1` (go1.27.2, net/http, FROM scratch), "
        "`golandia-rust-py-fastapi:1` (Python 3.13.17, fastapi 0.162.0 + uvicorn 0.54.0, 2 procesos), "
        "`golandia-rust-py-robyn:1` (robyn 0.88.0, 2 procesos).")
    R.w(f"- Servidores con {R.CPUS} CPU y {R.MEM}. Pyronaut: 1 proceso (la JVM), rutas `def` → ejecutor «io» de Micronaut, pool de "
        "contextos de GraalPy por defecto. `hey -c 16` con 4 CPU desde otro contenedor.\n")
    R.w("## 1 · Construir la imagen de Pyronaut desde cero (`docker build --no-cache`)\n")
    R.w(f"{seg:.0f} s. Incluye `pip install pyronaut` y `pyronaut setup`, que DESCARGA el GraalVM JDK, GraalPy y las dependencias "
        "de Maven (en el 848 las imágenes de Python tardaron 8,7-17,7 s y la de Go 4,0 s, con la imagen base ya descargada).\n")
    R.w("## 2 · Lo que despliegas\n")
    R.w("| imagen | tamaño (docker image inspect .Size, el mismo dato que el 848) |")
    R.w("|---|---|")
    for nombre, im in (("pyronaut (JVM)", PYRO_IMG), ("fastapi", c848.img("fastapi")), ("robyn", c848.img("robyn")),
                       ("go", c848.GO_IMG)):
        s = int(R.sh(["docker", "image", "inspect", "-f", "{{.Size}}", im]).stdout.strip())
        R.w(f"| {nombre} | {s / 1048576:.1f} MiB |")
    R.w()
    jar = R.sh(["docker", "run", "--rm", "--entrypoint", "sh", PYRO_IMG, "-c", "ls -l /app/banco.jar; du -sh /opt/graalvm"])
    R.w("Por dentro (`ls -l /app/banco.jar; du -sh /opt/graalvm`):\n")
    R.w("```")
    R.w(jar.stdout.strip())
    R.w("```\n")

    c848.mismo_resultado()
    if comprobar:
        (AQUI / "SALIDA-comprobacion.md").write_text("\n".join(R.OUT) + "\n", encoding="utf-8")
        print(f"\n[ok] {AQUI / 'SALIDA-comprobacion.md'} (sin carga)")
        return
    X.arranque_y_memoria()
    c848.computo_una()
    calentamiento()
    X.carga_max("trabajo", R.N_TRABAJO, "4 · Carga con memoria: /trabajo (objetos, dict, orden, JSON), a tope")
    X.carga_max("suma", R.N_SUMA, "5 · Carga de CPU pura: /suma, a tope")
    X.carga_max("suma", 1, "6 · El coste del servidor: /suma?n=1 (casi sin cálculo), a tope")
    sostenida("trabajo", R.N_TRABAJO, "L · Carga sostenida: /trabajo durante 60 s, en tramos de 10 s")
    X.ritmo_fijo("trabajo", R.N_TRABAJO, 0, "7 · Latencia de una petición sola: /trabajo")
    X.ritmo_fijo("suma", 1, 0, "8 · Latencia de una petición sola: /suma?n=1 (el servidor solo)")
    sin_gil()
    (AQUI / "SALIDA.md").write_text("\n".join(R.OUT) + "\n", encoding="utf-8")
    print(f"\n[ok] {AQUI / 'SALIDA.md'}")


if __name__ == "__main__":
    main()
