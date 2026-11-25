# -*- coding: utf-8 -*-
"""synctest — AMPLIACIÓN del banco (golandia, vídeo 366, 2026-10-11). Escribe ampliacion/SALIDA.md.

Aparte del banco principal para no rehacer sus tiempos. ampliacion/proyecto: reintentos con espera
exponencial (1+2+4+8 s) en sus tres caminos, el limitador de golang.org/x/time/rate, y un ticker de
una hora durante un año entero, todo dentro de synctest.Test.

    python correr.py
"""
import sys
from pathlib import Path

AQUI = Path(__file__).resolve().parent
sys.path.insert(0, str(AQUI.parent.parent))
from banco_go import Banco  # noqa: E402

with Banco(AQUI, "synctest — ampliación") as b:
    b.sh("go get golang.org/x/time@v0.14.0 > /dev/null 2>&1; go mod tidy > /dev/null 2>&1; grep golang.org/x/time go.mod", mostrar="go get golang.org/x/time")
    b.titulo("1 · Reintentos con espera exponencial, sus tres caminos")
    b.sh("go test -v -count=1 -run 'Acierta|Rinde|Cancelado' . 2>&1 | grep -E 'reintentos_test|^(ok|FAIL|---)'",
         mostrar="go test -v -run 'Acierta|Rinde|Cancelado'")
    b.titulo("2 · Una librería que no sabe nada: golang.org/x/time/rate")
    b.sh("go test -v -count=1 -run Limitador . 2>&1 | grep -E 'reintentos_test|^(ok|FAIL)'", mostrar="go test -v -run Limitador")
    b.titulo("3 · Una tarea cada hora, durante un año")
    b.sh("go test -v -count=1 -run UnAno . 2>&1 | grep -E 'reintentos_test|^(ok|FAIL)|PASS: TestUnAno'", mostrar="go test -v -run UnAno")
    b.titulo("4 · Un debounce con time.AfterFunc")
    b.sh("go test -v -count=1 -run Debounce . 2>&1 | grep -E 'debounce_test|^(ok|FAIL)|PASS: TestDebounce'", mostrar="go test -v -run Debounce")
    b.titulo("5 · Toda la ampliación, mil veces")
    b.sh("go test -count=1000 . 2>&1 | tail -1", mostrar="go test -count=1000 .   # 6 tests × 1.000")

