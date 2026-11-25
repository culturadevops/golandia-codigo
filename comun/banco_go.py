# -*- coding: utf-8 -*-
"""Lo común de los bancos de Go de golandia: ejecutar comandos dentro de golang:<VERSION> en Docker
(App Control bloquea binarios de Go en Windows) con la carpeta del proyecto montada en /w, y
escribir SALIDA.md con lo que se ejecutó y lo que salió, tal cual.

    from banco_go import Banco
    with Banco(Path(__file__).parent, "PGO en Go") as b:
        b.titulo("Sin PGO")
        b.sh("go test -bench . -count 6")

La caché de módulos y de compilación vive en volúmenes con nombre, para no bajar el mundo en cada
ejecución; aun así, el primer `go build` con un perfil nuevo recompila todo (está en la doc de PGO).
"""
import subprocess
import sys
from datetime import datetime, timezone
from pathlib import Path

sys.stdout.reconfigure(encoding="utf-8", errors="replace")
VERSION = "1.27.2"
IMAGEN = f"golang:{VERSION}"


class Banco:
    def __init__(self, aqui: Path, nombre: str, proyecto: str = "proyecto", salida: str = "SALIDA.md"):
        self.aqui, self.nombre, self.salida = aqui, nombre, salida
        self.proyecto = aqui / proyecto
        self.partes = []

    def __enter__(self):
        return self

    def __exit__(self, tipo, *_):
        if tipo is None:
            cpu = self.sh("nproc && grep -m1 'model name' /proc/cpuinfo | cut -d: -f2",
                          guardar=False)
            cab = [f"# SALIDA — {self.nombre} (ejecutado)", "",
                   f"Generado por `python {Path(sys.argv[0]).name}`. **No se edita a mano.**", "",
                   f"- fecha: {datetime.now(timezone.utc):%Y-%m-%d %H:%M} UTC",
                   f"- imagen: {IMAGEN} (Docker, linux/amd64)",
                   f"- máquina: {' · '.join(l.strip() for l in cpu.splitlines())} núcleos/CPU",
                   "- los tiempos son de ESTA máquina, dentro de Docker en Windows."]
            (self.aqui / self.salida).write_text("\n".join(cab + self.partes) + "\n",
                                                 encoding="utf-8")
            print(f"→ {self.salida}")

    def sh(self, orden, mostrar=None, guardar=True, env=None, timeout=1800):
        cmd = ["docker", "run", "--rm", "-v", f"{self.proyecto}:/w", "-w", "/w",
               "-v", "fab-go-mod:/go/pkg/mod", "-v", "fab-go-cache:/root/.cache/go-build"]
        for k, v in (env or {}).items():
            cmd += ["-e", f"{k}={v}"]
        cmd += [IMAGEN, "bash", "-c", orden]
        r = subprocess.run(cmd, capture_output=True, text=True, encoding="utf-8",
                           errors="replace", timeout=timeout)
        txt = (r.stdout + r.stderr).strip()
        if guardar:
            self.anotar(mostrar or orden, txt)
        return txt

    def anotar(self, mostrar, txt):
        self.partes.extend(["", f"`$ {mostrar}`", "", "```", txt or "(sin salida)", "```"])
        print(f"$ {mostrar}\n{txt}\n", flush=True)

    def titulo(self, t):
        self.partes.extend(["", f"## {t}"])
        print(f"\n=== {t}", flush=True)

    def nota(self, t):
        self.partes.extend(["", f"> {t}"])
