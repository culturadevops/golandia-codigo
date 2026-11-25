# -*- coding: utf-8 -*-
"""Corre todas las lecciones del curso de Rust para operadores Linux en Docker y regenera SALIDA.md.

La regla de la casa: nada de lo que sale en el vídeo se escribe a mano. Sale de aquí.
Solo hace falta Docker y Python 3.10+. El contenedor va con límites a propósito: 1 GiB de
memoria sin swap, 2 CPUs y /dev/shm de 256 MiB. Varias lecciones miden justo eso.

    python ejecutar_todo.py                  # todo, y reescribe SALIDA.md
    python ejecutar_todo.py parte2/02        # solo eso, por pantalla (no toca SALIDA.md)
"""
import subprocess
import sys
from datetime import datetime, timezone
from pathlib import Path

AQUI = Path(__file__).resolve().parent
IMAGEN = "rust:1.99.1-trixie"
REGISTRO = "roper-registry"   # caché de crates.io entre corridas


def main():
    args = sys.argv[1:]
    r = subprocess.run(["docker", "run", "--rm", "--memory=1g", "--memory-swap=1g", "--cpus=2", "--shm-size=256m",
                        "-v", f"{REGISTRO}:/usr/local/cargo/registry", "-v", f"{AQUI}:/src:ro",
                        IMAGEN, "bash", "/src/correr.sh", *args],
                       capture_output=True, text=True, encoding="utf-8", errors="replace", timeout=3600)
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
    partes = ["# SALIDA — Curso de Rust para operadores Linux 2027", "",
              "> Generado por `ejecutar_todo.py`. **No se escribe a mano.** Cada comando va con su código de salida;",
              "> los fallos y las alertas son a propósito: cada lección enseña una trampa.", "",
              f"- fecha: `{sello}`", *[f"- `{v}`" for v in versiones], ""]
    for d, lineas, codigo in bloques:
        partes += [f"## {d}", "", f"`bash {d}/demo.sh` · código de salida `{codigo}`", "",
                   "```", "\n".join(lineas).strip("\n"), "```", ""]
    (AQUI / "SALIDA.md").write_text("\n".join(partes), encoding="utf-8", newline="\n")
    print(f"\n  -> SALIDA.md regenerado ({len(bloques)} lecciones, {fallos} fallos)")
    return 1 if fallos or r.returncode else 0


if __name__ == "__main__":
    sys.exit(main())
