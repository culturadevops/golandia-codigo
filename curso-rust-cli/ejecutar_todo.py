# -*- coding: utf-8 -*-
"""Corre todas las lecciones del curso «Rust para crear CLI» en Docker y regenera SALIDA.md.

La regla de la casa: nada de lo que sale en el vídeo se escribe a mano. Sale de aquí.
Solo hace falta Docker y Python 3.10+ (Rust 1.99.1, jq y el target musl van en la imagen del Dockerfile).
Las crates se bajan una vez a un volumen (rcli-registry) y las dependencias se compilan una vez (rcli-target).

    python ejecutar_todo.py                    # todo, y reescribe SALIDA.md
    python ejecutar_todo.py parte2 parte3/03   # solo eso, por pantalla (no toca SALIDA.md)
    python ejecutar_todo.py --limpiar          # borra la imagen y los dos volúmenes al terminar
"""
import subprocess
import sys
from datetime import datetime, timezone
from pathlib import Path

AQUI = Path(__file__).resolve().parent
IMAGEN = "rcli-curso:1"


def docker(*a, **kw):
    return subprocess.run(["docker", *a], **kw)


def main():
    args = [a for a in sys.argv[1:] if not a.startswith("--")]
    docker("build", "-q", "-t", IMAGEN, str(AQUI), check=True, stdout=subprocess.DEVNULL)
    r = docker("run", "--rm", "-v", f"{AQUI}:/src:ro", "-v", "rcli-registry:/usr/local/cargo/registry",
               "-v", "rcli-target:/work/target", IMAGEN, "bash", "/src/correr.sh", *args,
               capture_output=True, text=True, encoding="utf-8", errors="replace", timeout=3600)
    if "--limpiar" in sys.argv:
        docker("volume", "rm", "-f", "rcli-registry", "rcli-target", capture_output=True)
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
    sello = datetime.now(timezone.utc).strftime("%Y-%m-%d %H:%M:%SZ")
    partes = ["# SALIDA — Curso de Rust para crear CLI 2027", "",
              "> Generado por `ejecutar_todo.py`. **No se escribe a mano.** Los errores son a propósito: cada lección",
              "> enseña la trampa y lo que pasa de verdad. Cada orden va con su «codigo de salida».", "",
              f"- fecha: `{sello}`", *[f"- `{v}`" for v in versiones], ""]
    for d, lineas, codigo in bloques:
        partes += [f"## {d}", "", f"`bash {d}/demo.sh` · código de salida `{codigo}`", "",
                   "```", "\n".join(lineas).strip("\n"), "```", ""]
    (AQUI / "SALIDA.md").write_text("\n".join(partes), encoding="utf-8", newline="\n")
    print(f"\n  -> SALIDA.md regenerado ({len(bloques)} lecciones, {fallos} fallos)")
    return 1 if fallos or r.returncode else 0


if __name__ == "__main__":
    sys.exit(main())
