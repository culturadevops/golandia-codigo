# -*- coding: utf-8 -*-
"""Corre las 12 lecciones en Docker y regenera SALIDA.md.

La regla de la casa: nada de lo que sale en el video se escribe a mano. Sale de aqui.
Solo hace falta Docker y Python 3.10+ (Go, jq y GoReleaser van dentro de la imagen).

    python ejecutar_todo.py              # todas, y reescribe SALIDA.md
    python ejecutar_todo.py 05 08        # solo esas lecciones, por pantalla (no toca SALIDA.md)
    python ejecutar_todo.py --dejar-cache   # no borra la imagen ni el volumen con los modulos de Go

Al terminar no deja nada: el contenedor va con --rm, y se borran la imagen y el volumen.
"""
import subprocess
import sys
from datetime import datetime, timezone
from pathlib import Path

AQUI = Path(__file__).resolve().parent
IMAGEN = "g689-curso-cli:1"
CONTENEDOR = "g689-curso-cli"
CACHE = "g689-curso-cli-gomod"


def docker(*args, **kw):
    return subprocess.run(["docker", *args], **kw)


def main():
    args = [a for a in sys.argv[1:] if not a.startswith("--")]
    print("  construyendo la imagen del laboratorio ...", flush=True)
    docker("build", "-q", "-t", IMAGEN, str(AQUI), check=True, stdout=subprocess.DEVNULL)
    docker("rm", "-f", CONTENEDOR, capture_output=True)
    print("  corriendo las lecciones en Docker ...", flush=True)
    r = docker("run", "--rm", "--name", CONTENEDOR,
               "-v", f"{AQUI}:/src:ro", "-v", f"{CACHE}:/go/pkg/mod",
               IMAGEN, "bash", "/src/correr.sh", *args,
               capture_output=True, text=True, encoding="utf-8", errors="replace", timeout=1800)
    if "--dejar-cache" not in sys.argv:
        docker("volume", "rm", "-f", CACHE, capture_output=True)
        docker("image", "rm", "-f", IMAGEN, capture_output=True)
    salida = r.stdout + r.stderr

    if args:
        print(salida)
        return 0

    bloques, versiones, actual, fallos = [], [], None, 0
    for linea in salida.splitlines():
        if linea.startswith("@@VERSIONES"):
            actual = versiones
        elif linea.startswith("@@INICIO "):
            _, n, d = linea.split(" ", 2)
            actual = []
            bloques.append((n, d, actual))
        elif linea.startswith("@@FIN "):
            codigo = int(linea.split()[2])
            fallos += codigo != 0
            bloques[-1] = (*bloques[-1][:2], actual, codigo)
            print(f"  [{bloques[-1][0]}] {bloques[-1][1]} ... " + ("ok" if codigo == 0 else f"FALLO ({codigo})"))
            actual = None
        elif actual is not None:
            actual.append(linea)

    sello = datetime.now(timezone.utc).strftime("%Y-%m-%d %H:%M:%SZ")
    partes = ["# SALIDA — curso «Herramientas de línea de comandos en Go, de cero a release»", "",
              "> Generado por `ejecutar_todo.py`. **No se escribe a mano.**", "",
              f"- fecha: `{sello}`", *[f"- `{v}`" for v in versiones], ""]
    for b in bloques:
        n, d, lineas = b[0], b[1], b[2]
        codigo = b[3] if len(b) > 3 else "sin terminar"
        partes += [f"## Lección {n} — `{d}`", "", f"`bash {d}/demo.sh` · código de salida `{codigo}`", "",
                   "```", "\n".join(lineas).strip("\n"), "```", ""]
    (AQUI / "SALIDA.md").write_text("\n".join(partes), encoding="utf-8", newline="\n")
    print(f"\n  -> SALIDA.md regenerado ({len(bloques)} lecciones, {fallos} fallos)")
    return 1 if fallos or r.returncode else 0


if __name__ == "__main__":
    sys.exit(main())
