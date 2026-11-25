# -*- coding: utf-8 -*-
"""pruebas_go_vs_caliente.py — la SEGUNDA compilación (caché caliente) de cada lenguaje de pruebas_go_vs.py: en el mismo
contenedor se compila una vez, se CAMBIA el fuente (una línea de comentario) y se mide la siguiente. Mediana de 3 contenedores.
Añade la sección «1b» a cada SALIDA-<comp>.md.

    python investigacion/pruebas_go_vs_caliente.py
"""
import statistics
import subprocess
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent))
import pruebas_go_vs as G  # noqa: E402

FUENTE = {"go": ("main.go", "//"), "csharp": ("Program.cs", "//"), "csharp-aot": ("Program.cs", "//"),
          "kotlin": ("Servidor.kt", "//"), "zig": ("main.zig", "//"), "elixir": ("lib/servidor.ex", "#")}


def caliente(lang):
    carpeta, imagen, orden, _, _ = G.L[lang]
    tiempos = []
    for _ in range(3):
        r = subprocess.run(["docker", "run", "--rm", "-v", f"{G.BASE / carpeta}:/src:ro", "-w", "/work", imagen, "sh", "-c",
                            f"cp -r /src/. /work/ && {orden} && echo '{FUENTE[lang][1]} cambio' >> {FUENTE[lang][0]} && t0=$(date +%s%N); {orden}; "
                            "t1=$(date +%s%N); echo COMPILA_MS $(( (t1 - t0) / 1000000 ))"],
                           capture_output=True, text=True, encoding="utf-8", errors="replace")
        ms = [int(x.split()[1]) for x in r.stdout.splitlines() if x.startswith("COMPILA_MS")]
        if r.returncode != 0 or not ms:
            raise SystemExit(f"[!] {lang}: {r.stdout[-1500:]} {r.stderr[-1500:]}")
        tiempos.append(ms[0])
    return tiempos


filas = {lang: caliente(lang) for lang in FUENTE}
for comp, langs in G.COMPARA.items():
    f = G.BASE / f"SALIDA-{comp}.md"
    li = ["", "## 1b · Segunda compilación (mismo contenedor, caché caliente, tras añadir una línea de comentario al fuente; mediana de 3)", "",
          "| lenguaje | corridas | mediana |", "|---|---|---|"]
    for lang in langs:
        li.append(f"| {G.L[lang][4]} | {filas[lang]} | **{statistics.median(filas[lang]):.0f} ms** |")
    f.write_text(f.read_text(encoding="utf-8") + "\n".join(li) + "\n", encoding="utf-8")
for lang, t in filas.items():
    print(f"{lang:11s} {t} → {statistics.median(t):.0f} ms")
