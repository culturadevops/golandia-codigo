# -*- coding: utf-8 -*-
"""Genera las muestras de IA del vídeo 710 (golandia): los MISMOS 4 prompts a dos IA.

- Ollama qwen2.5:7b (local): 10 por lenguaje, temperature 0.8, seed 1..10 → se puede repetir.
- Claude Code (claude.exe -p): 3 por lenguaje, en un directorio temporal vacío, con --safe-mode
  (sin CLAUDE.md, hooks, MCP ni skills), --tools "" (sin herramientas) y sin variables de credenciales.
  (--bare existe, pero solo acepta ANTHROPIC_API_KEY; la cuenta del dueño es OAuth.)

Criterio fijado ANTES de mirar resultados: ronda 0, sin reparación. Un archivo «tiene hallazgo» si su
verificador imprime al menos un diagnóstico (también un error de compilación). Se cuenta todo.
Uso:  python generar_ia.py [--solo ollama|claude]
"""
import json
import os
import re
import subprocess
import sys
import tempfile
import urllib.request
from pathlib import Path

AQUI = Path(__file__).resolve().parent
IA = AQUI / "ia"
CLAUDE = r"C:\Users\tu-usuario\.vscode\extensions\anthropic.claude-code-2.1.312-win32-x64\resources\native-binary\claude.exe"
MODELO = "qwen2.5:7b"
FIN = " Responde solo con el código, en un único bloque de código."

PROMPTS = {
    "go": ("go", "Escribe un programa en Go (un solo archivo, package main, solo biblioteca estándar) que procese una lista "
                 "de tareas en paralelo con goroutines, lleve la cuenta de las terminadas en un contador protegido con "
                 "sync.Mutex, y devuelva los fallos envueltos con fmt.Errorf y %w usando un tipo de error propio." + FIN),
    "python": ("py", "Escribe un módulo de Python 3.13 (un solo archivo, solo biblioteca estándar) con anotaciones de tipo en "
                     "todas las funciones, que lea un archivo CSV de productos y precios y calcule el promedio, la mediana y el "
                     "producto más caro." + FIN),
    "ts": ("ts", "Escribe un archivo TypeScript (Node 22, sin dependencias externas) que descargue una lista de usuarios con "
                 "fetch desde una API JSON, filtre los que tienen rol admin y los guarde en un archivo con una función async." + FIN),
    "bash": ("sh", "Escribe un script de Bash que reciba un directorio de origen y uno de destino, comprima con gzip los "
                   "archivos .log del origen modificados hace más de 7 días y los mueva al destino." + FIN),
}


def extraer(texto):
    m = re.findall(r"```[a-zA-Z0-9_+-]*\n(.*?)```", texto, re.S)
    return (max(m, key=len) if m else texto).strip() + "\n"


def ollama(prompt, seed):
    cuerpo = {"model": MODELO, "prompt": prompt, "stream": False, "options": {"temperature": 0.8, "seed": seed}}
    req = urllib.request.Request("http://127.0.0.1:11434/api/generate", json.dumps(cuerpo).encode(),
                                 {"Content-Type": "application/json"})
    with urllib.request.urlopen(req, timeout=600) as r:
        return json.loads(r.read())["response"]


def entorno_limpio():
    malo = re.compile(r"AWS|KUBE|TOKEN|SECRET|PASSWORD|GITHUB|GH_|AZURE|GOOGLE|DOCKER_AUTH", re.I)
    return {k: v for k, v in os.environ.items() if not malo.search(k)}


def claude(prompt):
    with tempfile.TemporaryDirectory(prefix="v710-cc-") as tmp:
        r = subprocess.run([CLAUDE, "-p", prompt, "--safe-mode", "--tools", "", "--output-format", "json",
                            "--no-session-persistence"], cwd=tmp, env=entorno_limpio(), capture_output=True,
                           text=True, encoding="utf-8", timeout=900)
    if r.returncode != 0:
        raise RuntimeError(r.stderr[-2000:] or r.stdout[-2000:])
    return json.loads(r.stdout)


def main():
    solo = sys.argv[sys.argv.index("--solo") + 1] if "--solo" in sys.argv else None
    meta = {"prompts": {k: v[1] for k, v in PROMPTS.items()}, "ollama": {}, "claude": {}}
    mp = IA / "meta.json"
    if mp.exists():
        meta.update(json.loads(mp.read_text(encoding="utf-8")))
    if solo in (None, "ollama"):
        tags = json.loads(urllib.request.urlopen("http://127.0.0.1:11434/api/tags").read())
        dig = next(m["digest"] for m in tags["models"] if m["name"] == MODELO)
        meta["ollama"] = {"modelo": MODELO, "digest": dig, "temperature": 0.8, "seeds": list(range(1, 11)),
                          "endpoint": "/api/generate"}
        for lang, (ext, p) in PROMPTS.items():
            for s in range(1, 11):
                d = IA / "qwen" / lang
                d.mkdir(parents=True, exist_ok=True)
                (d / f"s{s:02d}.{ext}").write_text(extraer(ollama(p, s)), encoding="utf-8", newline="\n")
                print("qwen", lang, s, flush=True)
    if solo in (None, "claude"):
        ver = subprocess.run([CLAUDE, "--version"], capture_output=True, text=True).stdout.strip()
        meta["claude"] = {"version": ver, "flags": "-p --safe-mode --tools \"\" --output-format json --no-session-persistence",
                          "respuestas": []}
        for lang, (ext, p) in PROMPTS.items():
            for i in range(1, 4):
                j = claude(p)
                d = IA / "claude" / lang
                d.mkdir(parents=True, exist_ok=True)
                (d / f"c{i:02d}.{ext}").write_text(extraer(j.get("result", "")), encoding="utf-8", newline="\n")
                meta["claude"]["respuestas"].append({"lang": lang, "n": i, "modelos": list((j.get("modelUsage") or {}).keys()),
                                                     "costo_usd": j.get("total_cost_usd")})
                print("claude", lang, i, list((j.get("modelUsage") or {}).keys()), flush=True)
    mp.write_text(json.dumps(meta, ensure_ascii=False, indent=2), encoding="utf-8")


if __name__ == "__main__":
    main()
