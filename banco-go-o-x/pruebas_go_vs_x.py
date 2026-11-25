# -*- coding: utf-8 -*-
"""pruebas_go_vs_x.py — «¿Go o X?» con el MISMO banco de la revancha de Java (pruebas_go_java_revancha.py), para X = python, cpp,
php o swift (pedido del dueño 2026-10-24: vídeos nuevos del formato que más vistas trae, «¿Go o Java?»).

Mismo servidor y mismo contrato (/suma de CPU, /trabajo con objetos, mapa, orden y JSON a mano), servidores con 2 CPU y 1 GiB,
`hey -c 16` desde otro contenedor, y la comprobación de que todos devuelven exactamente lo mismo antes de medir. Dos añadidos:
  · «imagen desde cero»: `docker build --no-cache` de la etapa que compila o instala (descarga de dependencias incluida; la
    imagen base ya descargada no cuenta).
  · latencia a RITMO FIJO: el mismo número de peticiones por segundo para todos (la mitad de lo que aguanta la variante más lenta),
    porque a máxima carga la latencia sale de la cola, no del lenguaje (memoria: un cliente que satura da picos falsos).

    python investigacion/pruebas_go_vs_x.py python|cpp|php|swift   → recursos/pruebas-go-vs-<x>/SALIDA.md
"""
import re
import statistics
import sys
import time
from datetime import datetime
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
import pruebas_go_java_revancha as R  # noqa: E402

AQUI = Path(__file__).resolve().parent.parent / "recursos"
GO_SRC = AQUI / "pruebas-go-java-revancha" / "go"
GO_IMG = "golandia-govs-go:1"
GO_DF = ("FROM golang:1.27.2 AS build\nWORKDIR /src\nCOPY . /src\nRUN CGO_ENABLED=0 go build -ldflags='-s -w' -o /servidor .\n"
         "FROM scratch\nCOPY --from=build /servidor /servidor\nEXPOSE 8080\nENTRYPOINT [\"/servidor\"]\n")

X = {
    "python": dict(nombre="Python", img="golandia-govs-python:1", df="Dockerfile.python", etapa=None,
                   base="python:3.13.17-slim + FastAPI 0.162 + uvicorn 0.54",
                   variantes=[("python (1 proceso)", {"WORKERS": "1"}), ("python (2 procesos)", {"WORKERS": "2"})],
                   version=["docker", "run", "--rm", "--entrypoint", "python", "golandia-govs-python:1", "--version"]),
    "cpp": dict(nombre="C++", img="golandia-govs-cpp:1", df="Dockerfile.cpp", etapa="build",
                base="scratch + binario estático (g++ 14.4, cpp-httplib 0.63.0)",
                variantes=[("c++ (2 hilos)", {"HILOS": "2"}), ("c++ (32 hilos)", {"HILOS": "32"})],
                version=["docker", "run", "--rm", "gcc:14.4.0", "g++", "--version"]),
    "php": dict(nombre="PHP", img="golandia-govs-php:1", df="Dockerfile.php", etapa=None,
                base="dunglas/frankenphp:1.13.3-php8.5",
                variantes=[("php clásico (FrankenPHP)", {}),
                           ("php worker (FrankenPHP)", {"MODO": "worker", "FRANKENPHP_CONFIG": "worker /app/public/index.php"}),
                           ("php worker + JIT", {"MODO": "worker", "FRANKENPHP_CONFIG": "worker /app/public/index.php",
                                                 "PHP_INI_SCAN_DIR": ":/app/jit"})],
                version=["docker", "run", "--rm", "--entrypoint", "frankenphp", "golandia-govs-php:1", "version"]),
    "swift": dict(nombre="Swift", img="golandia-govs-swift:1", df="Dockerfile.swift", etapa="build",
                  base="ubuntu:noble + binario (Swift 6.4.1, Vapor 4.122.3, stdlib estática)",
                  variantes=[("swift (Vapor)", {})],
                  version=["docker", "run", "--rm", "swift:6.4.1-noble", "swift", "--version"]),
}


def construir(x, cfg, base):
    """Imágenes, con `docker build --no-cache` cronometrado para la etapa que compila o instala."""
    comun = AQUI / "pruebas-go-vs-comun"
    comun.mkdir(exist_ok=True)
    (comun / "Dockerfile.go").write_text(GO_DF, encoding="utf-8")
    filas = []
    for nombre, df, ctx, tag, etapa in [("go", str(comun / "Dockerfile.go"), GO_SRC, GO_IMG, "build"),
                                        (x, cfg["df"], base, cfg["img"], cfg["etapa"])]:
        args = ["docker", "build", "-q", "--no-cache", "-f", str(ctx / df), "-t", tag]  # ctx / ruta absoluta = la ruta
        if etapa:
            args = ["docker", "build", "-q", "--no-cache", "--target", etapa, "-f", str(ctx / df), "-t", tag + "-etapa"]
        t0 = time.perf_counter()
        r = R.sh(args + [str(ctx)])
        seg = time.perf_counter() - t0
        assert r.returncode == 0, r.stderr[-2000:]
        if etapa:  # la imagen final, con caché (ya se cronometró la compilación)
            r = R.sh(["docker", "build", "-q", "-f", str(ctx / df), "-t", tag, str(ctx)])
            assert r.returncode == 0, r.stderr[-2000:]
        filas.append((nombre, seg))
    return filas


def arranque_y_memoria():
    """Como R.arranque_y_memoria, pero sin exigir la marca LISTO_EN_MS (PHP clásico no tiene «arranque» propio)."""
    R.w(f"## 3 · Arranque: contenedor nuevo → primera respuesta 200 (mediana de {R.REPES_FRIO}) y memoria en reposo\n")
    R.w("| variante | 1ª respuesta (docker incluido) | «listo» (su propio reloj) | memoria en reposo |")
    R.w("|---|---|---|---|")
    for nombre, imagen, env in R.VARIANTES:
        ts, listos, mems = [], [], []
        for _ in range(R.REPES_FRIO):
            t0 = R.arrancar(nombre, imagen, env, puerto=9300)
            ts.append(R.esperar_200(9300, t0))
            log = R.sh(["docker", "logs", "srv"]).stderr + R.sh(["docker", "logs", "srv"]).stdout
            m = re.search(r"LISTO_EN_MS ([\d.]+)", log)
            if m:
                listos.append(float(m.group(1)))
            time.sleep(1)
            mems.append(R.mem_mib())
        R.sh(["docker", "rm", "-f", "srv"])
        listo = f"{statistics.median(listos):.1f} ms" if listos else "—"
        R.w(f"| {nombre} | {statistics.median(ts):.0f} ms | {listo} | {statistics.median(mems):.1f} MiB |")
    R.w()


def carga_max(ruta, n, titulo):
    """R.carga, devolviendo las peticiones/s templadas de cada variante (para fijar luego el ritmo fijo)."""
    R.w(f"## {titulo}\n")
    R.w(f"`hey -c {R.C}` desde otro contenedor; servidor con {R.CPUS} CPU y {R.MEM}. «En frío»: los primeros {R.SEG_FRIO} s tras "
        f"arrancar. «Templado»: los {R.SEG_CALIENTE} s siguientes.\n")
    R.w("| variante | peticiones/s en frío | peticiones/s templado | p50 | p99 | memoria máx. bajo carga |")
    R.w("|---|---|---|---|---|---|")
    import threading
    res = {}
    for nombre, imagen, env in R.VARIANTES:
        t0 = R.arrancar(nombre, imagen, env, puerto=9300)
        R.esperar_200(9300, t0)
        url = f"http://srv:8080/{ruta}?n={n}"
        frio, _ = R.hey(url, R.SEG_FRIO)
        pico = [0.0]
        fin = threading.Event()

        def muestrear():
            while not fin.is_set():
                try:
                    pico[0] = max(pico[0], R.mem_mib())
                except Exception:
                    pass

        th = threading.Thread(target=muestrear)
        th.start()
        rps, lat = R.hey(url, R.SEG_CALIENTE)
        fin.set()
        th.join()
        R.sh(["docker", "rm", "-f", "srv"])
        res[nombre] = rps
        R.w(f"| {nombre} | {frio:.0f} | {rps:.0f} | {lat[50]:.1f} ms | {lat[99]:.1f} ms | {pico[0]:.0f} MiB |")
    R.w()
    return res


RITMO = 5  # peticiones/s de UN cliente: cada petición llega sola (a 5/s ninguna variante hace cola)


def ritmo_fijo(ruta, n, _total_rps, titulo):
    """Latencia de una petición SOLA, al mismo ritmo para todos: un cliente (`hey -c 1 -q 5`).

    ⚠️ La primera versión (2026-10-24) usaba 16 clientes con `-q`: hey los dispara a la vez en cada tic, así que las 16 peticiones
    llegaban juntas y hacían cola aunque el ritmo medio fuera la mitad de la capacidad (Python daba 195 ms a «ritmo fijo», casi lo
    mismo que saturado). Con un solo cliente no hay cola: lo que se mide es lo que tarda el servidor en atender."""
    q = RITMO
    R.w(f"## {titulo}\n")
    R.w(f"Un solo cliente a {q} peticiones/s (`hey -c 1 -q {q}`), {R.SEG_CALIENTE} s, tras {R.SEG_FRIO} s de calentamiento a tope. "
        "Cada petición llega sola: no hay cola, es lo que tarda el servidor en atender una.\n")
    R.w("| variante | peticiones/s logradas | p50 | p99 |")
    R.w("|---|---|---|---|")
    for nombre, imagen, env in R.VARIANTES:
        t0 = R.arrancar(nombre, imagen, env, puerto=9300)
        R.esperar_200(9300, t0)
        url = f"http://srv:8080/{ruta}?n={n}"
        R.hey(url, R.SEG_FRIO)
        r = R.sh(["docker", "run", "--rm", "--network", R.RED, "--cpus", "4", "golandia-revancha-hey:1",
                  "-z", f"{R.SEG_CALIENTE}s", "-c", "1", "-q", str(q), url])
        t = r.stdout
        rps = float(re.search(r"Requests/sec:\s+([\d.]+)", t).group(1))
        lat = {p: float(re.search(rf"\s{p}% in ([\d.]+) secs", t).group(1)) * 1000 for p in (50, 99)}
        R.sh(["docker", "rm", "-f", "srv"])
        R.w(f"| {nombre} | {rps:.0f} | {lat[50]:.2f} ms | {lat[99]:.2f} ms |")
    R.w()


def solo_ritmo(x, cfg, base):
    """`--solo-ritmo`: rehace solo las secciones 6 y 7 (latencia a ritmo fijo) de una SALIDA.md ya escrita."""
    salida = base / "SALIDA.md"
    previo = salida.read_text(encoding="utf-8").split("\n## 6 ")[0].rstrip("\n").split("\n")
    R.OUT[:] = previo + [""]
    R.sh(["docker", "network", "create", R.RED])
    R.VARIANTES[:] = [("go", GO_IMG, {})] + [(n, cfg["img"], env) for n, env in cfg["variantes"]]
    ritmo_fijo("trabajo", R.N_TRABAJO, 0, "6 · Latencia de una petición sola: /trabajo")
    ritmo_fijo("suma", R.N_SUMA, 0, "7 · Latencia de una petición sola: /suma")
    salida.write_text("\n".join(R.OUT) + "\n", encoding="utf-8")
    print(f"\n[ok] {salida} (secciones 6 y 7 rehechas)")


def main():
    x = sys.argv[1]
    cfg = X[x]
    base = AQUI / f"pruebas-go-vs-{x}"
    R.OUT.clear()
    if "--solo-ritmo" in sys.argv:
        return solo_ritmo(x, cfg, base)
    R.sh(["docker", "network", "create", R.RED])
    R.w(f"# SALIDA — ¿Go o {cfg['nombre']}? (golandia), con el banco de la revancha\n")
    R.w(f"> Corrida del {datetime.now():%Y-%m-%d %H:%M}. Regenerable con `python investigacion/pruebas_go_vs_x.py {x}`. "
        "⛔ Ninguna cifra de blog: todo salió de aquí.\n")
    filas = construir(x, cfg, base)
    ver_go = R.sh(["docker", "run", "--rm", "golang:1.27.2", "go", "version"]).stdout.strip()
    v = R.sh(cfg["version"])
    ver_x = (v.stdout or v.stderr).strip().splitlines()[0]
    R.w(f"- go: `{ver_go}` (net/http, solo librería estándar)")
    R.w(f"- {x}: `{ver_x}` — {cfg['base']}")
    R.w(f"- Docker `{R.sh(['docker', 'version', '--format', '{{.Server.Version}}']).stdout.strip()}`. Servidores con {R.CPUS} CPU y "
        f"{R.MEM}; `hey -c {R.C}` con 4 CPU desde otro contenedor.\n")
    R.w("## 1 · Construir la imagen desde cero (`docker build --no-cache`, dependencias incluidas)\n")
    R.w("| lenguaje | qué se cronometra | segundos |")
    R.w("|---|---|---|")
    for nombre, seg in filas:
        que = {"go": "go build", "python": "pip install de FastAPI + uvicorn", "cpp": "g++ -O2 -static",
               "php": "copiar el código (PHP no compila)", "swift": "swift package resolve + swift build -c release"}[nombre]
        R.w(f"| {nombre} | {que} | {seg:.1f} |")
    R.w()
    R.w("## 2 · Lo que despliegas\n")
    R.w("| imagen | base | tamaño |")
    R.w("|---|---|---|")
    for nombre, tag, b in [("go", GO_IMG, "scratch + binario estático"), (x, cfg["img"], cfg["base"])]:
        s = int(R.sh(["docker", "image", "inspect", "-f", "{{.Size}}", tag]).stdout.strip())
        R.w(f"| {nombre} | {b} | {s / 1048576:.1f} MiB |")
    R.w()
    R.VARIANTES[:] = [("go", GO_IMG, {})] + [(n, cfg["img"], env) for n, env in cfg["variantes"]]
    R.mismo_resultado()
    arranque_y_memoria()
    t = carga_max("trabajo", R.N_TRABAJO, "4 · Carga con memoria: /trabajo (objetos, mapa, orden, JSON), a tope")
    s = carga_max("suma", R.N_SUMA, "5 · Carga de CPU pura: /suma, a tope")
    ritmo_fijo("trabajo", R.N_TRABAJO, 0, "6 · Latencia de una petición sola: /trabajo")
    ritmo_fijo("suma", R.N_SUMA, 0, "7 · Latencia de una petición sola: /suma")
    (base / "SALIDA.md").write_text("\n".join(R.OUT) + "\n", encoding="utf-8")
    print(f"\n[ok] {base / 'SALIDA.md'}")


if __name__ == "__main__":
    main()
