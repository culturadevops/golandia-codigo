# -*- coding: utf-8 -*-
"""Generador de respuestas de IA para los vídeos de golandia 703 y 705 (y los que lo reutilicen).

Un solo generador para los dos vídeos: lee un archivo de tareas y uno de plantillas de prompt,
pide la respuesta a la IA y guarda, por muestra, la respuesta cruda, el .go extraído y meta.json
(modelo exacto, digest, parámetros, duración, costo). Reanudable: si la muestra ya existe, la salta.

Backends:
  ollama  http://127.0.0.1:11434/api/generate, temperature y seed fijos (repetible).
  claude  Claude Code en modo no interactivo (-p), desde un directorio temporal vacío (mkdtemp),
          sin herramientas (--tools ""), --safe-mode (sin hooks, MCP, CLAUDE.md ni plugins),
          --strict-mcp-config, --permission-prompts none, sin sesión guardada, y sin credenciales
          de nube en el entorno. --bare no sirve: con --bare solo acepta ANTHROPIC_API_KEY y la
          cuenta del dueño entra por OAuth («Not logged in», probado el 2026-10-29).
          Nunca --dangerously-skip-permissions ni bypassPermissions.

Uso:
  python generar.py --tareas T.json --prompts P.json --backend ollama --modelo qwen2.5-coder:7b \
      --variantes A,B --seeds 1,2,3 --salida DIR
  python generar.py ... --backend claude --variantes A,B --solo-claude --salida DIR
"""
import argparse
import json
import os
import re
import shutil
import subprocess
import sys
import tempfile
import time
import urllib.request

OLLAMA = "http://127.0.0.1:11434"
CLAUDE = r"C:\Users\tu-usuario\.vscode\extensions\anthropic.claude-code-2.1.312-win32-x64\resources\native-binary\claude.exe"
TEMPERATURA = 0.7
NUM_PREDICT = 2048
# Variables que nunca deben llegar a la IA (canon: sin credenciales reales en el entorno).
FUERA = ("AWS_", "KUBECONFIG", "GITHUB_TOKEN", "GH_TOKEN", "AZURE_", "GOOGLE_", "ARM_", "DOCKER_", "TF_")


def prompt_de(plantillas, variante, tarea):
    p = plantillas["variantes"][variante]
    return p["plantilla"].replace("{TAREA}", tarea["enunciado"].strip())


def extraer_go(texto):
    bloques = re.findall(r"```(?:go|golang)?[ \t]*\r?\n(.*?)```", texto, flags=re.S)
    if bloques:
        return max(bloques, key=len)
    return texto if "package " in texto else ""


def ollama_digest(modelo):
    with urllib.request.urlopen(OLLAMA + "/api/tags", timeout=30) as r:
        for m in json.load(r)["models"]:
            if m["name"] == modelo:
                return m["digest"]
    return "?"


def pedir_ollama(modelo, prompt, seed):
    cuerpo = {"model": modelo, "prompt": prompt, "stream": False,
              "options": {"temperature": TEMPERATURA, "seed": seed, "num_predict": NUM_PREDICT}}
    req = urllib.request.Request(OLLAMA + "/api/generate", data=json.dumps(cuerpo).encode(),
                                 headers={"Content-Type": "application/json"})
    t0 = time.time()
    with urllib.request.urlopen(req, timeout=600) as r:
        d = json.load(r)
    return d["response"], {"segundos": round(time.time() - t0, 2), "eval_count": d.get("eval_count"),
                           "opciones": cuerpo["options"]}


def entorno_limpio():
    env = {k: v for k, v in os.environ.items() if not k.upper().startswith(FUERA)}
    return env


def pedir_claude(prompt, modelo=None):
    tmp = tempfile.mkdtemp(prefix="v70x-claude-")
    args = [CLAUDE, "-p", "--safe-mode", "--tools", "", "--strict-mcp-config", "--permission-prompts", "none",
            "--no-session-persistence", "--output-format", "json", "--max-budget-usd", "0.50"]
    if modelo:
        args += ["--model", modelo]
    t0 = time.time()
    try:
        r = subprocess.run(args, input=prompt, capture_output=True, text=True, encoding="utf-8", cwd=tmp,
                           env=entorno_limpio(), timeout=900)
    finally:
        shutil.rmtree(tmp, ignore_errors=True)
    d = json.loads(r.stdout)
    modelos = d.get("modelUsage", {})
    return d.get("result", ""), {"segundos": round(time.time() - t0, 2), "costo_usd": d.get("total_cost_usd"),
                                 "modelos": {k: v.get("outputTokens") for k, v in modelos.items()},
                                 "is_error": d.get("is_error"), "flags": args[1:]}


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--tareas", required=True)
    ap.add_argument("--prompts", required=True)
    ap.add_argument("--backend", choices=["ollama", "claude"], required=True)
    ap.add_argument("--modelo", default="qwen2.5-coder:7b")
    ap.add_argument("--variantes", default="A")
    ap.add_argument("--seeds", default="1")
    ap.add_argument("--solo-claude", action="store_true", help="solo las tareas marcadas claude=true")
    ap.add_argument("--salida", required=True)
    a = ap.parse_args()

    tareas = json.load(open(a.tareas, encoding="utf-8"))["tareas"]
    plantillas = json.load(open(a.prompts, encoding="utf-8"))
    if a.solo_claude:
        tareas = [t for t in tareas if t.get("claude")]
    seeds = [int(s) for s in a.seeds.split(",")] if a.backend == "ollama" else [1]
    nombre = a.modelo if a.backend == "ollama" else "claude-code"
    digest = ollama_digest(a.modelo) if a.backend == "ollama" else None
    hechos = 0
    for v in a.variantes.split(","):
        for t in tareas:
            for s in seeds:
                d = os.path.join(a.salida, nombre.replace(":", "_"), v, t["id"], "s%d" % s)
                if os.path.exists(os.path.join(d, "meta.json")):
                    continue
                os.makedirs(d, exist_ok=True)
                prompt = prompt_de(plantillas, v, t)
                if a.backend == "ollama":
                    texto, meta = pedir_ollama(a.modelo, prompt, s)
                    meta.update({"backend": "ollama", "modelo": a.modelo, "digest": digest})
                else:
                    texto, meta = pedir_claude(prompt)
                    meta.update({"backend": "claude", "claude_code": "2.1.312"})
                meta.update({"tarea": t["id"], "variante": v, "seed": s, "prompt": prompt})
                open(os.path.join(d, "respuesta.md"), "w", encoding="utf-8").write(texto)
                open(os.path.join(d, "tarea.go"), "w", encoding="utf-8", newline="\n").write(extraer_go(texto))
                json.dump(meta, open(os.path.join(d, "meta.json"), "w", encoding="utf-8"), ensure_ascii=False, indent=1)
                hechos += 1
                print(nombre, v, t["id"], s, meta["segundos"], "s", flush=True)
    print("nuevas:", hechos)


if __name__ == "__main__":
    sys.exit(main())
