# -*- coding: utf-8 -*-
"""demos.py — lo que NO es cronómetro en «¿Go o TypeScript compilado a nativo?» (golandia 880) → SALIDA-demos.md.

1. La cobertura: once programas pequeños con cosas normales de TypeScript (eval, new Function, any, Proxy, parchear un
   prototipo, genérico interno, import() dinámico, worker_threads, == con objeto, Express y Hono) compilados con
   scriptc 0.4.14, bun --compile --aot (canary) y bun --compile (1.4.4), y ejecutados. Lo que sale, tal cual.
2. scriptc coverage del servidor de la serie.
3. Bun 1.4.4 estable con --aot: ¿qué hace?
4. ldd de cada binario.

Sin candado: no mide tiempos. Usa las imágenes de etapa gl-880-*:1-etapa (las crea el banco o `docker build --target build`).

    python recursos/pruebas-go-vs-ts-nativo/demos.py
"""
import subprocess
import sys
from datetime import datetime
from pathlib import Path

AQUI = Path(__file__).resolve().parent
OUT = []
PROGS = sorted(p.name for p in (AQUI / "cobertura").glob("*.ts"))


def w(s=""):
    print(s, flush=True)
    OUT.append(s)


def docker(imagen, script, entry="bash"):
    r = subprocess.run(["docker", "run", "--rm", "--ulimit", "core=1", "--cpus", "2", "--entrypoint", entry,
                        "-v", f"{AQUI / 'cobertura'}:/cob:ro", "-v", f"{AQUI / 'src'}:/src:ro", imagen, "-c", script],
                       capture_output=True, text=True, encoding="utf-8", errors="replace", timeout=900)
    return (r.stdout + r.stderr).rstrip()


def bloque(titulo, texto):
    w(f"### {titulo}\n")
    w("```")
    w(texto)
    w("```\n")


PREP_NPM = "mkdir -p /w && cp /cob/* /w/ && cd /w && "
w("# SALIDA-demos — ¿Go o TypeScript compilado a nativo? (golandia 880)\n")
w(f"> Corrida del {datetime.now():%Y-%m-%d %H:%M}. Sin cronómetro. Cada programa se compila y se ejecuta en un contenedor nuevo; "
  "la salida es la literal (recortada a las líneas que importan con `head`).\n")

w("## 1 · Cobertura: once programas, tres compiladores\n")
scr = PREP_NPM + "npm install --silent --no-audit --no-fund >/dev/null 2>&1; for f in " + " ".join(PROGS) + \
      "; do echo \"=== $f\"; scriptc build $f -o ./bin 2>&1 | head -6; echo \"[build: ${PIPESTATUS[0]}]\"; " \
      "if [ -x ./bin ]; then timeout 10 ./bin 2>&1 | head -4; echo \"[run: ${PIPESTATUS[0]}]\"; rm -f ./bin; fi; done"
bloque("scriptc build + ejecutar (scriptc 0.4.14)", docker("gl-880-scriptc:1-etapa", scr))
for etiqueta, imagen, flag in (("bun build --compile --aot + ejecutar (Bun 1.4.5-canary.1)", "gl-880-bun-aot:1-etapa", "--aot"),
                               ("bun build --compile + ejecutar (Bun 1.4.4)", "gl-880-bun:1-etapa", "")):
    scr = PREP_NPM + "bun install --silent >/dev/null 2>&1; for f in " + " ".join(PROGS) + \
          f"; do echo \"=== $f\"; bun build --compile {flag} $f --outfile ./bin 2>&1 | grep -v -E 'bundle|compile  |build id|check ' | head -6; " \
          "echo \"[build: ${PIPESTATUS[0]}]\"; if [ -x ./bin ]; then timeout 10 ./bin 2>&1 | head -4; echo \"[run: ${PIPESTATUS[0]}]\"; " \
          "rm -f ./bin; fi; done"
    bloque(etiqueta, docker(imagen, scr))

w("## 1b · Express con scriptc, instalando lo que pide el error (el paquete libnode de la misma versión)\n")
bloque("npm install @scriptc/libnode-linux-x64-gnu@0.4.14 + scriptc build 10_express.ts", docker(
    "gl-880-scriptc:1-etapa", PREP_NPM + "npm install --silent --no-audit --no-fund >/dev/null 2>&1; "
    "npm install --silent --no-audit --no-fund @scriptc/libnode-linux-x64-gnu@0.4.14 >/dev/null 2>&1; "
    "du -sh node_modules/@scriptc/libnode-linux-x64-gnu | cut -f1; scriptc build 10_express.ts -o bin 2>&1 | grep -E 'SC8001|^/w/bin' | head -4; "
    "echo \"[build: ${PIPESTATUS[0]}]\"; stat -c '%n %s' bin; timeout 10 ./bin 2>&1 | head -3; echo \"[run: ${PIPESTATUS[0]}]\""))

w("## 2 · scriptc coverage del servidor de la serie\n")
bloque("scriptc coverage servidor.ts", docker("gl-880-scriptc:1-etapa", "cd /src && scriptc coverage servidor.ts"))
bloque("scriptc coverage 08_worker.ts", docker("gl-880-scriptc:1-etapa", "cd /cob && scriptc coverage 08_worker.ts 2>&1 | head -20"))

w("## 3 · Bun 1.4.4 (estable) con --aot\n")
bloque("bun build --compile --aot en 1.4.4", docker(
    "gl-880-bun:1-etapa", "cd /tmp && cp /src/servidor.ts . && bun --revision && bun build --help | grep -c -- '--aot'; "
    "bun build --compile --aot servidor.ts --outfile a 2>&1; echo \"[codigo de salida: $?]\"; "
    "bun build --compile servidor.ts --outfile b >/dev/null 2>&1; stat -c '%n %s' a b"))
bloque("bun build --compile --aot en 1.4.5-canary.1", docker(
    "gl-880-bun-aot:1-etapa", "cd /tmp && cp /src/servidor.ts . && bun --revision && bun build --compile --aot servidor.ts --outfile a 2>&1; "
    "bun build --compile servidor.ts --outfile b >/dev/null 2>&1; stat -c '%n %s' a b"))

w("## 4 · ¿De qué depende cada binario? (ldd)\n")
bloque("ldd", docker("gl-880-scriptc:1-etapa", "cd /tmp && cp /src/servidor.ts . && scriptc build servidor.ts -o s >/dev/null 2>&1 && "
                     "echo '$ ldd scriptc' && ldd s"))
bloque("ldd bun", docker("gl-880-bun-aot:1-etapa", "cd /tmp && cp /src/servidor.ts . && bun build --compile --aot servidor.ts --outfile a "
                         ">/dev/null 2>&1 && echo '$ ldd bun-aot' && ldd a"))

(AQUI / "SALIDA-demos.md").write_text("\n".join(OUT) + "\n", encoding="utf-8")
print("[ok]", file=sys.stderr)
