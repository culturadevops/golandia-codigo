# -*- coding: utf-8 -*-
"""correr.py — banco del vídeo «Concurrencia en Go: goroutines, canales y select, medido» (golandia).

Cada carpeta es un programa; todos se ejecutan en golang:1.27.2 dentro de Docker (ver memoria
go-tests-en-docker). El contenedor se limita a 8 CPU para que la prueba de cálculo tenga un techo claro.

    python correr.py      -> SALIDA.md
"""
import datetime
import os
import subprocess
from pathlib import Path

AQUI = Path(__file__).resolve().parent
ENV = dict(os.environ, MSYS_NO_PATHCONV="1")
GO = "golang:1.27.2"
salida = []


def p(t=""):
    print(t, flush=True)
    salida.append(t)


def go(carpeta, *args, race=False, mostrar=None):
    cmd = ["go", "run"] + (["-race"] if race else []) + ["."] + list(args)
    p(f"$ {mostrar or ' '.join(cmd)}")
    r = subprocess.run(["docker", "run", "--rm", "--cpus", "8", "-v", f"{AQUI.as_posix()}:/w", "-v", "fab-go-cache:/root/.cache",
                        "-w", f"/w/{carpeta}", "-e", "CGO_ENABLED=1" if race else "CGO_ENABLED=0", GO, "sh", "-c",
                        "[ -f go.mod ] || go mod init ejemplo.com/" + carpeta + " >/dev/null 2>&1; " + " ".join(cmd)],
                       capture_output=True, text=True, encoding="utf-8", errors="replace", env=ENV)
    lineas = (r.stdout + r.stderr).strip().splitlines()
    for l in lineas[:14]:
        p("  " + l if not l.startswith("  ") else l)
    if len(lineas) > 14:
        p(f"  … ({len(lineas) - 14} líneas más)")
    if r.returncode:
        p(f"  [código de salida {r.returncode}]")


def main():
    cab = ["# SALIDA — «Concurrencia en Go: goroutines, canales y select, medido» (golandia)", "",
           "> Generado por `correr.py`. **No se escribe a mano.**", "",
           f"- fecha: `{datetime.datetime.now(datetime.timezone.utc):%Y-%m-%d %H:%M:%SZ}`",
           f"- {GO} en Docker, limitado a 8 CPU (`--cpus 8`); una sola máquina, los tiempos son de esta."]
    salida.append("\n```")
    p("== 1 · Cuánto cuesta una goroutine ==")
    for n in ("1000", "100000", "1000000"):
        go("coste", n, mostrar=f"go run . {n}")
    p("")
    p("== 2 · Canales sin búfer y con búfer ==")
    go("canales")
    p("")
    p("== 3 · Un pool de trabajadores ==")
    go("pool")
    p("")
    p("== 4 · Una carrera de datos ==")
    for modo in ("sin", "mutex", "atomic"):
        go("carrera", modo, mostrar=f"go run . {modo}")
    go("carrera", "sin", race=True, mostrar="go run -race . sin")
    p("")
    p("== 5 · Un envío que nadie recibe ==")
    go("bloqueo")
    p("")
    p("== 6 · La goroutine que se queda colgada ==")
    for modo in ("fuga", "bufer", "contexto"):
        go("fuga", modo, mostrar=f"go run . {modo}")
    p("")
    p("== 7 · Cancelar con context ==")
    go("cancelar")
    salida.append("```")
    (AQUI / "SALIDA.md").write_bytes(("\n".join(cab) + "\n" + "\n".join(salida) + "\n").encode("utf-8"))
    print("\n[ok] SALIDA.md")


if __name__ == "__main__":
    main()
