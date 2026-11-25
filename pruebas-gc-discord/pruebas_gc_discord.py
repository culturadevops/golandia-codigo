# -*- coding: utf-8 -*-
"""pruebas_gc_discord.py — reproduce el caso del blog de Discord (4-feb-2020, Jesse Howarth, «Why Discord is switching from Go to
Rust»): una caché enorme, poca basura nueva, y la recolección que Go FUERZA cada 2 minutos. Tres variantes, 5 minutos cada una:
  · Go 1.9.2 (la versión que usaba Discord) con la caché de punteros
  · Go 1.26 con la misma caché de punteros
  · Go 1.26 con la caché SIN punteros (array de structs dentro del mapa)
Servidor con 2 CPU y 4 GiB, GODEBUG=gctrace=1; cliente con 16 trabajadores midiendo cada segundo.
Guarda `recursos/pruebas-gc-discord/SALIDA.md` (+ las salidas crudas en _crudo/).
    python investigacion/pruebas_gc_discord.py
"""
import io
import re
import statistics
import subprocess
import sys
import time
from datetime import datetime
from pathlib import Path

sys.stdout = io.TextIOWrapper(sys.stdout.buffer, encoding="utf-8")
BASE = Path(__file__).resolve().parent.parent / "recursos" / "pruebas-gc-discord"
BIN, CRUDO = BASE / "_bin", BASE / "_crudo"
RED, SEG, K, N, RPS = "gcdiscord", 300, 16, "3000000", 2000
OUT = []


def sh(a, **kw):
    return subprocess.run(a, capture_output=True, text=True, encoding="utf-8", errors="replace", **kw)


def w(s=""):
    print(s, flush=True)
    OUT.append(s)


def win(p):
    return str(p).replace("/", "\\")


def compilar():
    BIN.mkdir(exist_ok=True)
    CRUDO.mkdir(exist_ok=True)
    # 2026-10-19: Go 1.26 se compila CON go.mod (go 1.26). Sin él, Go usa los valores por defecto antiguos y NO ajusta GOMAXPROCS al
    # límite de CPU del contenedor (salía «24 P» con --cpus 2). Go 1.9.2 no conoce los módulos: GOPATH, como en 2020.
    for img, src, sal, modulos in [("golang:1.9.2", "servidor", "srv-go1.9.2", False), ("golang:1.26", "servidor", "srv-go1.26", True),
                                   ("golang:1.26", "cliente", "cliente", False)]:
        r = sh(["docker", "run", "--rm", "-e", "CGO_ENABLED=0", "-e", "GO111MODULE=" + ("on" if modulos else "off"), "-v",
                f"{win(BASE / src)}:/src", "-v", f"{win(BIN)}:/bin_out", "-w", "/src", img, "go", "build", "-o", f"/bin_out/{sal}",
                "." if modulos else "main.go"])
        assert r.returncode == 0, (img, src, r.stderr)
    (BASE / "Dockerfile").write_text("FROM debian:trixie-slim\nCOPY _bin/ /app/\n", encoding="utf-8")
    assert sh(["docker", "build", "-q", "-t", "golandia-gcdiscord:1", str(BASE)]).returncode == 0


def correr(nombre, binario, modo, extra=None, clave=None):
    sh(["docker", "rm", "-f", "srv"])
    r = sh(["docker", "run", "-d", "--name", "srv", "--network", RED, "--cpus", "2", "--memory", "4g", "-e", "GODEBUG=gctrace=1",
            "-e", f"MODO={modo}", "-e", f"N={N}"] + sum((["-e", f"{k}={v}"] for k, v in (extra or {}).items()), []) +
           ["golandia-gcdiscord:1", f"/app/{binario}"])
    assert r.returncode == 0, r.stderr
    while "LISTO" not in sh(["docker", "logs", "srv"]).stderr:
        time.sleep(1)
    c = sh(["docker", "run", "--rm", "--network", RED, "--cpus", "4", "golandia-gcdiscord:1", "/app/cliente", "http://srv:8080",
            str(SEG), str(K), str(RPS)])
    log = sh(["docker", "logs", "srv"]).stderr
    sh(["docker", "rm", "-f", "srv"])
    clave = clave or f"{binario}-{modo}"
    (CRUDO / f"{clave}-cliente.txt").write_text(c.stdout, encoding="utf-8")
    (CRUDO / f"{clave}-gctrace.txt").write_text(log, encoding="utf-8")


FILA = re.compile(r"t=\s*(\d+)s pet=\s*(\d+) p50=\s*([\d.]+)ms p99=\s*([\d.]+)ms max=\s*([\d.]+)ms")
GC = re.compile(r"gc \d+ @([\d.]+)s \d+%: ([\d.]+)\+([\d.]+)\+([\d.]+) ms clock.*?(\d+)->(\d+)->(\d+) MB.*?(\(forced\))?$", re.M)


def analizar(clave):
    """Lee _crudo/. 2026-10-19: la primera versión tomaba el segundo (t=) como «peticiones» y corría las columnas un lugar."""
    cli = (CRUDO / f"{clave}-cliente.txt").read_text(encoding="utf-8")
    log = (CRUDO / f"{clave}-gctrace.txt").read_text(encoding="utf-8")
    filas = [dict(t=int(m[0]), pet=int(m[1]), p50=float(m[2]), p99=float(m[3]), max=float(m[4])) for m in FILA.findall(cli)]
    antes, tras = log.split("LISTO", 1)
    listo = "LISTO" + tras.splitlines()[0]
    # el reloj de gctrace cuenta desde que arrancó el proceso; la carga empezó justo después de LISTO (última recolección del llenado)
    previas = GC.findall(antes)
    desfase = float(previas[-1][0]) if previas else 0.0
    gcs = GC.findall(tras)
    ventanas = [(float(g[0]) - desfase, float(g[0]) - desfase + float(g[2]) / 1000) for g in gcs]
    return listo, filas, gcs, ventanas, desfase


def informe(nombre, clave, resumen):
    listo, filas, gcs, ventanas, desfase = analizar(clave)
    w(f"## {nombre}\n")
    w(f"`{listo.strip()}`\n")
    procs = re.findall(r"(\d+) P", (CRUDO / f"{clave}-gctrace.txt").read_text(encoding="utf-8"))
    w(f"GOMAXPROCS que usó el proceso (gctrace): {procs[-1] if procs else '?'} P\n")
    w("Recolecciones durante la carga (gctrace: inicio · pausa + marcado concurrente + pausa · montón):\n")
    w("```")
    for t, a, m, b, h1, h2, h3, forzada in gcs:
        w(f"@{float(t):7.1f}s  pausa {a} ms + marcado {m} ms + pausa {b} ms  · {h1}->{h2}->{h3} MB {forzada}")
    w("```\n")
    picos = [f for f in filas if f["max"] > 10]
    en_gc = [f for f in picos if any(a - 1.5 <= f["t"] <= b + 1.5 for a, b in ventanas)]
    pet = sum(f["pet"] for f in filas)
    w(f"- Segundos medidos: {len(filas)} · peticiones: {pet:,} ({pet / len(filas):,.0f} por segundo)".replace(",", "."))
    w(f"- p99 típico (mediana de los p99 por segundo): {statistics.median(f['p99'] for f in filas):.2f} ms · "
      f"peor p99 de un segundo: {max(f['p99'] for f in filas):.2f} ms")
    w(f"- Peor latencia típica de un segundo (mediana de los máx.): {statistics.median(f['max'] for f in filas):.2f} ms · "
      f"la peor de todas: {max(f['max'] for f in filas):.2f} ms")
    w(f"- Segundos con alguna petición de más de 10 ms: {len(picos)} · de ellos, dentro de una recolección (±1,5 s): {len(en_gc)}")
    w(f"- Recolecciones: {len(gcs)} (forzadas por los 2 minutos: {sum(1 for g in gcs if g[7])}) · una cada "
      f"{(ventanas[-1][0] - ventanas[0][0]) / max(1, len(ventanas) - 1):.1f} s · marcado típico "
      f"{statistics.median(float(g[2]) for g in gcs):.0f} ms\n")
    resumen.append((nombre, statistics.median(f["p99"] for f in filas), max(f["p99"] for f in filas), max(f["max"] for f in filas),
                    len(picos), len(en_gc), len(gcs), statistics.median(float(g[2]) for g in gcs), pet / len(filas)))


VARIANTES = [  # (nombre, binario, modo, entorno extra, clave de los crudos)
    ("Go 1.9.2 · caché con punteros", "srv-go1.9.2", "punteros", {}, "go1.9.2-punteros"),
    ("Go 1.9.2 · caché con punteros · GOMAXPROCS=2", "srv-go1.9.2", "punteros", {"GOMAXPROCS": "2"}, "go1.9.2-punteros-procs2"),
    ("Go 1.26 · caché con punteros", "srv-go1.26", "punteros", {}, "go1.26-punteros"),
    ("Go 1.26 · caché sin punteros", "srv-go1.26", "plano", {}, "go1.26-plano"),
]


def main():
    if "--resumen" not in sys.argv:   # --resumen: solo rehace SALIDA.md desde _crudo/
        sh(["docker", "network", "create", RED])
        compilar()
        for nombre, binario, modo, extra, clave in VARIANTES:
            if "--saltar-hechos" in sys.argv and (CRUDO / f"{clave}-cliente.txt").exists():
                continue
            correr(nombre, binario, modo, extra, clave)
    w("# SALIDA — el caso Discord, reproducido (golandia)\n")
    w(f"> Análisis del {datetime.now():%Y-%m-%d %H:%M} . Regenerable con "
      f"`python investigacion/pruebas_gc_discord.py` (o `--resumen` desde _crudo/). Servidor con 2 CPU y 4 GiB, "
      f"{int(N):,} usuarios en caché; cliente con {K} trabajadores a {RPS} peticiones/s en total durante {SEG} s (ritmo fijo: servicio sin saturar, como el de Discord).\n".replace(",", "."))
    resumen = []
    for nombre, binario, modo, extra, clave in VARIANTES:
        informe(nombre, clave, resumen)
    w("## Resumen\n")
    w("| variante | pet/s | p99 típico | peor p99 (1 s) | peor petición | segundos con >10 ms (en recolección) | recolecciones · marcado típico |")
    w("|---|---|---|---|---|---|---|")
    for r in resumen:
        w(f"| {r[0]} | {r[8]:,.0f} | {r[1]:.2f} ms | {r[2]:.2f} ms | {r[3]:.2f} ms | {r[4]} ({r[5]}) | {r[6]} · {r[7]:.0f} ms |".replace(",", "."))
    (BASE / "SALIDA.md").write_text("\n".join(OUT) + "\n", encoding="utf-8")
    print("\n[ok] SALIDA.md escrita")


if __name__ == "__main__":
    main()
