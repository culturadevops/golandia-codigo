# -*- coding: utf-8 -*-
"""testing/synctest — banco ejecutable (golandia, vídeo 366). Escribe SALIDA.md.

Una caché con TTL de 30 s y un limpiador cada 10 s:
  1  el test honesto, que espera de verdad (41 s);
  2  el mismo dentro de synctest.Test: reloj falso, milisegundos; y la hora con la que nace;
  3  un deadlock dentro de la burbuja: synctest lo detecta;
  4  lo que NO avanza el reloj: una goroutine esperando en la red (no es un bloqueo «duradero»).

    python correr.py
"""
import sys
from pathlib import Path

AQUI = Path(__file__).resolve().parent
sys.path.insert(0, str(AQUI.parent))
from banco_go import Banco  # noqa: E402

with Banco(AQUI, "testing/synctest en Go") as b:
    b.titulo("La versión de Go")
    b.sh("go version")

    b.titulo("1 · El test honesto: espera de verdad")
    b.sh("go test -count 1 -run CaducaEsperando -v 2>&1 | grep -E '^(--- |ok|FAIL)'",
         mostrar="go test -count 1 -run CaducaEsperando -v")

    b.titulo("2 · El mismo test dentro de una burbuja")
    b.sh("go test -count 1 -run CaducaSynctest -v 2>&1 | grep -E 'hora|^(--- |ok|FAIL)' | sed 's/^ *//'",
         mostrar="go test -count 1 -run CaducaSynctest -v")

    b.titulo("3 · Un deadlock dentro de la burbuja")
    b.sh("go test -count 1 -run Deadlock 2>&1 | grep -E '^(--- FAIL|panic: deadlock)' | head -2",
         mostrar="go test -count 1 -run Deadlock")

    b.titulo("4 · Lo que no avanza el reloj: esperar en la red")
    b.sh("s=$(date +%s); go test -count 1 -run BloqueadoEnRed -timeout 8s 2>&1 | "
         "grep -E '^panic: test timed out|TestBloqueadoEnRed \\(' | head -2; "
         "e=$(date +%s); echo \"(tardó $((e-s)) s y no terminó solo)\"",
         mostrar="go test -count 1 -run BloqueadoEnRed -timeout 8s")

    b.titulo("5 · Go 1.27: synctest.Sleep")
    b.sh("go test -count 1 -run CaducaSleep127 -v 2>&1 | grep -E 'synctest.Sleep|^(--- |ok|FAIL)' | sed 's/^ *//'",
         mostrar="go test -count 1 -run CaducaSleep127 -v")

    b.titulo("6 · Un timeout de contexto de 5 s")
    b.sh("go test -count 1 -run ContextoTimeout -v 2>&1 | grep -E 'error:|^(--- |ok|FAIL)' | sed 's/^ *//'",
         mostrar="go test -count 1 -run ContextoTimeout -v")

    b.titulo("7 · Un ticker de un segundo durante un minuto")
    b.sh("go test -count 1 -run TickerUnMinuto -v 2>&1 | grep -E 'ticks|^(--- |ok|FAIL)' | sed 's/^ *//'",
         mostrar="go test -count 1 -run TickerUnMinuto -v")

    b.titulo("8 · La solución a la red: net.Pipe")
    b.sh("go test -count 1 -run ConPipe -v 2>&1 | grep -E 'recibido|^(--- |ok|FAIL)' | sed 's/^ *//'",
         mostrar="go test -count 1 -run ConPipe -v")

    b.titulo("9 · Un mutex tampoco deja avanzar el reloj")
    b.sh("s=$(date +%s); go test -count 1 -run BloqueadoEnMutex -timeout 8s 2>&1 | "
         "grep -E '^panic: test timed out|TestBloqueadoEnMutex \(' | head -2; "
         "e=$(date +%s); echo \"(tardó $((e-s)) s y no terminó solo)\"",
         mostrar="go test -count 1 -run BloqueadoEnMutex -timeout 8s")

    b.titulo("10 · Una goroutine que queda colgada al terminar")
    b.sh("go test -count 1 -run FugaAlTerminar 2>&1 | grep -E '^(--- FAIL|panic)' | head -3",
         mostrar="go test -count 1 -run FugaAlTerminar")

    b.titulo("12 · Dos relojes: fuera y dentro")
    b.sh("go test -count 1 -run DosRelojes -v 2>&1 | grep -E 'burbuja|^(--- |ok|FAIL)' | sed 's/^ *//'",
         mostrar="go test -count 1 -run DosRelojes -v")

    b.titulo("13 · Un canal creado fuera de la burbuja")
    b.sh("s=$(date +%s); go test -count 1 -run CanalDeFuera -timeout 8s 2>&1 | "
         "grep -E '^panic|TestCanalDeFuera \(' | head -2; "
         "e=$(date +%s); echo \"(tardó $((e-s)) s)\"",
         mostrar="go test -count 1 -run CanalDeFuera -timeout 8s")

    b.titulo("14 · synctest con el detector de carreras")
    b.sh("go test -count 1 -race -run 'CaducaSynctest|CaducaSleep127|TickerUnMinuto' -v 2>&1 | grep -E '^(--- |ok|FAIL)|DATA RACE'",
         mostrar="go test -count 1 -race -run 'CaducaSynctest|CaducaSleep127|TickerUnMinuto' -v")

    b.titulo("11 · Toda la batería, sin los que se cuelgan")
    b.sh("go test -count 1 -skip 'BloqueadoEn|Deadlock|FugaAlTerminar|CaducaEsperando|CanalDeFuera|DosRelojes' -v 2>&1 | grep -E '^(--- |ok|FAIL)'",
         mostrar="go test -count 1 -skip 'BloqueadoEn|Deadlock|FugaAlTerminar|CaducaEsperando|CanalDeFuera|DosRelojes' -v")

