# -*- coding: utf-8 -*-
"""pruebas_go_vs_x_2026_11_23.py — la tanda «¿Go o X? Lo medí» del 2026-11-23 (golandia 902 Crystal, 903 OCaml 5, 905 Odin;
el 904 Mojo no tiene servidor web y va con recursos/pruebas-go-vs-mojo/banco.py), con el MISMO banco que 879 Kotlin Native
(pruebas_go_vs_nativo_2026_11.py → 865 → la revancha): mismo servidor de Go, mismo contrato, servidores con 2 CPU y 1 GiB,
`hey -c 16` con 4 CPU desde otro contenedor, la comprobación de que todos devuelven lo mismo ANTES de medir (si no, se
para), espera a que la máquina esté en reposo, latencia de una petición sola y cómputo puro con el reloj de cada proceso.

No se copia nada: se importa pruebas_go_vs_nativo_2026_11 y se le añaden las tres configuraciones. Lo que cambia:
  · TODO `docker run` lleva `--rm` y `--ulimit core=1` (orden del dueño para esta tanda; el banco viejo arrancaba el
    servidor sin `--rm` y lo borraba con `docker rm -f`, que sigue haciéndose);
  · Odin: odin-http usa io_uring, y el perfil seccomp por defecto de Docker lo bloquea (el servidor se cae con
    «runtime assertion: acquire_err == nil», ver SALIDA-demos.md). Sus contenedores se arrancan con el perfil por defecto de
    Docker MÁS las tres llamadas de io_uring (recursos/pruebas-go-vs-odin/seccomp-io_uring.json), nada más;
  · OCaml: Eio usa io_uring si puede y, si no, cae al motor posix SIN avisar (es lo que pasa en Docker). La variante
    «+ io_uring» lleva el mismo perfil que Odin y EIO_BACKEND=linux para exigirlo;
  · «ajenos» lista todo contenedor que no sea de ESTE vídeo (el banco viejo escondía todo lo que empezara por gl-).

⛔ Solo se ejecuta con el candado de medición tomado (`mkdir <proyecto>/_logs/banco.lock`); el guion lo comprueba.

    python investigacion/pruebas_go_vs_x_2026_11_23.py crystal|ocaml|odin <corrida>
        → recursos/pruebas-go-vs-<x>/SALIDA-corrida-<corrida>.md
"""
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
import pruebas_go_vs_nativo_2026_11 as N  # noqa: E402

L, R = N.L, N.R

L.X["crystal"] = dict(
    id=902, nombre="Crystal", carpeta="pruebas-go-vs-crystal", subred="10.90.2.0/24",
    construir=[("crystal --release", "crystal build --release --no-debug --static", "Dockerfile.crystal", "crystal", "build"),
               ("crystal (modo por defecto)", "crystal build --static (sin --release)", "Dockerfile.crystal-defecto",
                "crystal-defecto", "build")],
    variantes=[("crystal (1 hilo, por defecto)", "crystal", {}), ("crystal, HILOS=2", "crystal", {"HILOS": "2"}),
               ("crystal sin --release", "crystal-defecto", {})],
    bases={"crystal": "scratch + binario estático (musl) de crystal build --release",
           "crystal-defecto": "scratch + binario estático (musl) de crystal build, sin --release"},
    binarios=[("crystal", "/servidor"), ("crystal-defecto", "/servidor")],
    version=["docker", "run", "--rm", "crystallang/crystal:1.22.3-alpine", "crystal", "--version"])

L.X["ocaml"] = dict(
    id=903, nombre="OCaml 5", carpeta="pruebas-go-vs-ocaml", subred="10.90.3.0/24",
    construir=[("ocaml", "opam install dune eio_main cohttp-eio mtime + dune build --profile release", "Dockerfile.ocaml",
                "ocaml", "build")],
    variantes=[("ocaml (1 dominio, por defecto)", "ocaml", {}), ("ocaml, 2 dominios", "ocaml", {"DOMINIOS": "2"}),
               ("ocaml, 24 dominios", "ocaml", {"DOMINIOS": "24"}),
               # Eio usa io_uring si puede; en Docker no puede (seccomp) y cae sin avisar al motor posix. Esta variante le
               # permite io_uring (mismo perfil que Odin) y lo exige con EIO_BACKEND=linux.
               ("ocaml, 2 dominios + io_uring", "ocaml", {"DOMINIOS": "2", "EIO_BACKEND": "linux"})],
    bases={"ocaml": "debian:trixie-slim + binario de ocamlopt (OCaml 5.5.2, cohttp-eio 6.3.0)"},
    binarios=[("ocaml", "/home/opam/servidor")],
    version=["docker", "run", "--rm", "--entrypoint", "bash", "gl-903-ocaml:1-etapa", "-c",
             "ocamlfind ocamlopt -version 2>/dev/null || opam exec -- ocamlopt -version; "
             "opam list --installed --columns=name,version eio eio_linux cohttp-eio 2>/dev/null | grep -v '^#'"])

L.X["odin"] = dict(
    id=905, nombre="Odin", carpeta="pruebas-go-vs-odin", subred="10.90.5.0/24",
    construir=[("odin", "descarga de Odin dev-2026-11a + git clone de odin-http + odin build -o:speed", "Dockerfile.odin", "odin",
                "build")],
    # con io_uring permitido. «item-N» se arma con strconv (como Go); la tercera, con fmt.tprintf (lo cómodo en Odin)
    variantes=[("odin (hilos por defecto)", "odin", {}), ("odin, HILOS=2", "odin", {"HILOS": "2"}),
               ("odin, HILOS=2, nombres con fmt", "odin", {"HILOS": "2", "NOMBRES": "fmt"})],
    bases={"odin": "debian:trixie-slim + binario de odin build -o:speed (odin-http, io_uring)"},
    binarios=[("odin", "/servidor")],
    version=["docker", "run", "--rm", "--entrypoint", "bash", "gl-905-odin:1-etapa", "-c",
             "odin version; echo \"odin-http $(git -C /src/odin-http rev-parse --short HEAD)\""])

SECCOMP = L.REC / "pruebas-go-vs-odin" / "seccomp-io_uring.json"
ID = {"crystal": 902, "ocaml": 903, "odin": 905}
_preparar_viejo = N.preparar
_sh_viejo = None


def sh_rm(args, **kw):
    """`docker run` siempre con --rm y --ulimit core=1; los servidores de Odin, con io_uring permitido."""
    a = list(args)
    if a[:2] == ["docker", "run"]:
        extra = []
        if "--rm" not in a:
            extra.append("--rm")
        if "--ulimit" not in a:
            extra += ["--ulimit", "core=1"]
        if "-d" in a and (any(str(x).startswith("gl-905-odin") for x in a) or "EIO_BACKEND=linux" in a):
            extra += ["--security-opt", f"seccomp={SECCOMP}"]
        a = a[:2] + extra + a[2:]
    return _sh_viejo(a, **kw)


def preparar(cfg):
    global _sh_viejo
    _preparar_viejo(cfg)
    _sh_viejo = R.sh
    R.sh = sh_rm


def ajenos():
    yo = f"gl-{ACTUAL}-"
    r = R.sh(["docker", "stats", "--no-stream", "--format", "{{.Name}}\t{{.CPUPerc}}"])
    return [tuple(l.split("\t")) for l in r.stdout.splitlines() if "\t" in l and not l.startswith(yo)]


ACTUAL = 0
_w_viejo = R.w


def w(s=""):
    """La cabecera de N.main nombra su propio guion: aquí se regenera con este."""
    _w_viejo(s.replace("pruebas_go_vs_nativo_2026_11.py", "pruebas_go_vs_x_2026_11_23.py").replace("(mismo banco que 865)",
                                                                                                   "(mismo banco que 879 y 865)"))


R.w = w

if __name__ == "__main__":
    ACTUAL = ID[sys.argv[1]]
    N.preparar = preparar
    L.preparar = preparar
    N.ajenos = ajenos
    N.main()
