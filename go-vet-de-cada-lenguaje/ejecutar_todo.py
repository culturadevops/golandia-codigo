# -*- coding: utf-8 -*-
"""Vídeo 710 (golandia): corre cada verificador con versión fijada en Docker y regenera SALIDA.md.

Casos a mano (casos/) + muestras de IA (ia/, las genera generar_ia.py). Contenedores con --rm,
nombres v710-*, carpeta montada de solo lectura, como mucho uno a la vez.
Uso:  python ejecutar_todo.py
"""
import datetime as dt
import json
import re
import subprocess
from collections import Counter, defaultdict
from pathlib import Path

AQUI = Path(__file__).resolve().parent
IMG = {"go": "golang:1.27.2", "python": "python:3.13.18-slim", "ts": "node:22.23.4-slim",
       "bash": "koalaman/shellcheck-alpine:v0.11.0", "bash_ejecuta": "python:3.13.18-slim"}
SHELL = {"bash_ejecuta": "bash"}
SUPR = re.compile(r"nolint|type:\s*ignore|shellcheck\s+disable|eslint-disable|@ts-ignore|@ts-expect-error|noqa")


def docker(nombre):
    w = str(AQUI).replace("\\", "/")
    cmd = ["docker", "run", "--rm", "--name", f"v710-{nombre.replace('_', '-')}", "-v", f"{w}:/w:ro",
           IMG[nombre], SHELL.get(nombre, "sh"), f"/w/scripts/{nombre}.sh"]
    print("·", " ".join(cmd), flush=True)
    r = subprocess.run(cmd, capture_output=True, text=True, encoding="utf-8", errors="replace", timeout=3600)
    return r.stdout + r.stderr


def bloques(texto):
    cab, out, cur = [], {}, None
    for ln in texto.splitlines():
        if ln.startswith("=== "):
            cur = tuple(ln[4:].split("|"))
            out[cur] = []
        elif cur is None:
            cab.append(ln)
        else:
            out[cur].append(ln)
    return cab, out


def codigos(herr, lineas):
    txt = "\n".join(lineas)
    if herr == "eslint":
        try:
            j = json.loads(txt[txt.index("["):txt.rindex("]") + 1])
            return [m.get("ruleId") or "parse" for f in j for m in f["messages"]]
        except ValueError:
            return ["eslint-sin-json"]
    pat = {"ruff": r":\d+:\d+: ([A-Z]+\d+)", "mypy": r":\d+: error: .*\[([a-z-]+)\]",
           "tsc": r"error (TS\d+)", "shellcheck": r"\[(SC\d+)\]", "staticcheck": r":\d+:\d+: .*\(([A-Z]+\d+)\)$"}
    if herr in pat:
        return re.findall(pat[herr], txt, re.M)
    if herr == "vet":
        cs = []
        for ln in lineas:
            if not re.search(r"\.go:\d+:\d+: ", ln):
                continue
            m = ln.lower()
            # «vet: ./main.go:…» = error de tipos/compilación: vet no llega a correr sus análisis en ese paquete.
            cs.append("no-compila" if ln.startswith("vet: ") else "copylocks" if "lock" in m
                      else "printf" if ("%w" in m or "format" in m or "errorf" in m)
                      else "loopclosure" if "loop variable" in m else "otro")
        return cs
    return []


def ts7():
    """TypeScript 7.0.2 sobre el mismo caso, y `go version -m` sobre su binario. Volumen v710-ts7, que se borra al final."""
    w = str(AQUI).replace("\\", "/")
    sh = ("cd /t && cp /w/casos/ts_config/tsconfig.json /w/casos/ts/usuarios.ts . && "
          "npm i --silent --no-audit --no-fund typescript@7.0.2 @types/node@22.20.5 >/dev/null 2>&1; "
          "echo '$ npx tsc -v'; npx tsc -v; echo '$ npx tsc -p .'; npx tsc -p .; echo \"--- exit=$?\"")
    a = subprocess.run(["docker", "run", "--rm", "--name", "v710-ts7", "-v", f"{w}:/w:ro", "-v", "v710-ts7:/t",
                        IMG["ts"], "sh", "-c", sh], capture_output=True, text=True, encoding="utf-8", errors="replace")
    b = subprocess.run(["docker", "run", "--rm", "--name", "v710-ts7-go", "-v", "v710-ts7:/t:ro", IMG["go"], "sh", "-c",
                        "echo '$ go version -m node_modules/@typescript/typescript-linux-x64/lib/tsc'; "
                        "go version -m /t/node_modules/@typescript/typescript-linux-x64/lib/tsc | head -3"],
                       capture_output=True, text=True, encoding="utf-8", errors="replace")
    subprocess.run(["docker", "volume", "rm", "v710-ts7"], capture_output=True)
    return [l for l in (a.stdout + b.stdout).splitlines() if "npm notice" not in l]


def main():
    hoy = dt.datetime.now().strftime("%Y-%m-%d %H:%M")
    res = {n: bloques(docker(n)) for n in ("go", "python", "ts", "bash", "bash_ejecuta")}
    t7 = ts7()
    meta = json.loads((AQUI / "ia" / "meta.json").read_text(encoding="utf-8")) if (AQUI / "ia" / "meta.json").exists() else {}
    L = [f"# SALIDA — vídeo 710 (golandia) «go vet de cada lenguaje»", "",
         f"Generado por `python ejecutar_todo.py` el {hoy}. Todo corre en Docker con `--rm`; la carpeta va montada de solo lectura.", "",
         "## Versiones", ""]
    for n, (cab, _) in res.items():
        L += [f"- `{IMG[n]}`: " + " · ".join(c.strip() for c in cab if c.strip())]
    L += ["", "## A · Casos escritos a mano (el error típico de una IA, uno por lenguaje)", ""]
    for n, (_, bl) in res.items():
        for k, v in bl.items():
            if k[0] == "caso":
                L += [f"### {k[1]} · {k[2]}", "", "```", *v, "```", ""]
    L += ["### TypeScript 7.0.2 (el mismo usuarios.ts) y su binario", "", "```", *t7, "```", ""]
    # ── B · muestras de IA
    L += ["## B · Muestras de IA (ronda 0, sin reparar)", "",
          "Criterio fijado antes de mirar: un archivo «tiene hallazgo» si su verificador imprime al menos un diagnóstico "
          "(un error de compilación también cuenta). Se cuenta todo. Supresiones = líneas con nolint, type: ignore, "
          "shellcheck disable, eslint-disable, @ts-ignore/@ts-expect-error o noqa en el código generado.", "",
          "Las muestras de qwen se guardaron al principio con fin de línea CRLF (fallo de generar_ia.py en Windows, ya corregido) y se pasaron a LF con sed antes de verificar: el CRLF no es del modelo. ", "",
          "**No es un estudio:** 10 muestras por lenguaje de un modelo local y 3 de Claude Code.", ""]
    if meta:
        L += ["### Cómo se generaron", "", "```json", json.dumps(meta, ensure_ascii=False, indent=2), "```", ""]
    porarch = defaultdict(dict)
    for n, (_, bl) in res.items():
        for k, v in bl.items():
            if k[0] == "ia":
                porarch[k[1]][k[2]] = (codigos(k[2], v), v)
    tabla = defaultdict(lambda: {"n": 0, "con": 0, "diag": 0, "cod": Counter(), "supr": 0, "nocomp": 0, "porherr": Counter()})
    for rel, herrs in sorted(porarch.items()):
        _, fuente, lang, _ = rel.split("/")
        t = tabla[(lang, fuente)]
        t["n"] += 1
        tot = sum(len(c) for c, _ in herrs.values())
        t["con"] += tot > 0
        t["nocomp"] += any("no-compila" in c for c, _ in herrs.values())
        t["diag"] += tot
        for h, (c, _) in herrs.items():
            t["cod"].update(f"{h}:{x}" for x in c)
            t["porherr"][h] += len(c) > 0
        src = (AQUI / rel).read_text(encoding="utf-8", errors="replace")
        t["supr"] += len(SUPR.findall(src))
    L += ["### Resumen", "", "| lenguaje | IA | archivos | con hallazgo | por verificador (archivos) | Go sin compilar | diagnósticos | supresiones | códigos más vistos |",
          "|---|---|---|---|---|---|---|---|---|"]
    for (lang, fuente), t in sorted(tabla.items()):
        ph = ", ".join(f"{h} {c}" for h, c in sorted(t["porherr"].items()))
        top = ", ".join(f"{c}×{k}" for c, k in t["cod"].most_common(6))
        L += [f"| {lang} | {fuente} | {t['n']} | {t['con']} | {ph} | {t['nocomp']} | {t['diag']} | {t['supr']} | {top} |"]
    L += ["", "### Salida literal por archivo", ""]
    for rel, herrs in sorted(porarch.items()):
        for h, (c, v) in herrs.items():
            L += [f"#### {rel} · {h} ({len(c)})", "", "```", *v[:60], "```", ""]
    (AQUI / "SALIDA.md").write_text("\n".join(L) + "\n", encoding="utf-8")
    (AQUI / "resumen_ia.json").write_text(json.dumps({f"{a}|{b}": {**t, "cod": dict(t["cod"]), "porherr": dict(t["porherr"])}
                                                       for (a, b), t in tabla.items()}, ensure_ascii=False, indent=2), encoding="utf-8")
    print("SALIDA.md escrita")


if __name__ == "__main__":
    main()
