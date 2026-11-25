# -*- coding: utf-8 -*-
"""demos.py — lo de «¿Go o OCaml 5?» (golandia 903) que no lleva cronómetro → SALIDA-demos.md

  1. cuántos núcleos cree tener OCaml (Domain.recommended_domain_count) en un contenedor con --cpus 2, y cuántos hilos
     del sistema arranca el servidor con 1, 2 y 24 dominios;
  2. el motor de Eio en Docker: con io_uring bloqueado por seccomp cae a posix sin avisar; si lo exiges, falla;
  3. versiones de lo instalado con opam y ldd del binario.

Sin candado: no mide tiempos. Usa las imágenes gl-903-* de pruebas_go_vs_x_2026_11_23.py (o las construye).

    python recursos/pruebas-go-vs-ocaml/demos.py
"""
import subprocess
from datetime import datetime
from pathlib import Path

AQUI = Path(__file__).resolve().parent
IMG = "gl-903-ocaml:1"
SECCOMP = AQUI.parent / "pruebas-go-vs-odin" / "seccomp-io_uring.json"
OUT = []
PRUEBA = ("/servidor 2>&1 & p=$!; sleep 2; if kill -0 $p 2>/dev/null; then echo \"# sigue vivo · $(grep Threads /proc/$p/status)\"; "
          "else echo '# el servidor se cayó'; fi")


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


def probar(titulo, extra):
    r = sh(["docker", "run", "--rm", "--ulimit", "core=1", "--cpus", "2"] + extra + ["--entrypoint", "sh", IMG, "-c", PRUEBA])
    bloque(titulo, r.stdout + r.stderr)


def main():
    sh(["docker", "build", "-q", "-f", str(AQUI / "Dockerfile.ocaml"), "-t", IMG, str(AQUI)])
    sh(["docker", "build", "-q", "--target", "build", "-f", str(AQUI / "Dockerfile.ocaml"), "-t", "gl-903-ocaml:1-etapa", str(AQUI)])
    w("# SALIDA-demos — ¿Go o OCaml 5? (golandia 903), sin cronómetro\n")
    w(f"> {datetime.now():%Y-%m-%d %H:%M}. Regenerable con `python recursos/pruebas-go-vs-ocaml/demos.py`.\n")
    w("## 1 · Núcleos que ve y hilos que arranca, en un contenedor con --cpus 2\n")
    for d in ("1", "2", "24"):
        probar(f"docker run --cpus 2 -e DOMINIOS={d} gl-903-ocaml:1", ["-e", f"DOMINIOS={d}"])
    w("## 2 · El motor de Eio dentro de Docker\n")
    probar("docker run --cpus 2 -e EIO_BACKEND=linux gl-903-ocaml:1   # exigir io_uring", ["-e", "EIO_BACKEND=linux"])
    probar("docker run --cpus 2 --security-opt seccomp=seccomp-io_uring.json -e EIO_BACKEND=linux gl-903-ocaml:1",
           ["--security-opt", f"seccomp={SECCOMP}", "-e", "EIO_BACKEND=linux"])
    w("Sin EIO_BACKEND, el mismo contenedor arranca igual (sección 1): Eio prueba io_uring, no puede, y usa posix sin decir nada.\n")
    w("## 3 · Versiones y lo que necesita el binario\n")
    r = sh(["docker", "run", "--rm", "--ulimit", "core=1", "--entrypoint", "bash", "gl-903-ocaml:1-etapa", "-c",
            "opam exec -- ocamlopt -version; opam list --installed --columns=name,version ocaml eio eio_linux eio_posix cohttp-eio "
            "dune 2>/dev/null | grep -v '^#'"])
    bloque("ocamlopt -version; opam list --installed ...", r.stdout + r.stderr)
    r = sh(["docker", "run", "--rm", "--ulimit", "core=1", "--entrypoint", "sh", IMG, "-c", "ldd /servidor | sed 's/ (0x.*//'"])
    bloque("ldd /servidor", r.stdout + r.stderr)
    (AQUI / "SALIDA-demos.md").write_text("\n".join(OUT) + "\n", encoding="utf-8")
    print("[ok] SALIDA-demos.md")


if __name__ == "__main__":
    main()
