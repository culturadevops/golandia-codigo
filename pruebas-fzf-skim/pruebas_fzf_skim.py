# -*- coding: utf-8 -*-
"""pruebas_fzf_skim.py — «Go contra Rust: herramientas populares, lo medí» · fzf (Go) contra skim (Rust).

Qué se mide, en modo filtro no interactivo (`--filter`), sobre la MISMA lista de rutas sintéticas (1, 3 y 5 millones;
recursos/pruebas-fzf-skim/generar_rutas.py, semilla fija):
  · ¿llegan al mismo resultado? (mismo conjunto de líneas; si no, se para);
  · tiempo hasta tener los resultados, RAM máxima y CPU usada (/usr/bin/time), mediana de 5, alternando el orden;
  · calidad: en qué puesto sale la ruta que buscabas, con dos formas de escribir la consulta.
Binarios oficiales de cada release (Dockerfile). Contenedor con --cpus 2 y --memory 1g, prefijo gl-grb-.

    python investigacion/pruebas_fzf_skim.py <corrida>   → recursos/pruebas-fzf-skim/SALIDA-corrida-<corrida>.md
"""
import json
import statistics
import sys
from datetime import datetime
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
import grb_comun as G  # noqa: E402

BASE = G.REC / "pruebas-fzf-skim"
IMG = "gl-grb-fzf:1"
VOL = "gl-grb-fzf-datos"
TAMANOS = [1000000, 3000000, 5000000]
CONSULTAS = ["authhandler", "src main.go", "usrctrl", "payments/invoices"]
REPS = 5
LIM = ["--cpus", "2", "--memory", "1g"]


def correr(args, timeout):
    return G.docker_run(LIM + ["-v", f"{VOL}:/datos:ro", "-v", f"{BASE}:/s:ro", IMG] + args, timeout=timeout)


def jsonl(r):
    return [json.loads(l) for l in r.stdout.splitlines() if l.startswith("{")]


def miles(n):
    return f"{n:,}".replace(",", ".")


def main():
    corrida = sys.argv[1]
    S = G.Salida(BASE / f"SALIDA-corrida-{corrida}.md")
    S.w(f"# fzf (Go) contra skim (Rust) · corrida {corrida}\n")
    S.w(f"> Corrida del {datetime.now():%Y-%m-%d %H:%M}. Cada bloque, con el candado `_logs/banco.lock` tomado. "
        f"Regenerable con `python investigacion/pruebas_fzf_skim.py {corrida}`.\n")
    S.w(f"- Máquina (docker info): {G.host()}")
    S.w(f"- Contenedor: `{' '.join(LIM)}`, imagen {IMG} (debian:trixie-slim + binarios oficiales con su suma)\n")

    with G.candado("fzf-skim: versiones"):
        r = correr(["sh", "-c", "fzf --version; sk --version; ls -l /usr/local/bin/fzf /usr/local/bin/sk; "
                    "file /usr/local/bin/fzf 2>/dev/null; nproc; sha256sum /datos/*.txt; wc -l /datos/*.txt"], 300)
    S.w("## 0 · Versiones, binarios y listas\n")
    S.w("```\n" + r.stdout.strip() + "\n```\n")

    # ── A · ¿el mismo trabajo? ────────────────────────────────────────────────────────────────────────────────────
    S.w("## A · ¿Devuelven lo mismo? (mismo conjunto de líneas, 1 millón)\n")
    with G.candado("fzf-skim: iguales"):
        r = correr(["python3", "/s/medir_filtro.py", "iguales", "/datos/rutas-1000000.txt"] + CONSULTAS, 900)
    distinto = False
    for d in jsonl(r):
        res = d["res"]
        S.w(f"- «{d['q']}»: " + " · ".join(f"{h} {miles(v['n'])} líneas (conjunto {v['conjunto']})" for h, v in res.items()))
        if res["fzf"]["conjunto"] != res["sk"]["conjunto"]:
            distinto = True
            S.w("  - ⛔ fzf y sk (por defecto) NO devuelven el mismo conjunto")
        S.w(f"  - top 3 fzf: {res['fzf']['top10'][:3]}")
        S.w(f"  - top 3 sk:  {res['sk']['top10'][:3]}")
        S.w(f"  - top 10 en común fzf∩sk: {len(set(res['fzf']['top10']) & set(res['sk']['top10']))} de 10")
    S.w()
    if distinto:
        S.w("⛔ Los conjuntos no coinciden: los tiempos no serían del mismo trabajo. Paro.")
        sys.exit(1)

    # ── B · tiempos ──────────────────────────────────────────────────────────────────────────────────────────────
    S.w(f"## B · Tiempo hasta los resultados, RAM y CPU (mediana de {REPS}; --filter, salida a archivo)\n")
    S.w("| líneas | consulta | herramienta | coincidencias | s (mediana) | s (mín–máx) | RAM máx. MiB | CPU s (u+s) |")
    S.w("|---|---|---|---|---|---|---|---|")
    for n in TAMANOS:
        with G.candado(f"fzf-skim: tiempos {n}"):
            try:
                r = correr(["python3", "/s/medir_filtro.py", "tiempos", f"/datos/rutas-{n}.txt", str(REPS)] + CONSULTAS, 2400)
            except RuntimeError as e:
                S.w(f"\n⛔ {n}: {str(e)[:800]}\n")
                continue
        filas = jsonl(r)
        for q in CONSULTAS:
            for h in ("fzf", "sk", "sk-skim_v2"):
                xs = [f for f in filas if f["q"] == q and f["herr"] == h]
                if not xs:
                    continue
                rcs = {f["rc"] for f in xs}
                seg = [f["seg"] for f in xs]
                S.w(f"| {miles(n)} | {q} | {h} | rc {sorted(rcs)} | {statistics.median(seg):.2f} | {min(seg):.2f}–{max(seg):.2f} | "
                    f"{statistics.median([f['rss_mib'] for f in xs]):.1f} | {statistics.median([f['cpu'] for f in xs]):.2f} |")
    S.w()

    # ── C · calidad ──────────────────────────────────────────────────────────────────────────────────────────────
    S.w("## C · ¿En qué puesto sale la ruta que buscabas? (1 millón, 30 rutas al azar, semilla 7)\n")
    S.w("Consulta «fragmentos»: `proyecto + 4 letras de la última carpeta + nombre del archivo`. "
        "Consulta «abreviado»: `3 letras del proyecto + 3 de la carpeta + nombre sin extensión`, sin espacios.\n")
    with G.candado("fzf-skim: calidad"):
        r = correr(["python3", "/s/medir_filtro.py", "calidad", "/datos/rutas-1000000.txt", "30"], 2400)
    filas = jsonl(r)
    for estilo in ("fragmentos", "abreviado"):
        xs = [f for f in filas if f["estilo"] == estilo]
        S.w(f"### {estilo} ({len(xs)} búsquedas)\n")
        S.w("| herramienta | 1.º | top 10 | top 100 | no aparece | mediana del puesto |")
        S.w("|---|---|---|---|---|---|")
        for h in ("fzf", "sk", "sk-skim_v2"):
            ps = [f[h]["puesto"] for f in xs]
            ok = [p for p in ps if p is not None]
            S.w(f"| {h} | {sum(1 for p in ok if p == 1)} | {sum(1 for p in ok if p <= 10)} | {sum(1 for p in ok if p <= 100)} | "
                f"{sum(1 for p in ps if p is None)} | {statistics.median(ok) if ok else '-'} |")
        S.w()
    S.w("Detalle (puesto fzf · sk · sk-skim_v2, y coincidencias de cada uno):\n")
    S.w("```")
    for f in filas:
        S.w(f"{f['estilo'][:4]} {f['q']!r:48} → {f['fzf']['puesto']} · {f['sk']['puesto']} · {f['sk-skim_v2']['puesto']}   "
            f"(n {f['fzf']['n']} · {f['sk']['n']} · {f['sk-skim_v2']['n']})   objetivo {f['objetivo']}")
    S.w("```\n")
    S.w("FIN")


if __name__ == "__main__":
    main()
