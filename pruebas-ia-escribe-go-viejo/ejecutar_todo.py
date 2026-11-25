# -*- coding: utf-8 -*-
"""Vídeo 703 «La IA escribe Go viejo». Regenera SALIDA.md.

  python ejecutar_todo.py              → mide (medir.py, Docker golang:1.27.2) y escribe SALIDA.md
  python ejecutar_todo.py --generar    → antes, pide las respuestas que falten (reanudable) con
                                         ../ia-generador-go/generar.py (el mismo generador del 705)
  python ejecutar_todo.py --solo-resumen → solo rehace SALIDA.md desde resultados.json
"""
import collections
import datetime
import glob
import json
import os
import re
import subprocess
import sys

AQUI = os.path.dirname(os.path.abspath(__file__))
GEN = os.path.join(AQUI, "..", "ia-generador-go", "generar.py")
LOCAL = "qwen2.5-coder:7b"

# Conteo descriptivo (no es el criterio principal): ¿usa la respuesta el modismo nuevo de su tarea?
# Se añadió después de ver que go fix no marca todo lo viejo (p. ej. «t := 30; &t» o errors.As con && x != nil).
USA = {"t01": r"range n\b", "t02": r"range n\b", "t03": r"\bm(in|ax)\(", "t04": r"\bmax\(", "t05": r"\bany\b",
       "t06": r"\bany\b", "t07": r"strings\.Cut\(", "t08": r"strings\.Cut\(", "t09": r"CutPrefix", "t10": r"FieldsSeq",
       "t11": r"SplitSeq", "t12": r"strings\.Builder|strings\.Join", "t13": r"maps\.Copy", "t14": r"maps\.Clone",
       "t15": r"slices\.Contains", "t16": r"slices\.ContainsFunc", "t17": r"slices\.Sort", "t18": r"slices\.Backward|slices\.Reverse",
       "t19": r"errors\.AsType", "t20": r"errors\.AsType", "t21": r"new\(\s*(30|\"api\")", "t22": r"wg\.Go\(",
       "t23": r"TypeFor", "t24": r"atomic\.Int64"}


def generar():
    base = ["--tareas", os.path.join(AQUI, "tareas.json"), "--prompts", os.path.join(AQUI, "prompts.json"),
            "--salida", os.path.join(AQUI, "respuestas")]
    subprocess.run([sys.executable, GEN, *base, "--backend", "ollama", "--modelo", LOCAL, "--variantes", "A,B,C",
                    "--seeds", "1,2,3"], check=True)
    subprocess.run([sys.executable, GEN, *base, "--backend", "claude", "--variantes", "A,B", "--solo-claude"], check=True)


def pct(a, b):
    return "%d/%d (%d %%)" % (a, b, round(100 * a / b)) if b else "0/0"


def resumen():
    R = json.load(open(os.path.join(AQUI, "resultados.json"), encoding="utf-8"))
    L = []
    w = L.append
    metas = [json.load(open(p, encoding="utf-8")) for p in glob.glob(os.path.join(AQUI, "respuestas", "*", "*", "*", "*", "meta.json"))]
    loc = next(m for m in metas if m["backend"] == "ollama")
    cl = [m for m in metas if m["backend"] == "claude"]
    w("# SALIDA — 703 «La IA escribe Go viejo: lo medí con go fix»\n")
    w("> Regenerado el %s con `python ejecutar_todo.py`. Medido en `golang:1.27.2` (Docker, --cpus 2), go.mod `go 1.27`." %
      datetime.datetime.now().strftime("%Y-%m-%d %H:%M"))
    w("> No es un estudio: 24 tareas propias, una sola máquina, un modelo local y una muestra pequeña de Claude Code.\n")
    w("## Quién generó")
    w("- Local: Ollama `%s`, digest `%s`, `/api/generate`, temperature %s, seeds 1-3, num_predict %s." %
      (loc["modelo"], loc["digest"], loc["opciones"]["temperature"], loc["opciones"]["num_predict"]))
    mods = collections.Counter(k for m in cl for k in m["modelos"])
    w("- Claude Code 2.1.312, `claude -p` desde un directorio temporal vacío, flags `%s`. Modelos en modelUsage: %s. "
      "Costo total de las %d llamadas: %.2f USD. Sin seed: una muestra por tarea y variante." %
      (" ".join(cl[0]["flags"]), dict(mods), len(cl), sum(m["costo_usd"] or 0 for m in cl)))
    w("- `--bare` no se pudo usar: solo acepta ANTHROPIC_API_KEY y la cuenta entra por OAuth (respuesta: «Not logged in · Please run /login»). "
      "Se usó `--safe-mode` (sin hooks, MCP, CLAUDE.md ni plugins) y `--tools \"\"`.")
    w("- Prompts exactos: prompts.json (A sin versión · B añade «Usa Go 1.27.» · C añade además la lista de modismos, solo local). Tareas: tareas.json.\n")

    w("## Controles del banco")
    for r in R:
        if r["modelo"].startswith("control"):
            w("- %s: compila=%s vet=%s avisos de go fix=%d %s" % (r["modelo"], r["compila"], r["vet_ok"], sum(r["fix"].values()), dict(r["fix"])))
    w("")
    w("## Tabla principal (por modelo y prompt)")
    w("| modelo | prompt | muestras | compila | go vet limpio* | con ≥1 aviso de go fix* | avisos* | usa el modismo nuevo de su tarea* |")
    w("|---|---|---|---|---|---|---|---|")
    g = collections.defaultdict(list)
    for r in R:
        if not r["modelo"].startswith("control"):
            g[(r["modelo"], r["variante"])].append(r)
    for (m, v), rs in sorted(g.items()):
        c = [r for r in rs if r["compila"]]
        usa = sum(bool(re.search(USA[r["tarea"][:3]], open(os.path.join(AQUI, "respuestas", r["muestra"], "tarea.go"), encoding="utf-8").read())) for r in c)
        w("| %s | %s | %d | %s | %s | %s | %d | %s |" % (m, v, len(rs), pct(len(c), len(rs)), pct(sum(r["vet_ok"] for r in c), len(c)),
                                                        pct(sum(sum(r["fix"].values()) > 0 for r in c), len(c)),
                                                        sum(sum(r["fix"].values()) for r in c), pct(usa, len(c))))
    w("\n*sobre las que compilan.\n")
    w("## Avisos de go fix por modernizer (solo las que compilan)")
    for (m, v), rs in sorted(g.items()):
        an = collections.Counter()
        for r in rs:
            if r["compila"]:
                an.update(r["fix"])
        w("- %s %s: %s" % (m, v, dict(an.most_common()) or "ninguno"))
    w("\n## ¿Cambió algo la frase «Usa Go 1.27»? (mismo modelo, misma tarea, misma seed)")
    for m in ("qwen2.5-coder_7b", "claude-code"):
        pares = [(r, s) for r in g.get((m, "A"), []) for s in g.get((m, "B"), []) if r["tarea"] == s["tarea"] and r["seed"] == s["seed"]]
        iguales = sum(open(os.path.join(AQUI, "respuestas", a["muestra"], "tarea.go"), "rb").read() ==
                      open(os.path.join(AQUI, "respuestas", b["muestra"], "tarea.go"), "rb").read() for a, b in pares)
        w("- %s: %d de %d respuestas B idénticas byte a byte a su A." % (m, iguales, len(pares)))
    w("\n## Por qué no compilan (primer error)")
    for (m, v), rs in sorted(g.items()):
        c = collections.Counter()
        for r in rs:
            if not r["compila"]:
                ls = [l for l in r["build"].splitlines() if l.startswith("./")]
                e = re.sub(r"^\./tarea\.go:\d+:\d+: ", "", ls[0]) if ls else "sin código"
                c[re.sub(r'"[^"]*"', '"…"', e)[:90]] += 1
        if c:
            w("- %s %s: %s" % (m, v, "; ".join("%d× %s" % (n, k) for k, n in c.most_common(6))))
    imp = collections.Counter()
    for r in g.get(("qwen2.5-coder_7b", "C"), []):
        if not r["compila"]:
            imp.update(re.findall(r'"([a-z/]+)" imported and not used', r["build"]))
    w("- Imports sin usar en el prompt C (local), por paquete: %s" % dict(imp.most_common()))
    w("- `go tool fix help` en golang:1.27.2: %d analizadores (go-tool-fix-help.txt). `go help fix`: «exit with a non-zero status if the diff is not empty» (go-help-fix.txt)." %
      sum(1 for l in open(os.path.join(AQUI, "go-tool-fix-help.txt"), encoding="utf-8") if re.match(r"^    [a-z]", l)))
    w("\n## go vet: lo que encontró")
    for r in R:
        if r["compila"] and not r["vet_ok"]:
            w("- %s: %s" % (r["muestra"], " | ".join(l for l in r["vet"].splitlines() if ".go:" in l)[:300]))
    w("\n## Por tarea: ¿usa el modismo nuevo? (A / B / C)")
    w("| tarea | modismo | Claude A | Claude B | local A | local B | local C |")
    w("|---|---|---|---|---|---|---|")
    T = json.load(open(os.path.join(AQUI, "tareas.json"), encoding="utf-8"))["tareas"]
    for t in T:
        celdas = []
        for m in ("claude-code", "qwen2.5-coder_7b"):
            for v in ("A", "B", "C"):
                rs = [r for r in g.get((m, v), []) if r["tarea"] == t["id"]]
                if not rs:
                    if m == "claude-code" and v == "C":
                        continue
                    celdas.append("—")
                    continue
                ok = sum(r["compila"] and bool(re.search(USA[t["id"][:3]], open(os.path.join(AQUI, "respuestas", r["muestra"], "tarea.go"), encoding="utf-8").read())) for r in rs)
                celdas.append("%d/%d" % (ok, len(rs)))
        w("| %s | %s | %s |" % (t["id"], t["modismo"], " | ".join(celdas)))
    w("\n## Diffs de go fix -diff de Claude Code (completos)")
    for r in R:
        if r["modelo"] == "claude-code" and r["fix"]:
            w("### %s %s\n```diff\n%s\n```" % (r["muestra"], dict(r["fix"]), r["fix_diff"].strip()))
    open(os.path.join(AQUI, "SALIDA.md"), "w", encoding="utf-8").write("\n".join(L) + "\n")
    print("SALIDA.md escrita")


if __name__ == "__main__":
    if "--generar" in sys.argv:
        generar()
    if "--solo-resumen" not in sys.argv:
        subprocess.run([sys.executable, os.path.join(AQUI, "medir.py")], check=True)
    resumen()
