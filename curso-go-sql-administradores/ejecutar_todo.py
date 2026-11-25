# -*- coding: utf-8 -*-
"""Levanta un Postgres limpio en Docker, corre las 16 lecciones y regenera SALIDA.md.

La regla de la casa: nada de lo que sale en el video se escribe a mano. Sale de aqui.
Hace falta Go 1.24+, Docker y Python 3.10+.

    python ejecutar_todo.py            # y al terminar borra el contenedor
    python ejecutar_todo.py --dejar    # lo deja corriendo, para jugar con psql
"""
import os
import platform
import subprocess
import sys
import tempfile
import time
from datetime import datetime, timezone
from pathlib import Path

AQUI = Path(__file__).resolve().parent
ENV = dict(os.environ, PYTHONIOENCODING="utf-8")
CONTENEDOR = "curso-sql-pg"
IMAGEN = "postgres:17-alpine"
DSN = "postgres://curso:curso@127.0.0.1:55432/tienda?sslmode=disable"

LECCIONES = [
    ("01", "Conectar: sql.Open no conecta", "01_conectar"),
    ("02", "Una fila y sql.ErrNoRows", "02_una_fila"),
    ("03", "Muchas filas y rows.Err", "03_muchas_filas"),
    ("04", "NULL", "04_null"),
    ("05", "rows sin cerrar: el pool que se agota", "05_rows_sin_cerrar"),
    ("06", "El pool, medido", "06_pool"),
    ("07", "Inyeccion SQL y placeholders", "07_inyeccion"),
    ("08", "Exec y RowsAffected", "08_rows_affected"),
    ("09", "Transacciones", "09_transacciones"),
    ("10", "context y timeouts", "10_context"),
    ("11", "La actualizacion perdida", "11_actualizacion_perdida"),
    ("12", "Bloqueos: la migracion que tumba las lecturas", "12_bloqueos"),
    ("13", "Carga masiva", "13_carga_masiva"),
    ("14", "Migraciones", "14_migraciones"),
    ("15", "Fechas: timestamp contra timestamptz", "15_tiempo"),
]


def correr(cmd, cwd=AQUI, timeout=600):
    r = subprocess.run(cmd, capture_output=True, text=True, encoding="utf-8",
                       errors="replace", cwd=cwd, env=ENV, timeout=timeout)
    return r.returncode, (r.stdout or "") + (r.stderr or "")


def levantar_postgres():
    subprocess.run(["docker", "rm", "-f", CONTENEDOR], capture_output=True)
    subprocess.run(["docker", "run", "-d", "--name", CONTENEDOR,
                    "-e", "POSTGRES_USER=curso", "-e", "POSTGRES_PASSWORD=curso",
                    "-e", "POSTGRES_DB=tienda", "-p", "55432:5432", IMAGEN],
                   check=True, capture_output=True)
    # pg_isready dice «listo» durante el arranque temporal del contenedor: se espera a que
    # conteste una consulta de verdad, y dos veces seguidas.
    seguidas = 0
    for _ in range(60):
        c, _ = correr(["docker", "exec", CONTENEDOR, "psql", "-U", "curso", "-d", "tienda",
                       "-h", "127.0.0.1", "-c", "select 1"])
        seguidas = seguidas + 1 if c == 0 else 0
        if seguidas == 2:
            return
        time.sleep(1)
    raise SystemExit("Postgres no arranco")


def proyecto():
    """La lección 16: pgsalud contra un escenario con problemas, y sin ellos."""
    tmp = Path(tempfile.mkdtemp(prefix="curso-sql-"))
    exe = tmp / ("pgsalud.exe" if os.name == "nt" else "pgsalud")
    subprocess.run(["go", "build", "-o", str(exe), "./16_proyecto"], cwd=AQUI, check=True)
    partes, fallos = [], 0

    def paso(titulo, args, esperado):
        nonlocal fallos
        c, s = correr([str(exe), *args])
        fallos += c != esperado
        partes.extend([f"### {titulo}", "", f"`pgsalud {' '.join(args)}` · codigo de salida `{c}`",
                       "", "```", s.rstrip(), "```", ""])

    esc = subprocess.Popen(["go", "run", "./16_proyecto/escenario", "-durante", "12s"], cwd=AQUI,
                           stdout=subprocess.PIPE, stderr=subprocess.STDOUT, text=True, env=ENV)
    esc.stdout.readline()  # «escenario listo»
    time.sleep(3)
    paso("Con problemas", ["-max-consulta", "2s", "-max-idle-tx", "2s"], 1)
    paso("Lo mismo, en JSON", ["-max-consulta", "2s", "-max-idle-tx", "2s", "-json"], 1)
    esc.wait(timeout=60)
    time.sleep(1)
    paso("Sin problemas", [], 0)
    paso("Sin base de datos", ["-dsn", "postgres://curso:curso@127.0.0.1:1/tienda?sslmode=disable&connect_timeout=2"], 2)
    return partes, fallos


def main():
    sello = datetime.now(timezone.utc).strftime("%Y-%m-%d %H:%M:%SZ")
    print("  levantando Postgres limpio ...", flush=True)
    levantar_postgres()
    _, gov = correr(["go", "version"])
    _, pgv = correr(["docker", "exec", CONTENEDOR, "postgres", "--version"])
    partes = ["# SALIDA — curso «Bases de datos SQL desde Go, para administradores»", "",
              "> Generado por `ejecutar_todo.py`. **No se escribe a mano.**", "",
              f"- fecha: `{sello}`", f"- {gov.strip()}", f"- servidor: `{pgv.strip()}` (imagen `{IMAGEN}`)",
              "- driver: `github.com/jackc/pgx/v5` a traves de `database/sql`",
              f"- sistema: `{platform.system()} {platform.release()}` · Postgres en Docker", ""]
    fallos = 0
    for num, titulo, d in LECCIONES:
        print(f"  [{num}] {titulo} ...", end=" ", flush=True)
        codigo, salida = correr(["go", "run", f"./{d}"])
        fallos += codigo != 0
        partes += [f"## Leccion {num} — {titulo}", "",
                   f"`go run ./{d}` · codigo de salida `{codigo}`", "", "```", salida.strip("\n"), "```", ""]
        print("ok" if codigo == 0 else f"FALLO (codigo {codigo})")

    print("  [16] El proyecto: pgsalud ...", end=" ", flush=True)
    p, f = proyecto()
    fallos += f
    partes += ["## Leccion 16 — El proyecto: pgsalud", "", *p]
    print("ok" if not f else f"FALLO ({f})")

    (AQUI / "SALIDA.md").write_text("\n".join(partes), encoding="utf-8")
    if "--dejar" not in sys.argv:
        subprocess.run(["docker", "rm", "-f", CONTENEDOR], capture_output=True)
    print(f"\n  -> SALIDA.md regenerado (16 lecciones, {fallos} fallos)")
    return 1 if fallos else 0


if __name__ == "__main__":
    sys.exit(main())
