# -*- coding: utf-8 -*-
"""Corre las 10 lecciones dentro de Docker, con la CPU fija, y regenera SALIDA.md.

La regla de la casa: nada de lo que sale en el video se escribe a mano. Sale de aqui.
Hace falta Docker y Python 3.10+. Go no hace falta en tu maquina: todo corre en golang:1.27.2.

    python ejecutar_todo.py                    # todo, ~10 minutos; borra contenedor y volumenes
    python ejecutar_todo.py --solo 03          # una leccion, a la pantalla (no toca SALIDA.md)
    python ejecutar_todo.py --conservar-cache  # deja los volumenes de modulos y compilacion
    python ejecutar_todo.py --cpus 4           # otra cantidad de CPU (por defecto, 2)

Los numeros son de UNA maquina. En la tuya saldran otros: lo que se compara es una version
contra otra en la misma maquina, no tus milisegundos contra los mios.
"""
import argparse
import os
import platform
import re
import subprocess
import sys
import time
from datetime import datetime, timezone
from pathlib import Path

AQUI = Path(__file__).resolve().parent
IMAGEN = "golang:1.27.2"
PREFIJO = os.environ.get("CURSO_PREFIJO", "curso-go-rend")
CONTENEDOR = f"{PREFIJO}-go"
VOL_MOD = f"{PREFIJO}-gomod"
VOL_CACHE = f"{PREFIJO}-gocache"

LECCIONES = [
    ("01", "Benchmarks bien hechos", "01_benchmarks"),
    ("02", "pprof de CPU", "02_pprof_cpu"),
    ("03", "pprof de memoria: alloc_space contra inuse_space", "03_memoria"),
    ("04", "Analisis de escape y asignaciones", "04_escape"),
    ("05", "sync.Pool, y cuando no sirve", "05_pool"),
    ("06", "go tool trace: goroutines bloqueadas", "06_trace"),
    ("07", "El recolector: GOGC y GOMEMLIMIT", "07_gc"),
    ("08", "La condicion de carrera que los tests no ven", "08_carrera"),
    ("09", "pprof en un servicio vivo", "09_vivo"),
    ("10", "El antes y el despues, con trafico", "10_carga"),
]

# Señales de que algo no corrio (los FAIL de la leccion 8 son a proposito y no cuentan).
ROTO = re.compile(r"\[build failed\]|cannot find package|not found|No such file|"
                  r"syntax error|undefined: |Input/output error")


def docker(*args, timeout=1800, check=False):
    r = subprocess.run(["docker", *args], capture_output=True, text=True, encoding="utf-8",
                       errors="replace", timeout=timeout)
    if check and r.returncode != 0:
        raise SystemExit(f"docker {' '.join(args[:3])} ... fallo:\n{r.stderr}")
    return r.returncode, (r.stdout or "") + (r.stderr or "")


def montajes():
    return ["-v", f"{AQUI}:/curso", "-v", f"{VOL_MOD}:/go/pkg/mod",
            "-v", f"{VOL_CACHE}:/root/.cache/go-build", "-w", "/curso"]


def levantar(cpus):
    docker("rm", "-f", CONTENEDOR)
    docker("run", "-d", "--rm", "--name", CONTENEDOR, f"--cpus={cpus}", *montajes(), IMAGEN,
           "sleep", "infinity", check=True)
    # Descarga los modulos (benchstat) una vez, antes de medir nada.
    docker("exec", CONTENEDOR, "go", "mod", "download", check=True)


def leccion(d):
    return docker("exec", CONTENEDOR, "sh", f"{d}/comandos.sh")


def oom(cpus):
    """La segunda parte de la leccion 7 necesita su propio contenedor con --memory."""
    nombre = f"{PREFIJO}-oom"
    docker("rm", "-f", nombre)
    c, s = docker("run", "--rm", "--name", nombre, f"--cpus={cpus}", "--memory=200m",
                  "--memory-swap=200m", *montajes(), IMAGEN, "sh", "07_gc/oom.sh")
    return c, s


def limpiar(conservar):
    docker("rm", "-f", CONTENEDOR)
    if not conservar:
        docker("volume", "rm", VOL_MOD, VOL_CACHE)


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--solo", help="numero de leccion (01..10)")
    ap.add_argument("--cpus", default="2")
    ap.add_argument("--conservar-cache", action="store_true")
    a = ap.parse_args()

    print(f"  levantando {IMAGEN} con --cpus={a.cpus} ...", flush=True)
    levantar(a.cpus)
    try:
        if a.solo:
            for num, titulo, d in LECCIONES:
                if num == a.solo:
                    print(leccion(d)[1])
                    if num == "07":
                        print(oom(a.cpus)[1])
            return 0

        sello = datetime.now(timezone.utc).strftime("%Y-%m-%d %H:%M:%SZ")
        _, gov = docker("exec", CONTENEDOR, "go", "version")
        _, cpu = docker("exec", CONTENEDOR, "sh", "-c",
                        "grep -m1 'model name' /proc/cpuinfo | cut -d: -f2; nproc")
        modelo, nucleos = (cpu.strip().splitlines() + ["?", "?"])[:2]
        _, dv = docker("version", "--format", "{{.Server.Version}}")
        partes = ["# SALIDA — curso «Rendimiento en Go medido: pprof, trace, GC y race»", "",
                  "> Generado por `ejecutar_todo.py`. **No se escribe a mano.**", "",
                  f"- fecha: `{sello}`",
                  f"- {gov.strip()} (imagen `{IMAGEN}`)",
                  f"- CPU: `{modelo.strip()}` · el anfitrion tiene {nucleos.strip()} nucleos; "
                  f"el contenedor, `--cpus={a.cpus}`",
                  f"- Docker {dv.strip()} · anfitrion `{platform.system()} {platform.release()}`",
                  "- **Los numeros son de esta maquina.** Lo que vale es la comparacion entre "
                  "versiones en la misma maquina.", ""]
        fallos = 0
        for num, titulo, d in LECCIONES:
            print(f"  [{num}] {titulo} ...", end=" ", flush=True)
            t0 = time.time()
            codigo, salida = leccion(d)
            roto = codigo != 0 or ROTO.search(salida)
            partes += [f"## Leccion {num} — {titulo}", "",
                       f"`sh {d}/comandos.sh` · codigo de salida `{codigo}`", "",
                       "```", salida.strip("\n"), "```", ""]
            if num == "07":
                c2, s2 = oom(a.cpus)
                roto = roto or ROTO.search(s2)
                partes += [f"### En un contenedor con 200 MB (`docker run --memory=200m`)", "",
                           f"`sh 07_gc/oom.sh` · codigo de salida `{c2}`", "",
                           "```", s2.strip("\n"), "```", ""]
            fallos += bool(roto)
            print(("FALLO" if roto else "ok") + f" ({time.time() - t0:.0f} s)")

        (AQUI / "SALIDA.md").write_text("\n".join(partes) + "\n", encoding="utf-8",
                                          newline="\n")
        print(f"\n  -> SALIDA.md regenerado (10 lecciones, {fallos} fallos)")
        return 1 if fallos else 0
    finally:
        limpiar(a.conservar_cache)


if __name__ == "__main__":
    sys.exit(main())
