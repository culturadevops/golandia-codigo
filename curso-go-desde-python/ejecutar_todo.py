# -*- coding: utf-8 -*-
"""Corre las 12 lecciones (Go y, donde la hay, su comparacion en Python) y regenera SALIDA.md.

La regla de la casa: nada de lo que sale en el video se escribe a mano. Sale de aqui.
Hace falta Go 1.22+ y Python 3.10+.

    python ejecutar_todo.py
"""
import os
import platform
import statistics
import subprocess
import sys
import tempfile
import time
from datetime import datetime, timezone
from pathlib import Path

AQUI = Path(__file__).resolve().parent
ENV = dict(os.environ, PYTHONIOENCODING="utf-8")

LECCIONES = [
    ("01", "De script a binario", "01_binario"),
    ("02", "Tipos y el valor cero", "02_valor_cero"),
    ("03", "Slices que comparten memoria", "03_slices"),
    ("04", "Mapas sin orden", "04_mapas"),
    ("05", "Cadenas: bytes y letras", "05_cadenas"),
    ("06", "Errores como valores", "06_errores"),
    ("07", "El nil que no es nil", "07_nil"),
    ("08", "Structs y el receptor", "08_structs"),
    ("09", "defer", "09_defer"),
    ("10", "Goroutines que no esperas", "10_goroutines"),
    ("11", "Treinta chequeos en paralelo", "11_paralelo"),
    ("12", "El proyecto, en Go", "12_proyecto"),
]


def correr(cmd, cwd=AQUI):
    r = subprocess.run(cmd, capture_output=True, text=True, encoding="utf-8",
                       errors="replace", cwd=cwd, env=ENV, timeout=600)
    return r.returncode, (r.stdout or "") + (r.stderr or "")


def arranque(cmd, veces=20):
    """Mediana de lo que tarda en arrancar, hacer su trabajo y salir."""
    t = []
    for _ in range(veces):
        t0 = time.perf_counter()
        subprocess.run(cmd, capture_output=True, cwd=AQUI, env=ENV)
        t.append((time.perf_counter() - t0) * 1000)
    return statistics.median(t)


def medir_binario():
    tmp = Path(tempfile.mkdtemp(prefix="curso-go-"))
    exe = tmp / ("revisar.exe" if os.name == "nt" else "revisar")
    subprocess.run(["go", "build", "-o", str(exe), "./12_proyecto"], cwd=AQUI, check=True)
    exe_min = tmp / ("revisar-min.exe" if os.name == "nt" else "revisar-min")
    subprocess.run(["go", "build", "-ldflags=-s -w", "-o", str(exe_min), "./12_proyecto"],
                   cwd=AQUI, check=True)
    go_ms = arranque([str(exe), "-json"])
    py_ms = arranque([sys.executable, "12_proyecto/revisar_servicios.py", "--json"])
    return [
        "== Medida: el mismo programa, en Go y en Python ==",
        f"  binario de Go                 {exe.stat().st_size / 1e6:6.1f} MB",
        f"  binario de Go (-ldflags=-s -w) {exe_min.stat().st_size / 1e6:5.1f} MB",
        f"  script de Python              {(AQUI / '12_proyecto/revisar_servicios.py').stat().st_size / 1e3:6.1f} KB"
        "  + un interprete instalado",
        f"  arranque de Go     (mediana de 20)  {go_ms:6.1f} ms",
        f"  arranque de Python (mediana de 20)  {py_ms:6.1f} ms",
        f"  -> Go arranca {py_ms / go_ms:.1f} veces antes",
    ]


def main():
    sello = datetime.now(timezone.utc).strftime("%Y-%m-%d %H:%M:%SZ")
    _, gov = correr(["go", "version"])
    partes = ["# SALIDA — curso «Go para quien ya sabe Python»", "",
              "> Generado por `ejecutar_todo.py`. **No se escribe a mano.**", "",
              f"- fecha: `{sello}`", f"- {gov.strip()}",
              f"- Python: `{platform.python_version()}`",
              f"- sistema: `{platform.system()} {platform.release()}`", ""]
    fallos = 0
    for num, titulo, d in LECCIONES:
        print(f"  [{num}] {titulo} ...", end=" ", flush=True)
        codigo, salida = correr(["go", "run", f"./{d}"])
        esperado = 1 if num == "12" else 0
        ok = codigo == esperado
        fallos += not ok
        partes += [f"## Leccion {num} — {titulo}", "",
                   f"`go run ./{d}` · codigo de salida `{codigo}`", "", "```", salida.rstrip(), "```", ""]
        py = AQUI / d / "python.py"
        if py.exists():
            c2, s2 = correr([sys.executable, str(py)])
            fallos += c2 != 0
            partes += ["Y lo mismo en Python:", "", "```", s2.rstrip(), "```", ""]
        if num == "12":
            partes += ["```", *medir_binario(), "```", ""]
        print("ok" if ok else f"FALLO (codigo {codigo})")
    (AQUI / "SALIDA.md").write_text("\n".join(partes), encoding="utf-8")
    print(f"\n  -> SALIDA.md regenerado ({len(LECCIONES)} lecciones, {fallos} fallos)")
    return 1 if fallos else 0


if __name__ == "__main__":
    sys.exit(main())
