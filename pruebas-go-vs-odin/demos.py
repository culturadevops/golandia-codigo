# -*- coding: utf-8 -*-
"""demos.py — lo de «¿Go o Odin?» (golandia 905) que no lleva cronómetro → SALIDA-demos.md

  1. el servidor (odin-http, io_uring) con el perfil seccomp por defecto de Docker, sin seccomp, y con el perfil por defecto
     más io_uring (seccomp-io_uring.json: el default.json de github.com/moby/profiles con io_uring_setup/enter/register);
     y en cada caso, cuántos hilos arranca odin-http por su cuenta dentro de un contenedor con --cpus 2;
  2. ldd del binario.

Sin candado: no mide tiempos. Usa la imagen gl-905-odin:1 de pruebas_go_vs_x_2026_11_23.py (o la construye).

    python recursos/pruebas-go-vs-odin/demos.py
"""
import subprocess
from datetime import datetime
from pathlib import Path

AQUI = Path(__file__).resolve().parent
IMG = "gl-905-odin:1"
OUT = []


def sh(args):
    return subprocess.run(args, capture_output=True, text=True, encoding="utf-8", errors="replace")


def w(s=""):
    print(s, flush=True)
    OUT.append(s)


def bloque(cmd, texto, n=12):
    w("```")
    w(f"$ {cmd}")
    for l in texto.strip().splitlines()[:n]:
        w(l.rstrip())
    w("```\n")


def main():
    sh(["docker", "build", "-q", "-f", str(AQUI / "Dockerfile.odin"), "-t", IMG, str(AQUI)])
    w("# SALIDA-demos — ¿Go o Odin? (golandia 905), sin cronómetro\n")
    w(f"> {datetime.now():%Y-%m-%d %H:%M}. Regenerable con `python recursos/pruebas-go-vs-odin/demos.py`.\n")
    w("## 1 · El servidor con el perfil seccomp por defecto de Docker, y con io_uring permitido\n")
    prueba = ("/servidor 2>&1 & p=$!; sleep 2; if kill -0 $p 2>/dev/null; then echo \"# sigue vivo · $(grep Threads /proc/$p/status)\"; "
              "else echo '# el servidor se cayó'; fi")
    for titulo, extra in [("docker run --cpus 2 gl-905-odin:1", []),
                          ("docker run --cpus 2 --security-opt seccomp=unconfined gl-905-odin:1", ["--security-opt", "seccomp=unconfined"]),
                          ("docker run --cpus 2 --security-opt seccomp=seccomp-io_uring.json gl-905-odin:1   # el de Docker + io_uring",
                           ["--security-opt", f"seccomp={AQUI / 'seccomp-io_uring.json'}"]),
                          ("docker run --cpus 2 --security-opt seccomp=seccomp-io_uring.json -e HILOS=2 gl-905-odin:1",
                           ["--security-opt", f"seccomp={AQUI / 'seccomp-io_uring.json'}", "-e", "HILOS=2"])]:
        r = sh(["docker", "run", "--rm", "--ulimit", "core=1", "--cpus", "2"] + extra + ["--entrypoint", "sh", IMG, "-c", prueba])
        bloque(titulo, r.stdout + r.stderr)
    w("## 2 · Lo que necesita el binario\n")
    r = sh(["docker", "run", "--rm", "--ulimit", "core=1", "--entrypoint", "sh", IMG, "-c", "ldd /servidor | sed 's/ (0x.*//'"])
    bloque("ldd /servidor", r.stdout + r.stderr)
    (AQUI / "SALIDA-demos.md").write_text("\n".join(OUT) + "\n", encoding="utf-8")
    print("[ok] SALIDA-demos.md")


if __name__ == "__main__":
    main()
