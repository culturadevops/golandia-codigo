# -*- coding: utf-8 -*-
"""Compila, abre y fotografía cada lección dentro de Docker (Xvfb), y regenera SALIDA.md y capturas/.

La regla de la casa: nada de lo que sale en el vídeo se escribe a mano. Las capturas son ventanas de
verdad, y lo que ve el JavaScript de cada ventana lo escribe Go (comun.Chivato) y acaba aquí.
Hace falta Docker y la imagen de nubecita (lección 8): docker build -t nubecita:curso <ruta de nubecita>.

    python ejecutar_todo.py
"""
import os
import subprocess
import sys
from datetime import datetime, timezone
from pathlib import Path

AQUI = Path(__file__).resolve().parent
ENV = dict(os.environ, MSYS_NO_PATHCONV="1", PYTHONIOENCODING="utf-8")
IMG = "curso-wails:1"
# En Windows, `bash` a secas puede ser el de WSL: se usa el de Git si está.
BASH = next((b for b in ("C:/Program Files/Git/bin/bash.exe",) if Path(b).exists()), "bash")
VOL = ["-v", "gomod-wails:/go/pkg/mod", "-v", "gocache-wails:/root/.cache/go-build", "-v", f"{AQUI}:/app"]

# (número, título, carpeta, espera antes de la captura, variables de entorno)
LECCIONES = [
    ("1", "La primera ventana", "01_ventana", 6, {}),
    ("2", "Los tipos, por el puente", "02_tipos", 8, {}),
    ("3", "Los errores, por el puente", "03_errores", 7, {}),
    ("4", "Eventos: los que se pierden al arrancar", "04_eventos", 7, {}),
    ("4b", "Eventos: esperando a que la ventana diga «listo»", "04_eventos", 7, {"MODO": "bueno", "NOMBRE": "04_eventos-bueno"}),
    ("5", "El contexto de Wails", "05_contexto", 3, {"DESPUES": "7"}),
    ("6", "El motor es del sistema", "06_motor", 6, {}),
]


def correr(cmd, timeout=1200):
    r = subprocess.run(cmd, capture_output=True, text=True, encoding="utf-8", errors="replace", env=ENV,
                       cwd=AQUI, timeout=timeout)
    salida = "\n".join(l for l in (r.stdout + r.stderr).splitlines()
                       if not l.startswith("go: downloading") and "signal 10" not in l)
    return r.returncode, salida


def main():
    sello = datetime.now(timezone.utc).strftime("%Y-%m-%d %H:%M:%SZ")
    c, s = correr(["docker", "build", "-q", "-t", IMG, "."])
    if c != 0:
        raise SystemExit(s)
    _, vers = correr(["docker", "run", "--rm", *VOL, IMG, "sh", "-c",
                      "go version; go list -m github.com/wailsapp/wails/v2; dpkg -l libwebkit2gtk-4.1-0 | tail -1 | awk '{print $2, $3}'"])
    partes = ["# SALIDA — curso «Wails: tu frontend web con un backend de Go de verdad»", "",
              "> Generado por `ejecutar_todo.py`. **No se escribe a mano.** Capturas en `capturas/`.", "",
              f"- fecha: `{sello}` · todo compilado y abierto en Docker (Debian 12) sobre Xvfb",
              *[f"- `{l}`" for l in vers.strip().splitlines()], ""]
    fallos = 0
    for num, titulo, carpeta, espera, entorno in LECCIONES:
        print(f"  [{num}] {titulo} ...", end=" ", flush=True)
        env = [x for k, v in entorno.items() for x in ("-e", f"{k}={v}")]
        codigo, salida = correr(["docker", "run", "--rm", *env, *VOL, IMG, "bash", "capturar.sh", carpeta, str(espera)])
        fallos += codigo != 0
        partes += [f"## Leccion {num} — {titulo}", "", f"`bash capturar.sh {carpeta} {espera}` "
                   f"{' '.join(f'{k}={v}' for k, v in entorno.items())} · codigo de salida `{codigo}`", "",
                   "```", salida.strip("\n"), "```", ""]
        print("ok" if codigo == 0 else f"FALLO ({codigo})")
    print("  [7] Empaquetar ...", end=" ", flush=True)
    codigo, salida = correr(["docker", "run", "--rm", *VOL, IMG, "bash", "/app/07_empaquetar.sh"])
    fallos += codigo != 0
    partes += ["## Leccion 7 — Empaquetar", "", f"`bash 07_empaquetar.sh` · codigo de salida `{codigo}`", "",
               "```", salida.strip("\n"), "```", ""]
    print("ok" if codigo == 0 else f"FALLO ({codigo})")
    print("  [8] El proyecto: inspector de colas ...", end=" ", flush=True)
    codigo, salida = correr([BASH, "08_proyecto.sh"])
    fallos += codigo != 0
    partes += ["## Leccion 8 — El proyecto: un inspector de colas para nubecita", "",
               f"`bash 08_proyecto.sh` · codigo de salida `{codigo}`", "", "```", salida.strip("\n"), "```", ""]
    print("ok" if codigo == 0 else f"FALLO ({codigo})")
    (AQUI / "SALIDA.md").write_text("\n".join(partes) + "\n", encoding="utf-8", newline="\n")
    print(f"\n  -> SALIDA.md y capturas/ regenerados ({fallos} fallos)")
    return 1 if fallos else 0


if __name__ == "__main__":
    sys.exit(main())
