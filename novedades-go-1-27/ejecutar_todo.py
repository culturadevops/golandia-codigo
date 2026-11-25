# -*- coding: utf-8 -*-
"""Regenera SALIDA.md del vídeo 700 (golandia, «Lo nuevo de Go 1.27, ejecutado»).

    python canales/golandia/recursos/novedades-go-1-27/ejecutar_todo.py

Corre correr.sh en golang:1.27.2 y en golang:1.26 (contenedores con --rm, nombre v700-*, montaje de solo lectura).
Los bancos van uno detrás de otro, nunca a la vez, con --cpus 2; de cada benchmark se toma la mediana de 5 pasadas.
"""
import datetime
import re
import statistics
import subprocess
from pathlib import Path

AQUI = Path(__file__).resolve().parent
IMAGENES = {"127": "golang:1.27.2", "126": "golang:1.26"}


def docker(tag, modo):
    cmd = ["docker", "run", "--rm", "--cpus", "2", "--name", f"v700-{modo}-{tag}",
           "-v", f"{AQUI}:/trabajo:ro", IMAGENES[tag], "bash", "/trabajo/correr.sh", modo]
    r = subprocess.run(cmd, capture_output=True, text=True, encoding="utf-8", errors="replace")
    return r.stdout + r.stderr


BENCH = re.compile(r"^(Benchmark\S+?)(?:-\d+)?\s+\d+\s+([\d.]+) ns/op(?:\s+([\d.]+) MB/s)?(?:\s+(\d+) B/op\s+(\d+) allocs/op)?")


def medianas(texto):
    """{(bloque, benchmark): mediana ns/op}; bloque = 'def' o 'nosize' según el GOEXPERIMENT."""
    res, bloque = {}, "def"
    for linea in texto.splitlines():
        if "nosizespecializedmalloc" in linea:
            bloque = "nosize"
        if linea.startswith("## 10-flate"):
            bloque = "def"
        m = BENCH.match(linea)
        if m:
            res.setdefault((bloque, m.group(1)), []).append(float(m.group(2)))
    return {k: statistics.median(v) for k, v in res.items()}


def main():
    ahora = datetime.datetime.now().strftime("%Y-%m-%d %H:%M")
    partes = [f"# SALIDA — Lo nuevo de Go 1.27, ejecutado (golandia 700)\n",
              f"> Corrida del {ahora}. Contenedores golang:1.27.2 y golang:1.26 con --cpus 2. "
              "Notas oficiales: fuente-go1.27.html / fuente-go1.27.txt (go.dev/doc/go1.27, bajadas el 2026-10-29).\n"]
    for tag in ("127", "126"):
        partes += [f"\n# Demos · {IMAGENES[tag]}\n", "```", docker(tag, "demos").rstrip(), "```"]
    bancos = {}
    for tag in ("126", "127"):
        t = docker(tag, "banco")
        bancos[tag] = medianas(t)
        partes += [f"\n# Bancos · {IMAGENES[tag]} (salida cruda)\n", "```", t.rstrip(), "```"]

    partes.append("\n# Medianas de 5 pasadas (ns/op; flate en ns/op por 4 MB)\n")
    partes.append("| benchmark | Go 1.26 | Go 1.27 | Go 1.27 nosizespecializedmalloc | 1.27 vs 1.26 | 1.27 vs nosize |")
    partes.append("|---|---|---|---|---|---|")
    nombres = sorted({k[1] for k in bancos["127"]})
    for n in nombres:
        a = bancos["126"].get(("def", n))
        b = bancos["127"].get(("def", n))
        c = bancos["127"].get(("nosize", n))
        f = lambda x: "—" if x is None else f"{x:.2f}"
        d = lambda x, y: "—" if None in (x, y) else f"{(y - x) / x * 100:+.1f} %"
        partes.append(f"| {n} | {f(a)} | {f(b)} | {f(c)} | {d(a, b)} | {d(c, b)} |")
    (AQUI / "SALIDA.md").write_text("\n".join(partes) + "\n", encoding="utf-8")
    print("\n".join(partes[-len(nombres) - 3:]))


if __name__ == "__main__":
    main()
