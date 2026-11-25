# -*- coding: utf-8 -*-
"""Demo del vídeo 680 «Pulumi con Go vs Terraform» (golandia): la misma infraestructura en los dos, medida contra moto.

    python correr.py

Levanta (y al final borra) solo cosas con nombre g680-*: la red g680, el emulador g680-moto (alias «moto» en la red) y
g680-tools (imagen g680-iac: golang:1.27.2 + Pulumi 3.272.0 + Terraform 1.16.2, ver Dockerfile). Dentro corre medir.sh,
que trabaja en un mktemp y deja los resultados en salidas/*.txt. SALIDA.md se escribe a mano a partir de ellas.
Las cachés de Go y de plugins viven en los volúmenes g680-gomod, g680-gocache y g680-pulumi (se conservan).
"""
import os
import subprocess
import sys
from pathlib import Path

AQUI = Path(__file__).parent
ENV = {**os.environ, "MSYS_NO_PATHCONV": "1"}
DIR_W = subprocess.run(["cygpath", "-w", str(AQUI)], capture_output=True, text=True).stdout.strip() or str(AQUI)


def sh(*a, check=True):
    print("$", " ".join(a), flush=True)
    r = subprocess.run(list(a), env=ENV, text=True, encoding="utf-8", errors="replace")
    if check and r.returncode != 0:
        raise SystemExit(f"falló: {' '.join(a)}")
    return r


def salida(*a):
    return subprocess.run(list(a), env=ENV, capture_output=True, text=True, encoding="utf-8", errors="replace").stdout.strip()


def limpiar():
    for n in ("g680-tools", "g680-moto"):
        subprocess.run(["docker", "rm", "-f", n], env=ENV, capture_output=True)
    subprocess.run(["docker", "network", "rm", "g680"], env=ENV, capture_output=True)


def main():
    limpiar()
    sh("docker", "build", "-q", "-t", "g680-iac", DIR_W)
    sh("docker", "network", "create", "g680")
    sh("docker", "run", "-d", "--name", "g680-moto", "--network", "g680", "--network-alias", "moto", "motoserver/moto:latest")
    moto = salida("docker", "exec", "g680-moto", "python", "-c", "import moto;print(moto.__version__)")
    sh("docker", "run", "-d", "--name", "g680-tools", "--network", "g680", "-v", f"{DIR_W}:/demo",
       "-v", "g680-gomod:/go/pkg/mod", "-v", "g680-gocache:/root/.cache/go-build", "-v", "g680-pulumi:/root/.pulumi-plugins",
       "-e", "PULUMI_HOME=/root/.pulumi-plugins", "-e", f"MOTO_VERSION={moto}", "g680-iac", "sleep", "infinity")
    try:
        sh("docker", "exec", "g680-tools", "bash", "/demo/medir.sh")
    finally:
        limpiar()


if __name__ == "__main__":
    sys.exit(main())
