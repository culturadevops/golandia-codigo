# -*- coding: utf-8 -*-
"""ejecutar_todo.py — regenera las dos SALIDA de «¿Go o TypeScript con Deno?» (golandia 699).

    python canales/golandia/recursos/pruebas-go-vs-deno/ejecutar_todo.py
  1. investigacion/pruebas_go_vs_deno.py → SALIDA.md (el banco de 651-654: imagen, memoria, arranque, peticiones/s, cálculo)
  2. ejecutar_demos.py                   → SALIDA-tipos-permisos.md
  3. limpieza: borra contenedores, red e imágenes v699-* (golang:1.27.2 y denoland/deno:2.9.8 se quedan, son compartidas)
"""
import os
import subprocess
import sys
from pathlib import Path

AQUI = Path(__file__).resolve().parent
BANCO = AQUI.parent.parent / "investigacion" / "pruebas_go_vs_deno.py"
env = dict(os.environ, PYTHONUTF8="1")
subprocess.run([sys.executable, str(BANCO)], check=True, env=env)
subprocess.run([sys.executable, str(AQUI / "ejecutar_demos.py")], check=True, env=env)
if "--sin-limpiar" not in sys.argv:
    def sh(*a):
        return subprocess.run(list(a), capture_output=True, text=True).stdout.split()
    for c in sh("docker", "ps", "-aq", "--filter", "name=v699-"):
        sh("docker", "rm", "-f", c)
    sh("docker", "network", "rm", "v699-red")
    imgs = [i for i in sh("docker", "images", "--format", "{{.Repository}}:{{.Tag}}") if i.startswith("v699-")]
    if imgs:
        sh("docker", "rmi", *imgs)
    print("[ok] limpieza v699-*:", len(imgs), "imágenes")
