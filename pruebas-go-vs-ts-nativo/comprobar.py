# -*- coding: utf-8 -*-
"""comprobar.py — ¿de verdad hacen el mismo trabajo? (golandia 880) → SALIDA-comprobar.md.

El banco compara total, primeros ids y suma. Como scriptc y Bun AOT salieron MÁS rápidos que Go en /trabajo, aquí se mira más:
  1. la respuesta ENTERA de /trabajo (sin «lang» ni «ns») para n = 1.000, 20.000 y 50.000, contra la de Go;
  2. cuánta CPU usa cada servidor con /trabajo a tope (docker stats durante hey -c 16): ¿un hilo o más?
  3. la memoria tras 3.000 peticiones de /trabajo seguidas (¿la libera?).

⛔ Con el candado (pone carga). Usa las imágenes gl-880-*:1 del banco.

    python recursos/pruebas-go-vs-ts-nativo/comprobar.py
"""
import json
import re
import subprocess
import sys
import threading
import time
import urllib.request
from datetime import datetime
from pathlib import Path

AQUI = Path(__file__).resolve().parent
LOCK = Path(r"<proyecto>\_logs\banco.lock")
if not LOCK.is_dir():
    raise SystemExit("⛔ sin candado")
RED, SRV, PUERTO = "gl-880-red", "gl-880-srv", 9880
OUT = []


def sh(a):
    return subprocess.run(a, capture_output=True, text=True, encoding="utf-8", errors="replace")


def w(s=""):
    print(s, flush=True)
    OUT.append(s)


def arrancar(img):
    sh(["docker", "rm", "-f", SRV])
    r = sh(["docker", "run", "-d", "--name", SRV, "--network", RED, "--cpus", "2", "--memory", "1g", "--ulimit", "core=1",
            "-p", f"{PUERTO}:8080", img])
    assert r.returncode == 0, r.stderr
    for _ in range(300):
        try:
            urllib.request.urlopen(f"http://127.0.0.1:{PUERTO}/suma?n=1", timeout=1).read()
            return
        except Exception:
            time.sleep(0.05)
    raise SystemExit("no arrancó " + img)


def limpio(n):
    j = json.loads(urllib.request.urlopen(f"http://127.0.0.1:{PUERTO}/trabajo?n={n}", timeout=60).read())
    j.pop("lang"), j.pop("ns")
    return j


def stats():
    s = sh(["docker", "stats", "--no-stream", "--format", "{{.CPUPerc}}|{{.MemUsage}}", SRV]).stdout.strip()
    cpu, mem = s.split("|")
    return float(cpu.rstrip("%")), mem.split("/")[0].strip()


sh(["docker", "network", "rm", RED])
sh(["docker", "network", "create", "--subnet", "10.88.0.0/24", RED])
IMGS = [("go", "gl-880-go:1"), ("bun --compile", "gl-880-bun:1"), ("bun --compile --aot", "gl-880-bun-aot:1"),
        ("scriptc", "gl-880-scriptc:1")]
w("# SALIDA-comprobar — ¿de verdad hacen el mismo trabajo? (golandia 880)\n")
w(f"> Corrida del {datetime.now():%Y-%m-%d %H:%M}, con el candado tomado. Servidores con 2 CPU y 1 GiB.\n")
w("## 1 · La respuesta entera de /trabajo (sin lang ni ns), contra Go\n")
ref = {}
for nombre, img in IMGS:
    arrancar(img)
    fila = []
    for n in (1000, 20000, 50000):
        j = limpio(n)
        if nombre == "go":
            ref[n] = j
        fila.append(f"n={n}: {len(json.dumps(j))} caracteres, {'igual que Go' if j == ref[n] else '⛔ DISTINTO'}")
    w(f"- {nombre}: " + " · ".join(fila))
sh(["docker", "rm", "-f", SRV])
w()

w("## 2 · CPU y memoria con /trabajo a tope (hey -c 16, 20 s; docker stats cada ~2 s)\n")
w("| variante | pet/s | CPU máx. (200 % = los 2 núcleos) | CPU mediana | memoria al final |")
w("|---|---|---|---|---|")
for nombre, img in IMGS:
    arrancar(img)
    cpus, fin = [], threading.Event()

    def muestrear():
        while not fin.is_set():
            try:
                cpus.append(stats()[0])
            except Exception:
                pass

    th = threading.Thread(target=muestrear)
    th.start()
    r = sh(["docker", "run", "--rm", "--network", RED, "--cpus", "4", "gl-880-hey:1", "-z", "20s", "-c", "16",
            f"http://{SRV}:8080/trabajo?n=20000"])
    fin.set()
    th.join()
    rps = float(re.search(r"Requests/sec:\s+([\d.]+)", r.stdout).group(1))
    c = sorted(cpus[1:]) or [0]
    w(f"| {nombre} | {rps:.0f} | {max(c):.0f} % | {c[len(c) // 2]:.0f} % | {stats()[1]} |")
    sh(["docker", "rm", "-f", SRV])
w()

w("## 3 · Memoria tras 3.000 peticiones seguidas de /trabajo (de una en una)\n")
for nombre, img in IMGS:
    arrancar(img)
    antes = stats()[1]
    for _ in range(3000):
        urllib.request.urlopen(f"http://127.0.0.1:{PUERTO}/trabajo?n=20000", timeout=60).read()
    time.sleep(1)
    w(f"- {nombre}: en reposo {antes} → tras 3.000 peticiones {stats()[1]}")
    sh(["docker", "rm", "-f", SRV])
w()
sh(["docker", "network", "rm", RED])
(AQUI / "SALIDA-comprobar.md").write_text("\n".join(OUT) + "\n", encoding="utf-8")
