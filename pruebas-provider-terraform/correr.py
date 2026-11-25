# -*- coding: utf-8 -*-
"""Demo del vídeo «Terraform provider en Go» (golandia 675).

    python correr.py [escenario ...]
    escenarios: versiones ciclo drift import trampa-clave trampa-404 trampa-lista trampa-vacio  (por defecto, todos)

Construye la imagen g675-tf (golang:1.27.2 + terraform 1.16.2), monta este directorio de solo lectura en /demo y
corre escenarios.sh <escenario> en un contenedor desechable. Todo pasa en un mktemp dentro del contenedor.
Cada escenario deja su salida, sin editar, en salidas/<escenario>.txt. SALIDA.md se escribe a mano a partir de ellas.
"""
import os
import subprocess
import sys
from pathlib import Path

AQUI = Path(__file__).parent
SAL = AQUI / "salidas"
SAL.mkdir(exist_ok=True)
ENV = {**os.environ, "MSYS_NO_PATHCONV": "1"}
IMG = "g675-tf"
TODOS = ["versiones", "ciclo", "drift", "import", "trampa-clave", "trampa-404", "trampa-lista", "trampa-vacio"]


def main():
    escenarios = sys.argv[1:] or TODOS
    subprocess.run(["docker", "build", "-q", "-t", IMG, str(AQUI)], check=True, env=ENV)
    for e in escenarios:
        r = subprocess.run(["docker", "run", "--rm", "--cpus", "4",
                            "-v", f"{AQUI.resolve()}:/demo:ro",
                            "-v", "g675-gomod:/go/pkg/mod", "-v", "g675-gocache:/root/.cache/go-build",
                            IMG, "sh", "/demo/escenarios.sh", e],
                           capture_output=True, text=True, encoding="utf-8", errors="replace", env=ENV)
        (SAL / f"{e}.txt").write_text(r.stdout + r.stderr, encoding="utf-8", newline="\n")
        print(f"[{'ok' if r.returncode == 0 else 'FALLO'}] {e} -> salidas/{e}.txt ({len(r.stdout.splitlines())} lineas)")


if __name__ == "__main__":
    main()
