# -*- coding: utf-8 -*-
"""demos.py — lo de «¿Go o Crystal?» (golandia 902) que no lleva cronómetro → SALIDA-demos.md

  1. cuántos núcleos cree tener Crystal (Fiber::ExecutionContext.default_workers_count) en un contenedor con --cpus 2, y
     cuántos hilos del sistema hay con y sin HILOS;
  2. el binario: estático (ldd) y lo que pesa con y sin --release / --no-debug;
  3. versión.

Sin candado: no mide tiempos. Usa las imágenes gl-902-* de pruebas_go_vs_x_2026_11_23.py (o las construye).

    python recursos/pruebas-go-vs-crystal/demos.py
"""
import subprocess
from datetime import datetime
from pathlib import Path

AQUI = Path(__file__).resolve().parent
OUT = []


def sh(args):
    return subprocess.run(args, capture_output=True, text=True, encoding="utf-8", errors="replace")


def w(s=""):
    print(s, flush=True)
    OUT.append(s)


def bloque(cmd, texto, n=14):
    w("```")
    w(f"$ {cmd}")
    for l in texto.strip().splitlines()[:n]:
        w(l.rstrip())
    w("```\n")


def main():
    for df, tag in (("Dockerfile.crystal", "gl-902-crystal:1"), ("Dockerfile.crystal-defecto", "gl-902-crystal-defecto:1")):
        sh(["docker", "build", "-q", "-f", str(AQUI / df), "-t", tag, str(AQUI)])
        sh(["docker", "build", "-q", "--target", "build", "-f", str(AQUI / df), "-t", tag + "-etapa", str(AQUI)])
    w("# SALIDA-demos — ¿Go o Crystal? (golandia 902), sin cronómetro\n")
    w(f"> {datetime.now():%Y-%m-%d %H:%M}. Regenerable con `python recursos/pruebas-go-vs-crystal/demos.py`.\n")
    w("## 1 · Núcleos que ve y hilos, en un contenedor con --cpus 2\n")
    w("La imagen final es scratch (no tiene shell): se mira el mismo binario desde la etapa de compilación (alpine).\n")
    prueba = ("/servidor 2>&1 & p=$!; sleep 1; if kill -0 $p 2>/dev/null; then echo \"# sigue vivo · $(grep Threads /proc/$p/status)\"; "
              "else echo '# el servidor se cayó'; fi")
    for extra, t in (([], "docker run --cpus 2 gl-902-crystal (sin HILOS)"), (["-e", "HILOS=2"], "docker run --cpus 2 -e HILOS=2 gl-902-crystal")):
        r = sh(["docker", "run", "--rm", "--ulimit", "core=1", "--cpus", "2"] + extra + ["--entrypoint", "sh", "gl-902-crystal:1-etapa",
                                                                                         "-c", prueba])
        bloque(t, r.stdout + r.stderr)
    w("## 2 · El binario\n")
    r = sh(["docker", "run", "--rm", "--ulimit", "core=1", "--entrypoint", "sh", "gl-902-crystal:1-etapa", "-c",
            "ldd /servidor 2>&1; cd /src && crystal build --release --static -o /tmp/con-debug servidor.cr && "
            "ls -l /servidor /tmp/con-debug | awk '{print $5, $9}'"])
    bloque("ldd /servidor; ls -l /servidor (--release --no-debug --static) /tmp/con-debug (--release --static)", r.stdout + r.stderr)
    r = sh(["docker", "run", "--rm", "--ulimit", "core=1", "--entrypoint", "sh", "gl-902-crystal-defecto:1-etapa", "-c",
            "ls -l /servidor | awk '{print $5, $9}'"])
    bloque("ls -l /servidor   # crystal build --static, sin --release", r.stdout + r.stderr)
    w("## 3 · Versión\n")
    r = sh(["docker", "run", "--rm", "crystallang/crystal:1.22.3-alpine", "crystal", "--version"])
    bloque("crystal --version", r.stdout + r.stderr)
    (AQUI / "SALIDA-demos.md").write_text("\n".join(OUT) + "\n", encoding="utf-8")
    print("[ok] SALIDA-demos.md")


if __name__ == "__main__":
    main()
