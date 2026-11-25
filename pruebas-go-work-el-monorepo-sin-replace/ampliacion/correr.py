# -*- coding: utf-8 -*-
"""go work — AMPLIACIÓN del banco (golandia, vídeo 367, 2026-10-11). Escribe ampliacion/SALIDA.md.

Aparte del banco principal (ese corre con GOPROXY=off; este necesita bajar golang.org/x/text).
Dos módulos independientes que piden x/text en versiones distintas (app v0.14.0, saludo v0.21.0):
  1  qué versión usa app dentro del workspace;
  2  go.work.sum: el fichero que aparece solo;
  3  go work sync: llevar al go.mod lo que eligió el workspace;
  4  go work vendor.

    python correr.py
"""
import sys
from pathlib import Path

AQUI = Path(__file__).resolve().parent
sys.path.insert(0, str(AQUI.parent.parent))
from banco_go import Banco  # noqa: E402

E = {"GOPROXY": "off"}   # como el banco principal: todo sale de la caché de módulos
with Banco(AQUI, "go work — ampliación") as b:
    b.sh("bash /w/preparar.sh 2>&1 | tail -1", mostrar="bash preparar.sh   # app pide x/text v0.14.0, saludo v0.21.0")
    b.sh("cd repo && grep -H x/text app/go.mod saludo/go.mod", mostrar="grep x/text app/go.mod saludo/go.mod")

    b.titulo("1 · Qué versión usa app dentro del workspace")
    b.sh("cd repo/app && echo \"solo app:      $(GOWORK=off go list -m golang.org/x/text)\" && cd .. && go work init ./app ./saludo && "
         "cd app && echo \"en el workspace: $(go list -m golang.org/x/text)\" && grep x/text go.mod",
         mostrar="GOWORK=off go list -m golang.org/x/text;  go work init ./app ./saludo;  go list -m golang.org/x/text", env=E)
    b.sh("cd repo/app && go build -o /tmp/con . && GOWORK=off go build -o /tmp/sin . && "
         "echo \"compilado en el workspace: $(go version -m /tmp/con | grep -o 'x/text.*' | cut -f1,2)\" && "
         "echo \"compilado solo:            $(go version -m /tmp/sin | grep -o 'x/text.*' | cut -f1,2)\"",
         mostrar="go build && GOWORK=off go build   # y go version -m de cada binario", env=E)

    b.titulo("2 · go.work.sum")
    b.sh("cd repo && git status --short && echo --- && (cat go.work.sum 2>/dev/null | cut -c1-70 || echo '(no hay go.work.sum)')",
         mostrar="git status --short && cat go.work.sum", env=E)

    b.titulo("3 · go work sync")
    b.sh("cd repo && go work sync && git diff -U0 app/go.mod app/go.sum | grep -E '^[-+][^-+]' | cut -c1-70",
         mostrar="go work sync && git diff app/go.mod", env=E)

    b.titulo("4 · go work vendor")
    b.sh("cd repo && go work vendor && ls vendor && grep -c '^# ' vendor/modules.txt && "
         "grep -E '^## workspace|^# ' vendor/modules.txt | head -5 && du -sh vendor | cut -f1",
         mostrar="go work vendor && ls vendor && head vendor/modules.txt", env=E)
