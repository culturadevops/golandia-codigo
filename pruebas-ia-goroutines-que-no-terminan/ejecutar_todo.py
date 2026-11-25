# -*- coding: utf-8 -*-
"""Vídeo 705 «Goroutines que no terminan». Regenera SALIDA.md.

  python ejecutar_todo.py               → mide (medir.py: respuestas + controles + el caso) y corre las demos
  python ejecutar_todo.py --generar     → antes, pide las respuestas que falten (reanudable) con
                                          ../ia-generador-go/generar.py (el mismo generador del 703)
  python ejecutar_todo.py --solo-resumen → solo rehace SALIDA.md desde resultados.json y demos.txt
"""
import collections
import datetime
import glob
import json
import os
import subprocess
import sys

AQUI = os.path.dirname(os.path.abspath(__file__))
GEN = os.path.join(AQUI, "..", "ia-generador-go", "generar.py")
LOCAL = "qwen2.5-coder:7b"


def generar():
    base = ["--tareas", os.path.join(AQUI, "tareas.json"), "--prompts", os.path.join(AQUI, "prompts.json"),
            "--salida", os.path.join(AQUI, "respuestas")]
    subprocess.run([sys.executable, GEN, *base, "--backend", "ollama", "--modelo", LOCAL, "--variantes", "A,B",
                    "--seeds", "1,2,3,4"], check=True)
    subprocess.run([sys.executable, GEN, *base, "--backend", "claude", "--variantes", "A,B", "--solo-claude"], check=True)


def demos():
    r = subprocess.run(["docker", "run", "--rm", "--name", "v705-demos", "--cpus", "2", "-v", os.path.join(AQUI, "demos") + ":/d",
                        "-w", "/d", "-e", "GOTOOLCHAIN=local", "golang:1.27.2", "sh", "-c",
                        "echo '$ go run ./global'; go run ./global; echo; echo '$ go run ./http'; go run ./http"],
                       capture_output=True, text=True)
    open(os.path.join(AQUI, "demos.txt"), "w", encoding="utf-8").write(r.stdout + r.stderr)


def estado(r):
    """Una sola etiqueta por muestra, en este orden (fijado antes de mirar los resultados)."""
    if not r["compila"]:
        return "no compila"
    if "build failed" in r["test_txt"] and r["fugas"] is None:
        return "no cumple la firma"
    if r["fugas"] is None:
        return "panic" if "panic:" in r["test_txt"] else "la prueba falla sin resultado"
    if not r["termina"]:
        return "se cuelga"
    return "termina"


def resumen():
    R = json.load(open(os.path.join(AQUI, "resultados.json"), encoding="utf-8"))
    metas = [json.load(open(p, encoding="utf-8")) for p in glob.glob(os.path.join(AQUI, "respuestas", "*", "*", "*", "*", "meta.json"))]
    loc = next(m for m in metas if m["backend"] == "ollama")
    cl = [m for m in metas if m["backend"] == "claude"]
    L = []
    w = L.append
    w("# SALIDA — 705 «Goroutines que no terminan: el código concurrente de la IA con el perfil goroutineleak»\n")
    w("> Regenerado el %s con `python ejecutar_todo.py`. Medido en `golang:1.27.2` (Docker, --cpus 2), go.mod `go 1.27`." %
      datetime.datetime.now().strftime("%Y-%m-%d %H:%M"))
    w("> No es un estudio: 12 tareas propias con su prueba fija, una sola máquina, un modelo local y una muestra pequeña de Claude Code.\n")
    w("## Quién generó")
    w("- Local: Ollama `%s`, digest `%s`, `/api/generate`, temperature %s, seeds 1-4, num_predict %s." %
      (loc["modelo"], loc["digest"], loc["opciones"]["temperature"], loc["opciones"]["num_predict"]))
    mods = collections.Counter(k for m in cl for k in m["modelos"])
    w("- Claude Code 2.1.312, `claude -p` desde un directorio temporal vacío, flags `%s`. Modelos en modelUsage: %s. "
      "Costo de las %d llamadas: %.2f USD. Sin seed: una muestra por tarea y variante." %
      (" ".join(cl[0]["flags"]), dict(mods), len(cl), sum(m["costo_usd"] or 0 for m in cl)))
    w("- `--bare` no se pudo usar (solo ANTHROPIC_API_KEY; la cuenta entra por OAuth): se usó `--safe-mode` y `--tools \"\"`.")
    w("- Prompts: prompts.json (A sin aviso · B añade «Asegúrate de que no quede ninguna goroutine bloqueada cuando la función termine.»). Tareas: tareas.json. Pruebas: pruebas/.\n")
    w("## Cómo se activa el perfil (Go 1.27)")
    w("- go.dev/doc/go1.27 (fuente-go1.27.html): «previously available as an experiment in Go 1.26, is now generally available» · "
      "«supported in the runtime/pprof package» · «also available as the net/http/pprof endpoint /debug/pprof/goroutineleak» · "
      "«The goroutineleakprofile GOEXPERIMENT setting is now deleted». Límite: «may fail to identify leaks caused by blocking on "
      "concurrency primitives reachable through global variables or the local variables of runnable goroutines».")
    w("- En la prueba: `pprof.Lookup(\"goroutineleak\").WriteTo(&b, 1)` tras el escenario (pruebas/ayuda_test.go). Sin flags ni GOEXPERIMENT.\n")
    w("## Controles del banco")
    for r in R:
        if r["modelo"].startswith(("control", "caso")):
            w("- %s %s: compila=%s termina=%s correcto=%s fugas=%s races=%s vet=%s" %
              (r["modelo"], r["tarea"], r["compila"], r["termina"], r["correcto"], r["fugas"], r["races"], r["vet_ok"]))
    g = collections.defaultdict(list)
    for r in R:
        if not r["modelo"].startswith(("control", "caso")):
            g[(r["modelo"], r["variante"])].append(r)
    w("\n## Tabla principal (por modelo y prompt)")
    w("| modelo | prompt | muestras | compila | go vet limpio* | termina la prueba | correctas | con fuga | goroutines fugadas | con data race* | sin ningún problema |")
    w("|---|---|---|---|---|---|---|---|---|---|---|")
    for (m, v), rs in sorted(g.items()):
        c = [r for r in rs if r["compila"]]
        t = [r for r in rs if estado(r) in ("termina", "se cuelga")]
        limpias = [r for r in rs if estado(r) == "termina" and r["correcto"] and r["fugas"] == 0 and r["races"] == 0 and r["vet_ok"]]
        w("| %s | %s | %d | %d | %d/%d | %d | %d | %d | %d | %d/%d | %d/%d |" % (
            m, v, len(rs), len(c), sum(r["vet_ok"] for r in c), len(c), sum(estado(r) == "termina" for r in rs),
            sum(bool(r["correcto"]) for r in t), sum(r["fugas"] > 0 for r in t), sum(r["fugas"] for r in t),
            sum(r["races"] > 0 for r in c), len(c), len(limpias), len(rs)))
    w("\n*sobre las que compilan. «con fuga» y «goroutines fugadas» incluyen las que se cuelgan (su goroutine atascada también cuenta).\n")
    w("## Estado de cada muestra")
    for (m, v), rs in sorted(g.items()):
        w("- %s %s: %s" % (m, v, dict(collections.Counter(estado(r) for r in rs))))
    w("\n## Las que pasan su prueba y aun así fugan")
    for r in R:
        if r["modelo"] in [k[0] for k in g] and r["correcto"] and r["fugas"]:
            w("- %s: fugas=%d" % (r["muestra"], r["fugas"]))
    w("\n## Por tarea (local A+B: con fuga / con race / panic; Claude A+B: con problema)")
    w("| tarea | patrón | local con fuga | local con race | local panic | Claude con algún problema |")
    w("|---|---|---|---|---|---|")
    T = json.load(open(os.path.join(AQUI, "tareas.json"), encoding="utf-8"))["tareas"]
    for t in T:
        lo = [r for r in R if r["modelo"].startswith("qwen") and r["tarea"] == t["id"]]
        cc = [r for r in R if r["modelo"] == "claude-code" and r["tarea"] == t["id"]]
        mal = sum(not (estado(r) == "termina" and r["correcto"] and r["fugas"] == 0 and r["races"] == 0) for r in cc)
        w("| %s | %s | %d/%d | %d/%d | %d/%d | %d/%d |" % (t["id"], t["patron"], sum((r["fugas"] or 0) > 0 for r in lo), len(lo),
                                                          sum(r["races"] > 0 for r in lo), len(lo), sum(estado(r) == "panic" for r in lo),
                                                          len(lo), mal, len(cc)))
    w("\n## El caso del vídeo: c04 ElMasRapido, qwen2.5-coder:7b, prompt A, seed 1")
    for r in R:
        if r["modelo"].startswith("caso"):
            w("### %s (fugas=%s, correcto=%s)\n```\n%s\n```" % (r["modelo"], r["fugas"], r["correcto"], r["pilas"].strip() or "sin fugas"))
    w("Código: demos/caso/ia/tarea.go (copia literal de respuestas/qwen2.5-coder_7b/A/c04-el-mas-rapido/s1) y demos/caso/arreglado/tarea.go.\n")
    w("## Pilas de las fugas (local)")
    for r in R:
        if r["modelo"].startswith("qwen") and r["fugas"]:
            w("### %s (fugas=%d, termina=%s)\n```\n%s\n```" % (r["muestra"], r["fugas"], r["termina"], r["pilas"].strip()[:1200]))
    w("\n## Races (primera de cada muestra)")
    for r in R:
        if r["races"] and not r["modelo"].startswith("control"):
            w("### %s (WARNING: DATA RACE ×%d)\n```\n%s\n```" % (r["muestra"], r["races"], r["race_txt"][:900].strip()))
    p2 = os.path.join(AQUI, "resultados-corrida2.json")
    if os.path.exists(p2):
        b = {r["muestra"]: r for r in json.load(open(p2, encoding="utf-8"))}
        w("\n## Repetibilidad: segunda corrida del mismo banco (`python medir.py resultados-corrida2.json`)")
        w("Muestras cuyo resultado (termina, correcto, fugas, race) cambió entre corridas:")
        t = lambda r: (r["termina"], r["correcto"], r["fugas"], r["races"] > 0)
        cambios = [(k, t(r), t(b[k])) for k, r in ((r["muestra"], r) for r in R) if k in b and t(r) != t(b[k])]
        for k, x, y in cambios:
            w("- %s: corrida 1 %s → corrida 2 %s" % (k, x, y))
        w("- Total: %d de %d muestras cambiaron." % (len(cambios), len(R)))
        for (m, v), rs in sorted(g.items()):
            bs = [b[r["muestra"]] for r in rs if r["muestra"] in b]
            tt = [r for r in bs if r["fugas"] is not None]
            w("- corrida 2 · %s %s: con fuga %d, goroutines fugadas %d, con race %d" %
              (m, v, sum(r["fugas"] > 0 for r in tt), sum(r["fugas"] for r in tt), sum(r["races"] > 0 for r in bs if r["compila"])))
    w("\n## Demos")
    w("```\n%s\n```" % open(os.path.join(AQUI, "demos.txt"), encoding="utf-8").read().strip())
    open(os.path.join(AQUI, "SALIDA.md"), "w", encoding="utf-8").write("\n".join(L) + "\n")
    print("SALIDA.md escrita")


if __name__ == "__main__":
    if "--generar" in sys.argv:
        generar()
    if "--solo-resumen" not in sys.argv:
        subprocess.run([sys.executable, os.path.join(AQUI, "medir.py")], check=True)
        subprocess.run([sys.executable, os.path.join(AQUI, "medir.py"), "resultados-corrida2.json"], check=True)
        demos()
    resumen()
