# -*- coding: utf-8 -*-
"""Vengo de Python: lo que Go me obliga a cambiar — banco ejecutable (golandia, vídeo 389). Escribe SALIDA.md.

  1  trabajo de CPU en 8 trozos: Python en fila, con 8 hilos y con 8 procesos; Go en fila y con 8 goroutines
  2  la errata en una rama que casi nunca se ejecuta: Python la descubre al ejecutar, Go al compilar
  3  el mismo servicio HTTP mínimo en una imagen: python:3.14-slim frente a scratch + binario
  4  memoria en reposo de los dos contenedores (docker stats)

    python correr.py
"""
import subprocess
import sys
import time
from pathlib import Path

AQUI = Path(__file__).resolve().parent
sys.path.insert(0, str(AQUI.parent))
from banco_go import Banco  # noqa: E402

PROY = AQUI / "proyecto"
PY = "python:3.14-slim"


def run(cmd):
    r = subprocess.run(cmd, capture_output=True, text=True, encoding="utf-8", errors="replace")
    return (r.stdout + r.stderr).strip()


def py(b, orden, mostrar):
    b.anotar(mostrar, run(["docker", "run", "--rm", "-v", f"{PROY}:/w", "-w", "/w", PY, "bash", "-c", orden]))


with Banco(AQUI, "Vengo de Python: lo que Go cambia") as b:
    b.titulo("Las versiones")
    b.sh("go version")
    py(b, "python -VV", "python -VV")

    b.titulo("1 · Trabajo de CPU repartido en 8 trozos")
    py(b, "python primos.py", "python primos.py")
    b.sh("go run ./primos", mostrar="go run ./primos")

    b.titulo("2 · La errata en la rama que casi nunca se ejecuta")
    py(b, "python -m py_compile errata.py && echo 'compila: sin errores'; python errata.py",
       "python -m py_compile errata.py && echo 'compila: sin errores'; python errata.py")
    b.sh("go build -o /dev/null ./errata", mostrar="go build ./errata")
    b.sh("go vet ./errata", mostrar="go vet ./errata")

    b.titulo("2b · Un tipo equivocado")
    py(b, "python tipos.py", "python tipos.py")
    b.sh("go build -o /dev/null ./tipos", mostrar="go build ./tipos")

    b.titulo("2c · Esperar a la red: 50 peticiones")
    b.sh("CGO_ENABLED=0 go build -o sim ./simulador", guardar=False)
    py(b, "(./sim &); sleep 1; python esperar.py", "python esperar.py")
    b.sh("(./sim &); sleep 1; go run ./esperar", mostrar="go run ./esperar")

    b.titulo("2d · Lo que tarda en arrancar")
    py(b, "s=$(date +%s%N); for i in $(seq 20); do python -c 'print(1)' >/dev/null; done; e=$(date +%s%N); "
          "echo \"python, 20 arranques: $(( (e-s)/20000000 )) ms cada uno\"",
       "for i in $(seq 20); do python -c 'print(1)'; done   # media de 20")
    b.sh("go build -o hola-bin ./hola && s=$(date +%s%N); for i in $(seq 20); do ./hola-bin >/dev/null; done; "
         "e=$(date +%s%N); echo \"binario de Go, 20 arranques: $(( (e-s)/20000000 )) ms cada uno\"; rm -f hola-bin sim",
         mostrar="for i in $(seq 20); do ./hola; done   # media de 20")

    b.titulo("3 · El mismo servicio, en una imagen")
    run(["docker", "build", "-q", "-f", str(PROY / "Dockerfile.python"), "-t", "servicio-python", str(PROY)])
    run(["docker", "build", "-q", "-f", str(PROY / "Dockerfile.go"), "-t", "servicio-go", str(PROY)])
    b.anotar("docker images servicio-python servicio-go",
             run(["docker", "images", "--format", "{{.Repository}}  {{.Size}}", "servicio-python"]) + "\n" +
             run(["docker", "images", "--format", "{{.Repository}}  {{.Size}}", "servicio-go"]))

    b.titulo("4 · Memoria en reposo")
    for n in ("servicio-python", "servicio-go"):
        run(["docker", "rm", "-f", n])
        run(["docker", "run", "-d", "--name", n, n])
    time.sleep(5)
    b.anotar("docker stats --no-stream servicio-python servicio-go",
             run(["docker", "stats", "--no-stream", "--format", "{{.Name}}  {{.MemUsage}}",
                  "servicio-python", "servicio-go"]))
    for n in ("servicio-python", "servicio-go"):
        run(["docker", "rm", "-f", n])
