# -*- coding: utf-8 -*-
"""ejecutar_todo.py — corre las lecciones del curso de observabilidad en Go y regenera SALIDA.md.

Todo corre en Docker (golang:1.26). Las lecciones 3, 4 y 5 corren junto a un Prometheus de verdad
(prom/prometheus:v3.1.0) que raspa la lección cada segundo: cada PromQL del vídeo es su respuesta.

    python ejecutar_todo.py            todas
    python ejecutar_todo.py 03 05      solo esas
"""
import datetime
import os
import subprocess
import sys
import time
from pathlib import Path

AQUI = Path(__file__).resolve().parent
GO = "golang:1.26"
PROM = "prom/prometheus:v3.1.0"
RED = "curso-obs-red"
VOL = ["-v", f"{AQUI.as_posix()}:/curso", "-w", "/curso", "-v", "curso-grpc-gomod:/go/pkg/mod",
       "-v", "curso-grpc-gocache:/root/.cache/go-build"]
ENV = dict(os.environ, MSYS_NO_PATHCONV="1")

LECCIONES = [
    ("01", "01_slog", "slog: atributos, la clave que falta y el secreto"),
    ("02", "02_niveles", "Niveles: el Debug que no imprime y cuesta"),
    ("03", "03_sin_iniciar", "La métrica que no existe hasta el primer error"),
    ("04", "04_cardinalidad", "Cardinalidad: un id en una etiqueta"),
    ("05", "05_histogramas", "Histogramas: el p99 de los buckets por defecto"),
    ("06", "06_trazas", "Trazas: el contexto que se pierde"),
    ("07", "07_spans_perdidos", "Los spans que nunca llegan"),
    ("08", "08_todo_junto", "Todo junto: logs, métricas y trazas que se encuentran"),
]
CON_PROMETHEUS = {"03_sin_iniciar", "04_cardinalidad", "05_histogramas"}


def docker(*args, check=True):
    r = subprocess.run(["docker", *args], capture_output=True, text=True, encoding="utf-8",
                       errors="replace", env=ENV)
    if check and r.returncode != 0:
        raise SystemExit(f"[X] docker {' '.join(args[:3])}: {r.stderr.strip()}")
    return r


def con_prometheus(directorio):
    """La lección en un contenedor «app» y Prometheus raspándola, en su propia red."""
    docker("rm", "-f", "prometheus", "app", check=False)
    docker("network", "rm", RED, check=False)
    docker("network", "create", RED)
    try:
        # Se compila antes, para que el reloj de Prometheus no cuente la compilación.
        r = docker("run", "--rm", *VOL, GO, "go", "build", "-o", f"bin/{directorio}", f"./{directorio}", check=False)
        if r.returncode:
            return r.returncode, r.stdout + r.stderr
        docker("run", "-d", "--name", "prometheus", "--network", RED,
               "-v", f"{(AQUI / 'prometheus.yml').as_posix()}:/etc/prometheus/prometheus.yml:ro", PROM)
        r = docker("run", "--rm", "--name", "app", "--hostname", "app", "--network", RED,
                   "-e", "PROMETHEUS=http://prometheus:9090", *VOL, GO, f"/curso/bin/{directorio}", check=False)
        return r.returncode, (r.stdout + (r.stderr if r.returncode else "")).strip("\n")
    finally:
        docker("rm", "-f", "prometheus", "app", check=False)
        docker("network", "rm", RED, check=False)


def correr(directorio):
    if directorio in CON_PROMETHEUS:
        return con_prometheus(directorio)
    r = docker("run", "--rm", *VOL, GO, "go", "run", f"./{directorio}", check=False)
    return r.returncode, (r.stdout + (r.stderr if r.returncode else "")).strip("\n")


def main():
    pedidas = sys.argv[1:]
    elegidas = [l for l in LECCIONES if not pedidas or l[0] in pedidas]
    version = docker("run", "--rm", GO, "go", "version").stdout.strip()
    mods = docker("run", "--rm", *VOL, GO, "go", "list", "-m", "github.com/prometheus/client_golang",
                  "go.opentelemetry.io/otel", "go.opentelemetry.io/otel/sdk",
                  "go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp").stdout.split("\n")
    partes = ["# SALIDA — curso «Observabilidad en Go: slog, Prometheus y OpenTelemetry»", "",
              "> Generado por `ejecutar_todo.py`. **No se escribe a mano.**", "",
              f"- fecha: `{datetime.datetime.now(datetime.timezone.utc):%Y-%m-%d %H:%M:%SZ}`",
              f"- {version}", *[f"- `{m}`" for m in mods if m], f"- Prometheus: `{PROM}`",
              f"- Docker `{docker('version', '--format', '{{.Server.Version}}').stdout.strip()}`"]
    fallos = 0
    for num, d, titulo in elegidas:
        if not (AQUI / d).exists():
            continue
        t0 = time.time()
        codigo, out = correr(d)
        fallos += codigo != 0
        print(f"[{'ok' if codigo == 0 else 'X '}] {num} {d} ({time.time() - t0:.0f} s)")
        partes += ["", f"## Leccion {num} — {titulo}", "",
                   f"`go run ./{d}` · codigo de salida `{codigo}`", "", "```", out, "```"]
    if not pedidas:
        (AQUI / "SALIDA.md").write_bytes(("\n".join(partes) + "\n").encode("utf-8"))
        print("[ok] SALIDA.md")
    else:
        print("\n".join(partes[9:]))
    sys.exit(1 if fallos else 0)


if __name__ == "__main__":
    main()
