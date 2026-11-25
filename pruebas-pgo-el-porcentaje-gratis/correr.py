# -*- coding: utf-8 -*-
"""PGO en Go — banco ejecutable (golandia, vídeo 365). Escribe SALIDA.md.

Lo que enseña, todo medido:
  1-4  benchmark con y sin PGO (benchstat) y POR QUÉ: desvirtualiza las llamadas por interfaz
       más calientes y luego las inlinea (salida de -gcflags=-m, diferencia entre las dos).
  5    la trampa: el perfil sacado de `go test` NO encaja con el binario (`ejemplo.com/evaluador.X`
       en el test, `main.X` en el binario): con él, `go build` no cambia nada.
  6    lo que dice la doc: el perfil se saca del PROGRAMA de verdad; default.pgo en el paquete
       main, `go build` lo coge solo y el binario lo deja escrito.
  7    lo que cuesta: tamaño y primera compilación con caché vacía.
  8    cuándo no sirve: un perfil que no es representativo.

    python correr.py
"""
import sys
from pathlib import Path

AQUI = Path(__file__).resolve().parent
sys.path.insert(0, str(AQUI.parent))
from banco_go import Banco  # noqa: E402

BENCH = "go test -run x -bench Trabajo -count 6"
STAT = "go run golang.org/x/perf/cmd/benchstat@latest {} 2>/dev/null | grep -v '^goos\\|^goarch\\|^pkg\\|^cpu'"
MS = "s=$(date +%s%N); {}; e=$(date +%s%N); echo \"{}: $(( (e-s)/1000000 )) ms\""

with Banco(AQUI, "PGO en Go") as b:
    b.sh("rm -f *.pgo *.pprof *.txt sin-pgo con-pgo evaluador.test", guardar=False)
    b.titulo("La versión de Go")
    b.sh("go version")

    b.titulo("1 · Benchmark sin PGO")
    b.sh(f"{BENCH} -pgo=off | tee sin.txt | grep Benchmark", mostrar=f"{BENCH} -pgo=off")

    b.titulo("2 · Perfil de CPU del benchmark")
    b.sh("go test -run x -bench Trabajo -cpuprofile test.pprof > /dev/null && ls -l test.pprof | awk '{print $5, $9}'",
         mostrar="go test -run x -bench Trabajo -cpuprofile test.pprof")

    b.titulo("3 · Benchmark con PGO")
    b.sh(f"{BENCH} -pgo=test.pprof | tee con.txt | grep Benchmark", mostrar=f"{BENCH} -pgo=test.pprof")
    b.sh(STAT.format("sin.txt con.txt"), mostrar="benchstat sin.txt con.txt")

    b.titulo("4 · Por qué: lo que cambia en el compilador")
    b.sh("export GOCACHE=/tmp/c1; go test -c -o /dev/null -pgo=off -gcflags=-m 2>&1 | sort > /tmp/a; "
         "export GOCACHE=/tmp/c2; go test -c -o /dev/null -pgo=test.pprof -gcflags=-m 2>&1 | sort > /tmp/b; "
         "diff /tmp/a /tmp/b | grep '^>' | grep -E 'devirtualizing|inlining call to (Func|Suma)' | sed 's/^> //'",
         mostrar="diff  <(go test -c -pgo=off -gcflags=-m)  <(go test -c -pgo=test.pprof -gcflags=-m)")

    b.titulo("5 · La trampa: ese perfil no encaja con el binario")
    b.sh("go tool pprof -top test.pprof 2>/dev/null | grep -E 'Eval' | head -3 | awk '{print $NF}'",
         mostrar="go tool pprof -top test.pprof | grep Eval")
    b.sh("export GOCACHE=/tmp/c3; go build -o /dev/null -pgo=test.pprof -gcflags=-m . 2>&1 | grep -c devirtualizing",
         mostrar="go build -pgo=test.pprof -gcflags=-m . | grep -c devirtualizing")

    b.titulo("6 · El perfil del programa de verdad")
    b.sh("go build -pgo=off -o sin-pgo . && ./sin-pgo -perfil default.pgo",
         mostrar="go build -pgo=off -o sin-pgo . && ./sin-pgo -perfil default.pgo")
    b.sh("go tool pprof -top default.pgo 2>/dev/null | grep -E 'Eval' | head -3 | awk '{print $NF}'",
         mostrar="go tool pprof -top default.pgo | grep Eval")
    b.sh("go build -o con-pgo . && go version -m con-pgo | grep -- '-pgo'",
         mostrar="go build -o con-pgo .      # default.pgo está en el paquete main\n$ go version -m con-pgo | grep pgo")
    b.sh("for i in 1 2 3; do ./sin-pgo; done; echo ---; for i in 1 2 3; do ./con-pgo; done",
         mostrar="for i in 1 2 3; do ./sin-pgo; done;  for i in 1 2 3; do ./con-pgo; done")

    b.titulo("7 · Lo que cuesta")
    b.sh("ls -l sin-pgo con-pgo | awk '{print $5, $9}'", mostrar="ls -l sin-pgo con-pgo")
    b.sh("export GOCACHE=/tmp/v1; " + MS.format("go build -pgo=off -o /tmp/a .", "sin PGO, caché vacía")
         + "; export GOCACHE=/tmp/v2; " + MS.format("go build -pgo=default.pgo -o /tmp/b .", "con PGO, caché vacía"),
         mostrar="time go build -pgo=off   vs   time go build -pgo=default.pgo   # caché vacía")

    b.titulo("8 · Cuándo no sirve: un perfil que no es representativo")
    b.sh("go test -run x -bench OtraCosa -cpuprofile otro.pprof > /dev/null && "
         f"{BENCH} -pgo=otro.pprof | tee mal.txt | grep Benchmark",
         mostrar=f"go test -bench OtraCosa -cpuprofile otro.pprof\n$ {BENCH} -pgo=otro.pprof")
    b.sh(STAT.format("sin.txt mal.txt"), mostrar="benchstat sin.txt mal.txt")
    b.sh("rm -f sin-pgo con-pgo evaluador.test otro.pprof test.pprof", guardar=False)
