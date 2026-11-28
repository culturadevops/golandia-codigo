# -*- coding: utf-8 -*-
"""pruebas_lazygit_gitui.py — «Go contra Rust: herramientas populares, lo medí» · lazygit (Go) contra gitui (Rust).

Son TUIs: no hay un «terminó» que cronometrar. Método (recursos/pruebas-lazygit-gitui/medir_tui.py): cada una arranca en
una pseudoterminal de 200×50, todo lo que pinta pasa por pyte (un emulador de terminal) y el cronómetro para cuando un
texto concreto APARECE en la pantalla emulada:
  · estado: «nuevo-00.txt», un archivo sin seguimiento del árbol de trabajo;
  · log: el asunto del último commit en la lista de commits (en gitui, tras pulsar «2» al ver el estado);
  · fondo: el asunto del PRIMER commit tras ir al final de la lista (lazygit «4» y «>», gitui la tecla Fin; la tecla se
    repite cada décima de segundo hasta verlo, porque el «>» de lazygit lleva al final de lo ya cargado).
Y la RAM de la TUI más la de sus procesos hijos (lazygit lanza git; gitui lee el repo dentro de su proceso).
Repos sintéticos con git fast-import (generar_repo.py), mismo repo para las dos. --cpus 2 y --memory 1g, prefijo gl-grb-.

    python investigacion/pruebas_lazygit_gitui.py <corrida>   → recursos/pruebas-lazygit-gitui/SALIDA-corrida-<corrida>.md
"""
import json
import statistics
import sys
from datetime import datetime
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
import grb_comun as G  # noqa: E402

BASE = G.REC / "pruebas-lazygit-gitui"
IMG = "gl-grb-git:1"
VOL = "gl-grb-git-datos"
REPOS = [(100000, 5000), (500000, 5000)]
REPS = 5
LIM = ["--cpus", "2", "--memory", "1g"]


def correr(args, timeout, check=True):
    return G.docker_run(LIM + ["-v", f"{VOL}:/datos:ro", "-v", f"{BASE}:/s:ro", IMG] + args, timeout=timeout, check=check)


def jsonl(r):
    return [json.loads(l) for l in r.stdout.splitlines() if l.startswith("{")]


def miles(n):
    return f"{n:,}".replace(",", ".")


def med(xs):
    xs = [x for x in xs if x is not None]
    return statistics.median(xs) if xs else None


def f3(x):
    return "-" if x is None else f"{x:.3f}"


def main():
    corrida = sys.argv[1]
    S = G.Salida(BASE / f"SALIDA-corrida-{corrida}.md")
    S.w(f"# lazygit (Go) contra gitui (Rust) · corrida {corrida}\n")
    S.w(f"> Corrida del {datetime.now():%Y-%m-%d %H:%M}. Cada bloque, con el candado `_logs/banco.lock` tomado. "
        f"Regenerable con `python investigacion/pruebas_lazygit_gitui.py {corrida}`.\n")
    S.w(f"- Máquina (docker info): {G.host()}")
    S.w(f"- Contenedor: `{' '.join(LIM)}`, imagen {IMG} (debian:trixie-slim + binarios oficiales + git + pyte)\n")
    with G.candado("lazygit-gitui: versiones"):
        r = correr(["sh", "-c", "lazygit --version; gitui --version; git --version; ls -l /usr/local/bin/lazygit /usr/local/bin/gitui; "
                    "cat /gitui-tar.sha256; for d in /datos/repo-*; do echo \"$d: $(git -C $d rev-list --count HEAD) commits, "
                    "$(git -C $d ls-files | wc -l) archivos, $(git -C $d status --porcelain | wc -l) cambios, .git $(du -sh $d/.git | cut -f1)\"; done"], 600)
    S.w("## 0 · Versiones, binarios y repos\n")
    S.w("```\n" + r.stdout.strip() + "\n```\n")

    S.w("## A · ¿Ven lo mismo? (primera pantalla de cada una, repo de 100.000)\n")
    S.w("Si alguna no llega a pintar el estado, el log o el primer commit, el banco para.\n")
    with G.candado("lazygit-gitui: iguales"):
        filas = []
        for h in ("lazygit", "gitui"):
            filas += jsonl(correr(["python3", "/s/medir_tui.py", h, "/datos/repo-100000", "1"], 900))
    for f in filas:
        S.w(f"- {f['herr']}: estado {f['estado']} s · log {f['log']} s · fondo {f['fondo']} s")
        S.w("  ```")
        for l in f["pantalla_fondo"]:
            S.w("  " + l[:160])
        S.w("  ```")
        if None in (f["estado"], f["log"], f["fondo"]):
            S.w("⛔ Una no llegó a pintar lo esperado. Paro.")
            sys.exit(1)
    S.w()

    S.w(f"## B · Arranque, log, ir al primer commit y RAM (mediana de {REPS}; cada repetición, proceso nuevo)\n")
    S.w("| commits | herramienta | estado s | log s | ir al 1.er commit s | RAM al ver el log MiB | RAM en reposo MiB | "
        "RAM tras ir al final MiB | VmHWM final MiB | procesos |")
    S.w("|---|---|---|---|---|---|---|---|---|---|")
    crudo = []
    for n, _ in REPOS:
        for h in ("lazygit", "gitui"):
            with G.candado(f"lazygit-gitui: {h} {n}"):
                r = correr(["python3", "/s/medir_tui.py", h, f"/datos/repo-{n}", str(REPS)], 3600, check=False)
            xs = jsonl(r)
            crudo += xs
            if r.returncode or not xs:
                S.w(f"| {miles(n)} | {h} | ⛔ rc {r.returncode} {r.stderr[-300:]!r} |")
                continue
            m = lambda k, sub=None: med([(x.get(k) or {}).get(sub) if sub else x.get(k) for x in xs])  # noqa: E731
            S.w(f"| {miles(n)} | {h} | {f3(m('estado'))} | {f3(m('log'))} | {f3(m('fondo'))} | {m('mem_log', 'arbol_MiB')} | "
                f"{m('mem_reposo', 'arbol_MiB')} | {m('mem_fondo', 'arbol_MiB')} | {m('mem_fondo', 'VmHWM')} | "
                f"{sorted({(x.get('mem_log') or {}).get('procesos') for x in xs})} |")
    S.w("\nCrudo (una línea por repetición):\n\n```")
    for x in crudo:
        x = dict(x)
        x.pop("pantalla_fondo", None)
        S.w(json.dumps(x, ensure_ascii=False))
    S.w("```\n")
    S.w("FIN")


if __name__ == "__main__":
    main()
