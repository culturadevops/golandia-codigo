# -*- coding: utf-8 -*-
"""Compila, abre (Xvfb) y mide cada lección dentro de Docker; regenera SALIDA.md y capturas/.

La regla de la casa: nada de lo que sale en el vídeo se escribe a mano. Cada lección termina sola
a los pocos segundos e imprime lo que midió.

    python ejecutar_todo.py
"""
import os
import subprocess
import sys
from datetime import datetime, timezone
from pathlib import Path

AQUI = Path(__file__).resolve().parent
ENV = dict(os.environ, MSYS_NO_PATHCONV="1", PYTHONIOENCODING="utf-8")
IMG = "curso-ebiten:1"
VOL = ["-v", "gomod-ebiten:/go/pkg/mod", "-v", "gocache-ebiten:/root/.cache/go-build", "-v", f"{AQUI}:/app"]

# (número, título, orden dentro del contenedor)
LECCIONES = [
    ("1", "El bucle: Update y Draw no son lo mismo",
     "bash correr.sh 01_bucle; VSYNC=no bash correr.sh 01_bucle 01_bucle-sin-vsync"),
    ("2", "La trampa: mover el personaje en Draw",
     "HZ=60 bash correr.sh 02_mover_en_draw 02_mover-60hz; HZ=144 bash correr.sh 02_mover_en_draw 02_mover-144hz"),
    ("3", "Tiempo fijo: la velocidad por segundo, no por tick",
     "TPS=60 bash correr.sh 03_tps; TPS=30 bash correr.sh 03_tps"),
    ("4", "Lo que tarda el dibujado no frena la lógica; lo que tarda la lógica, sí",
     "DRAW=0 UPDATE=0 bash correr.sh 04_lento; DRAW=50 UPDATE=0 bash correr.sh 04_lento; DRAW=0 UPDATE=30 bash correr.sh 04_lento"),
    ("5", "La entrada: pulsada no es «se acaba de pulsar»", "bash correr.sh 05_entrada"),
    ("6", "Probar la lógica sin ventana", "go test -v -count=1 ./mundo"),
    ("7", "El proyecto: esquiva", "DEMO=1 bash correr.sh 07_esquiva"),
]


def correr(cmd, timeout=1200):
    r = subprocess.run(cmd, capture_output=True, text=True, encoding="utf-8", errors="replace", env=ENV,
                       cwd=AQUI, timeout=timeout)
    salida = "\n".join(l for l in (r.stdout + r.stderr).splitlines() if not l.startswith("go: downloading"))
    return r.returncode, salida


def main():
    sello = datetime.now(timezone.utc).strftime("%Y-%m-%d %H:%M:%SZ")
    c, s = correr(["docker", "build", "-q", "-t", IMG, "."])
    if c != 0:
        raise SystemExit(s)
    _, vers = correr(["docker", "run", "--rm", *VOL, IMG, "sh", "-c", "go version; go list -m github.com/hajimehoshi/ebiten/v2"])
    partes = ["# SALIDA — curso «Ebitengine: haz un juego en Go y entiende el bucle de fotogramas»", "",
              "> Generado por `ejecutar_todo.py`. **No se escribe a mano.** Capturas en `capturas/`.", "",
              f"- fecha: `{sello}` · todo compilado y abierto en Docker (Debian 12) sobre Xvfb, con Mesa por software",
              *[f"- `{l}`" for l in vers.strip().splitlines()],
              "- ⚠️ en una pantalla virtual no hay refresco de monitor: el vsync no limita los fotogramas", ""]
    fallos = 0
    for num, titulo, orden in LECCIONES:
        print(f"  [{num}] {titulo} ...", end=" ", flush=True)
        codigo, salida = correr(["docker", "run", "--rm", *VOL, IMG, "bash", "-c", orden])
        fallos += codigo != 0
        partes += [f"## Leccion {num} — {titulo}", "", f"`{orden}` · codigo de salida `{codigo}`", "",
                   "```", salida.strip("\n"), "```", ""]
        print("ok" if codigo == 0 else f"FALLO ({codigo})")
    (AQUI / "SALIDA.md").write_text("\n".join(partes) + "\n", encoding="utf-8", newline="\n")
    print(f"\n  -> SALIDA.md y capturas/ regenerados ({fallos} fallos)")
    return 1 if fallos else 0


if __name__ == "__main__":
    sys.exit(main())
