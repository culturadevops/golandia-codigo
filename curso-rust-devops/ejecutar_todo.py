# -*- coding: utf-8 -*-
"""Corre todas las lecciones del curso de Rust para DevOps y regenera SALIDA.md.

La regla de la casa: nada de lo que sale en el vídeo se escribe a mano. Sale de aquí.
Solo hace falta Docker y Python 3.10+.

  · Las lecciones de parte1, parte2 y parte3/01-03 corren DENTRO de rust:1.99.1-trixie (correr.sh).
  · parte3/04_imagen y parte3/05_pid1 corren desde FUERA: construyen las imágenes y miden `docker stop`.

    python ejecutar_todo.py                    # todo, y reescribe SALIDA.md
    python ejecutar_todo.py parte1/02          # solo eso (de dentro), por pantalla

Limpia al terminar: contenedores con --rm o borrados, imágenes rdevops-* borradas. El volumen
rdevops-registry (caché de crates) se deja para no descargar todo otra vez; bórralo con
`docker volume rm rdevops-registry`.
"""
import json
import subprocess
import sys
import time
import urllib.request
from datetime import datetime, timezone
from pathlib import Path

AQUI = Path(__file__).resolve().parent
IMAGEN = "rust:1.99.1-trixie"
CACHE = "rdevops-registry"


def docker(*args, **kw):
    kw.setdefault("capture_output", True)
    kw.setdefault("text", True)
    kw.setdefault("encoding", "utf-8")
    kw.setdefault("errors", "replace")
    return subprocess.run(["docker", *args], **kw)


def dentro(args):
    r = docker("run", "--rm", "--name", "rdevops-lecciones", "-v", f"{AQUI}:/src:ro", "-v", f"{CACHE}:/usr/local/cargo/registry",
               IMAGEN, "bash", "/src/correr.sh", *args, timeout=3600)
    return r.stdout + r.stderr


def esperar_health(puerto):
    for _ in range(100):
        try:
            return urllib.request.urlopen(f"http://127.0.0.1:{puerto}/health", timeout=1).read().decode()
        except Exception:
            time.sleep(0.1)
    return "(no respondió)"


def imagenes():
    out, cod = [], 0
    for tag in ("glibc", "musl"):
        out.append(f"$ docker build -f parte3/04_imagen/Dockerfile.{tag} -t rdevops-servicio:{tag} .")
        r = docker("build", "-q", "-f", f"parte3/04_imagen/Dockerfile.{tag}", "-t", f"rdevops-servicio:{tag}", str(AQUI), timeout=3600)
        if r.returncode:
            out.append(r.stdout + r.stderr)
            return out, r.returncode
        out.append("listo")
    out.append("$ docker image inspect -f '{{.Size}}' rdevops-servicio:glibc rdevops-servicio:musl")
    for tag in ("glibc", "musl"):
        b = int(docker("image", "inspect", "-f", "{{.Size}}", f"rdevops-servicio:{tag}").stdout.strip())
        out.append(f"{tag}: {b} bytes ({b / 1048576:.1f} MiB)")
    out.append("$ docker run -d -p 18090:8090 rdevops-servicio:musl && curl localhost:18090/health")
    docker("rm", "-f", "rdevops-prueba")
    docker("run", "-d", "--name", "rdevops-prueba", "-p", "18090:8090", "rdevops-servicio:musl")
    out.append(esperar_health(18090))
    docker("rm", "-f", "rdevops-prueba")
    return out, cod


def pid1():
    out = []
    for nombre, extra in (("sin manejar SIGTERM", ["--sin-apagado"]), ("con apagado limpio", [])):
        docker("rm", "-f", "rdevops-pid1")
        out.append(f"# {nombre}")
        out.append(f"$ docker run -d --name rdevops-pid1 rdevops-servicio:musl {' '.join(extra)}".rstrip())
        docker("run", "-d", "--name", "rdevops-pid1", "-p", "18091:8090", "rdevops-servicio:musl", *extra)
        esperar_health(18091)
        out.append("$ time docker stop rdevops-pid1")
        t = time.perf_counter()
        docker("stop", "rdevops-pid1")
        out.append(f"[tardó {time.perf_counter() - t:.1f} s]")
        estado = json.loads(docker("inspect", "rdevops-pid1").stdout)[0]["State"]
        out.append(f"código de salida del contenedor: {estado['ExitCode']}")
        logs = docker("logs", "rdevops-pid1")
        ultima = [l for l in (logs.stdout + logs.stderr).splitlines() if l.strip()][-1]
        import re
        out.append("última línea del log: " + re.sub(r'"timestamp":"[^"]*",', "", ultima))
        docker("rm", "-f", "rdevops-pid1")
    return out, 0


def main():
    args = sys.argv[1:]
    if args:
        print(dentro(args))
        return 0
    salida = dentro([])
    bloques, versiones, actual, fallos = [], [], None, 0
    for linea in salida.splitlines():
        if linea.startswith("@@VERSIONES"):
            actual = versiones
        elif linea.startswith("@@INICIO "):
            actual = []
            bloques.append([linea.split(" ", 1)[1], actual, None])
        elif linea.startswith("@@FIN "):
            codigo = int(linea.split()[2])
            fallos += codigo != 0
            bloques[-1][2] = codigo
            print(f"  {bloques[-1][0]} ... " + ("ok" if codigo == 0 else f"FALLO ({codigo})"))
            actual = None
        elif actual is not None:
            actual.append(linea)
    for d, f in (("parte3/04_imagen", imagenes), ("parte3/05_pid1", pid1)):
        lineas, codigo = f()
        fallos += codigo != 0
        bloques.append([d, lineas, codigo])
        print(f"  {d} (desde fuera) ... " + ("ok" if codigo == 0 else f"FALLO ({codigo})"))
    for tag in ("glibc", "musl"):
        docker("image", "rm", "-f", f"rdevops-servicio:{tag}")
    versiones.append("docker " + docker("version", "-f", "{{.Server.Version}}").stdout.strip())
    sello = datetime.now(timezone.utc).strftime("%Y-%m-%d %H:%M:%SZ")
    partes = ["# SALIDA — Curso de Rust para DevOps 2027", "",
              "> Generado por `ejecutar_todo.py`. **No se escribe a mano.** Los fallos que se ven son a propósito:",
              "> cada lección enseña la trampa. Los tiempos cambian un poco de una corrida a otra.", "",
              f"- fecha: `{sello}`", *[f"- `{v}`" for v in versiones], ""]
    for d, lineas, codigo in bloques:
        partes += [f"## {d}", "", f"código de salida `{codigo}`", "", "```", "\n".join(lineas).strip("\n"), "```", ""]
    (AQUI / "SALIDA.md").write_text("\n".join(partes), encoding="utf-8", newline="\n")
    print(f"\n  -> SALIDA.md regenerado ({len(bloques)} lecciones, {fallos} fallos)")
    return 1 if fallos else 0


if __name__ == "__main__":
    sys.exit(main())
