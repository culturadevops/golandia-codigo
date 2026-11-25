# -*- coding: utf-8 -*-
"""PGO en Go — AMPLIACIÓN del banco (golandia, vídeo 365, 2026-10-10). Escribe ampliacion/SALIDA.md.

Va aparte del banco principal a propósito: volver a correr aquel cambiaría todos los tiempos de
benchmark que ya están en el guion. Este usa el mismo proyecto (../proyecto) y su default.pgo:
  1  lo que dice `go help build` del flag -pgo;
  2  qué hay dentro de un perfil (tamaño y funciones más calientes);
  3  cuántas llamadas desvirtualiza e inlinea el compilador con el perfil del programa;
  4  fusionar varios perfiles con go tool pprof -proto;
  5  sacar el perfil de un servicio en marcha con net/http/pprof.

    python correr.py
"""
import sys
from pathlib import Path

AQUI = Path(__file__).resolve().parent
sys.path.insert(0, str(AQUI.parent.parent))
from banco_go import Banco  # noqa: E402

SRV = r'''package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
)

func fib(n int) int {
	if n < 2 {
		return n
	}
	return fib(n-1) + fib(n-2)
}

func main() {
	http.HandleFunc("/trabajo", func(w http.ResponseWriter, r *http.Request) { fmt.Fprintln(w, fib(27)) })
	http.ListenAndServe("127.0.0.1:6060", nil)
}
'''

with Banco(AQUI, "PGO en Go — ampliación", proyecto="../proyecto") as b:
    b.titulo("1 · Lo que dice go help build de -pgo")
    b.sh("go help build | sed -n '/^\t-pgo file/,/^\t-[a-z]/p' | head -n -1", mostrar="go help build   # el flag -pgo")

    b.titulo("2 · Qué hay dentro de un perfil")
    b.sh("ls -l default.pgo | awk '{print $5, $9}' && go tool pprof -top -nodecount=5 default.pgo 2>/dev/null | sed -n '1,12p'",
         mostrar="ls -l default.pgo && go tool pprof -top -nodecount=5 default.pgo")

    b.titulo("3 · Lo que hace el compilador con el perfil del programa")
    b.sh("export GOCACHE=/tmp/a; echo \"sin perfil:  $(go build -o /dev/null -pgo=off -gcflags=-m . 2>&1 | grep -c devirtualizing) desvirtualizadas · $(go build -o /dev/null -pgo=off -gcflags=-m . 2>&1 | grep -c 'inlining call') llamadas inlineadas\"; "
         "export GOCACHE=/tmp/b; echo \"con perfil:  $(go build -o /dev/null -pgo=default.pgo -gcflags=-m . 2>&1 | grep -c devirtualizing) desvirtualizadas · $(go build -o /dev/null -pgo=default.pgo -gcflags=-m . 2>&1 | grep -c 'inlining call') llamadas inlineadas\"",
         mostrar="go build -gcflags=-m -pgo=off   vs   -pgo=default.pgo   # contar lo que dice el compilador")

    b.titulo("4 · Fusionar varios perfiles")
    b.sh("go build -pgo=off -o /tmp/sin . && /tmp/sin -perfil /tmp/p1.pprof > /dev/null && /tmp/sin -perfil /tmp/p2.pprof > /dev/null && "
         "go tool pprof -proto /tmp/p1.pprof /tmp/p2.pprof > /tmp/fusion.pgo 2>/dev/null && ls -l /tmp/p1.pprof /tmp/p2.pprof /tmp/fusion.pgo | awk '{print $5, $9}' && "
         "for f in p1.pprof p2.pprof fusion.pgo; do echo \"$f: $(go tool pprof -top /tmp/$f 2>/dev/null | grep -m1 'Total samples')\"; done",
         mostrar="./sin -perfil p1.pprof; ./sin -perfil p2.pprof\n$ go tool pprof -proto p1.pprof p2.pprof > fusion.pgo")

    b.titulo("5 · El perfil de un servicio en marcha: net/http/pprof")
    b.sh("mkdir -p /tmp/srv && cd /tmp/srv && printf 'module srv\n\ngo 1.27\n' > go.mod && cat > main.go <<'X'\n" + SRV + "X\n"
         "go build -o srv . && (./srv &) && sleep 1 && "
         "(for i in $(seq 1 400); do curl -s localhost:6060/trabajo > /dev/null; done &) && "
         "curl -s 'localhost:6060/debug/pprof/profile?seconds=3' -o cpu.pprof && ls -l cpu.pprof | awk '{print $5, $9}' && "
         "go tool pprof -top -nodecount=3 cpu.pprof 2>/dev/null | sed -n '/flat%/,$p'",
         mostrar="import _ \"net/http/pprof\"   # en el servicio\n$ curl 'localhost:6060/debug/pprof/profile?seconds=3' -o cpu.pprof\n$ go tool pprof -top -nodecount=3 cpu.pprof")
