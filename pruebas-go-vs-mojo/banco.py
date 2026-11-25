# -*- coding: utf-8 -*-
"""banco.py — «¿Go o Mojo?» (golandia 904) → SALIDA-corrida-<n>.md

Mojo 1.2 estable NO trae TCP en su librería estándar (`std.net` solo tiene la estructura: ver SALIDA-demos.md) y el servidor
web de la comunidad (lightbug_http) está archivado. Así que aquí no hay servidor web: se miden LAS MISMAS dos funciones de la
serie (/suma y /trabajo del servidor de Go) como programa de línea de comandos, en Go, en Mojo y en Python 3.13 (el lenguaje
del que Mojo se dice pariente), con el MISMO arnés en los tres (src/computo.mojo, go-computo/main.go, py-computo/computo.py):
50 llamadas de calentamiento y 300 medidas de cada función, de una en una, con el reloj del propio proceso.

Mismas condiciones que la serie: cada contenedor con 2 CPU y 1 GiB, antes de medir se comprueba que los tres devuelven lo
mismo (si no, se para) y se espera a que la máquina esté en reposo (carga de /proc/loadavg).

⛔ Solo con el candado de medición tomado (`mkdir <proyecto>/_logs/banco.lock`).

    python recursos/pruebas-go-vs-mojo/banco.py 1     → SALIDA-corrida-1.md
"""
import json
import statistics
import sys
import time
from datetime import datetime
from pathlib import Path

AQUI = Path(__file__).resolve().parent
sys.path.insert(0, str(AQUI.parent.parent / "investigacion"))
import pruebas_go_vs_nativo_2026_11 as N  # noqa: E402

L, R = N.L, N.R
CPUS, MEM = "2", "1g"
VAR = [  # nombre, Dockerfile, etapa que compila (None = no compila), comando dentro de la imagen «medir»
    ("go", "Dockerfile.go", "build", ["/computo"]),
    ("mojo", "Dockerfile.mojo", "build", ["/computo"]),
    ("python", "Dockerfile.python", None, ["python", "/computo.py"]),
]
OUT = R.OUT


def sh(args):
    return R.sh(args)


def run(img, cmd):
    return sh(["docker", "run", "--rm", "--ulimit", "core=1", "--cpus", CPUS, "--memory", MEM, "--entrypoint", cmd[0], img] + cmd[1:])


def construir():
    filas = []
    for nombre, df, etapa, _ in VAR:
        if etapa:
            t0 = time.perf_counter()
            r = sh(["docker", "build", "-q", "--no-cache", "--target", etapa, "-f", str(AQUI / df), "-t", f"gl-904-{nombre}-etapa:1",
                    str(AQUI)])
            seg = time.perf_counter() - t0
            assert r.returncode == 0, r.stderr[-2000:]
            filas.append((nombre, seg))
        for t in ("final", "medir"):
            r = sh(["docker", "build", "-q", "--target", t, "-f", str(AQUI / df), "-t", f"gl-904-{nombre}{'' if t == 'final' else '-medir'}:1",
                    str(AQUI)])
            assert r.returncode == 0, r.stderr[-2000:]
    return filas


def tam(tag):
    return int(sh(["docker", "image", "inspect", "-f", "{{.Size}}", tag]).stdout.strip()) / 1048576


def main():
    corrida = sys.argv[1]
    if not L.CANDADO.is_dir():
        raise SystemExit(f"⛔ sin candado: toma {L.CANDADO} antes de medir (mkdir)")
    OUT.clear()
    R.w(f"# SALIDA — ¿Go o Mojo? (golandia 904), corrida {corrida}: el cómputo de la serie, sin servidor web\n")
    R.w(f"> Corrida del {datetime.now():%Y-%m-%d %H:%M}, con el candado `_logs/banco.lock` tomado. Regenerable con "
        f"`python recursos/pruebas-go-vs-mojo/banco.py {corrida}`. ⛔ Ninguna cifra de blog: todo salió de aquí.\n")
    R.w("> Sin servidor web: Mojo 1.2 estable no trae TCP en su librería estándar (SALIDA-demos.md). Se miden las MISMAS dos "
        "funciones de /suma y /trabajo como programa, con el mismo arnés en los tres.\n")
    filas = construir()
    ver = {
        "go": sh(["docker", "run", "--rm", "golang:1.27.2", "go", "version"]).stdout.strip(),
        "mojo": sh(["docker", "run", "--rm", "--entrypoint", "mojo", "gl-904-mojo-etapa:1", "--version"]).stdout.strip(),
        "python": sh(["docker", "run", "--rm", "--entrypoint", "python", "gl-904-python:1", "--version"]).stdout.strip(),
    }
    for k, v in ver.items():
        R.w(f"- {k}: `{v}`")
    R.w(f"- Docker `{sh(['docker', 'version', '--format', '{{.Server.Version}}']).stdout.strip()}`. Cada contenedor con {CPUS} CPU "
        f"y {MEM}.\n")

    R.w("## 1 · Construir desde cero (`docker build --no-cache` de la etapa que compila, dependencias incluidas)\n")
    R.w("| lenguaje | qué se cronometra | segundos |")
    R.w("|---|---|---|")
    que = {"go": "go build", "mojo": "apt install gcc + pip install mojo==1.2.0 + mojo build"}
    for nombre, seg in filas:
        R.w(f"| {nombre} | {que[nombre]} | {seg:.1f} |")
    R.w("| python | (no compila: se copia el .py) | — |")
    R.w()

    R.w("## 2 · Lo que despliegas\n")
    R.w("| imagen | base | tamaño |")
    R.w("|---|---|---|")
    bases = {"go": "scratch + binario estático", "mojo": "debian:trixie-slim + binario + 3 librerías de Modular",
             "python": "python:3.13.17-slim + computo.py"}
    for nombre, *_ in VAR:
        R.w(f"| {nombre} | {bases[nombre]} | {tam(f'gl-904-{nombre}:1'):.1f} MiB |")
    b_go = int(sh(["docker", "run", "--rm", "--entrypoint", "stat", "gl-904-go-etapa:1", "-c", "%s", "/computo"]).stdout) / 1048576
    b_mo = int(sh(["docker", "run", "--rm", "--entrypoint", "stat", "gl-904-mojo-etapa:1", "-c", "%s", "/computo"]).stdout) / 1048576
    libs = sh(["docker", "run", "--rm", "--entrypoint", "du", "gl-904-mojo-etapa:1", "-sb", "/libs"]).stdout.split()[0]
    R.w(f"\nEl binario: Go, {b_go:.2f} MiB (estático); Mojo, {b_mo:.3f} MiB, más {int(libs) / 1048576:.2f} MiB de librerías "
        "de Modular que tiene que llevar al lado (libKGENCompilerRTShared, libAsyncRTRuntimeGlobals, libMSupportGlobals).\n")

    N.esperar_reposo({"id": 904})

    R.w("## 0 · ¿Hacen lo mismo? (antes de medir)\n")
    ref = None
    for nombre, _, _, cmd in VAR:
        r = run(f"gl-904-{nombre}-medir:1", cmd + ["una"])
        s, t = [json.loads(x) for x in r.stdout.strip().splitlines()]
        v = (t["total"], [x["id"] for x in t["primeros"]], s["suma"])
        ref = ref or v
        R.w(f"- {nombre}: total {v[0]}, primeros ids {v[1][:5]}…, suma {v[2]} → {'igual que Go' if v == ref else '⛔ DISTINTO'}")
        if v != ref:
            raise SystemExit("no devuelven lo mismo: no se compara")
    R.w()

    R.w("## 3 · Arranque del proceso (sin argumentos: arranca y termina), mediana de 30, dentro del contenedor\n")
    R.w("| lenguaje | mediana | p90 |")
    R.w("|---|---|---|")
    for nombre, _, _, cmd in VAR:
        c = " ".join(cmd)
        r = run(f"gl-904-{nombre}-medir:1", ["sh", "-c", f"for i in $(seq 33); do s=$(date +%s%N); {c}; e=$(date +%s%N); "
                                                       "echo $((e-s)); done"])
        ms = sorted(int(x) / 1e6 for x in r.stdout.split()[3:])
        R.w(f"| {nombre} | {statistics.median(ms):.2f} ms | {ms[int(len(ms) * 0.9) - 1]:.2f} ms |")
    R.w()

    R.w("## 4 · Cómputo puro: el reloj del propio proceso\n")
    R.w(f"50 llamadas de calentamiento y 300 medidas de cada función, de una en una. /suma?n={R.N_SUMA} y /trabajo?n={R.N_TRABAJO}, "
        "las mismas del servidor de Go. Memoria máxima del proceso con GNU time (`%M`) durante las 700 llamadas.\n")
    R.w(f"| lenguaje | suma mediana | p99 | trabajo mediana | p99 | memoria máx. | tiempo total |")
    R.w("|---|---|---|---|---|---|---|")
    for nombre, _, _, cmd in VAR:
        r = run(f"gl-904-{nombre}-medir:1", ["/usr/bin/time", "-f", "TIEMPO %M %e"] + cmd + ["medir"])
        d = {}
        for l in (r.stdout + r.stderr).splitlines():
            p = l.split()
            if p and p[0] in ("SUMA", "TRABAJO"):
                d[p[0]] = sorted(int(x) for x in p[1:])
            if p and p[0] == "TIEMPO":
                kb, seg = int(p[1]), float(p[2])
        s, t = d["SUMA"], d["TRABAJO"]
        assert len(s) == len(t) == 300, (len(s), len(t))
        R.w(f"| {nombre} | {statistics.median(s) / 1e6:.3f} ms | {s[296] / 1e6:.3f} ms | {statistics.median(t) / 1e6:.3f} ms | "
            f"{t[296] / 1e6:.3f} ms | {kb / 1024:.1f} MiB | {seg:.2f} s |")
    R.w()

    R.w("## 5 · `mojo run` contra el binario: lo que tarda `computo una` de punta a punta (mediana de 5)\n")
    R.w("`mojo run` compila el archivo cada vez antes de ejecutarlo; el binario de `mojo build` ya está compilado.\n")
    R.w("| cómo | mediana |")
    R.w("|---|---|")
    for nombre, img, c in [("mojo run src/computo.mojo una", "gl-904-mojo-etapa:1", "mojo run /src/computo.mojo una"),
                           ("binario de mojo build: /computo una", "gl-904-mojo-medir:1", "/computo una"),
                           ("binario de go build: /computo una", "gl-904-go-medir:1", "/computo una")]:
        r = run(img, ["sh", "-c", f"for i in 1 2 3 4 5 6; do s=$(date +%s%N); {c} > /dev/null; e=$(date +%s%N); echo $((e-s)); done"])
        ms = [int(x) / 1e6 for x in r.stdout.split()[1:]]
        R.w(f"| {nombre} | {statistics.median(ms):.0f} ms |")
    R.w()
    c, linea = N.carga_1min()
    R.w(f"Al terminar: /proc/loadavg `{linea}`; ajenos: " + (", ".join(f"{n} {c}" for n, c in N.ajenos()) or "ninguno") + "\n")
    (AQUI / f"SALIDA-corrida-{corrida}.md").write_text("\n".join(OUT) + "\n", encoding="utf-8")
    print(f"\n[ok] {AQUI / f'SALIDA-corrida-{corrida}.md'}")


if __name__ == "__main__":
    main()
