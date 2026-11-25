# -*- coding: utf-8 -*-
"""Corre todas las lecciones del Curso de Go desde cero 2027 en Docker y regenera SALIDA.md.

La regla de la casa: nada de lo que sale en el vídeo se escribe a mano. Sale de aquí.
Solo hace falta Docker y Python 3.10+ (Go va dentro de la imagen oficial golang:1.27.2).

    python ejecutar_todo.py                    # todo, y reescribe SALIDA.md
    python ejecutar_todo.py parte3 parte5/06   # solo eso, por pantalla (no toca SALIDA.md)
"""
import subprocess
import sys
from datetime import datetime, timezone
from pathlib import Path

AQUI = Path(__file__).resolve().parent
IMAGEN = "golang:1.27.2"


def main():
    args = sys.argv[1:]
    r = subprocess.run(["docker", "run", "--rm", "-v", f"{AQUI}:/src:ro", "-v", "gocero-cache:/root/.cache/go-build", "-v", "gocero-mod:/go/pkg/mod", IMAGEN, "bash", "/src/correr.sh", *args],
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
    partes = ["# SALIDA — Curso de Go desde cero 2027", "",
              "> Generado por `ejecutar_todo.py`. **No se escribe a mano.** Los errores de compilación son a propósito:",
              "> cada lección enseña qué dice el compilador. Con go run, un programa que falla o no compila devuelve 1 y escribe «exit status N» con el código real.", "",
              f"- fecha: `{sello}`", *[f"- `{v}`" for v in versiones], ""]
    for d, lineas, codigo in bloques:
        partes += [f"## {d}", "", f"`bash {d}/demo.sh` · código de salida `{codigo}`", "",
                   "```", "\n".join(lineas).strip("\n"), "```", ""]
    (AQUI / "SALIDA.md").write_text("\n".join(partes), encoding="utf-8", newline="\n")
    print(f"\n  -> SALIDA.md regenerado ({len(bloques)} lecciones, {fallos} fallos)")
    return 1 if fallos or r.returncode else 0


if __name__ == "__main__":
    sys.exit(main())
