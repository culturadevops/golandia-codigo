# -*- coding: utf-8 -*-
"""PGO en Go — SEGUNDA ampliación del banco (golandia, vídeo 365, 2026-10-11). Escribe ampliacion2/SALIDA.md.

Aparte, como la primera, para no rehacer los tiempos que ya están en el guion. Trabaja sobre una COPIA
de ../proyecto dentro del contenedor (montado de solo lectura):
  1  estabilidad de fuente: 10 líneas añadidas arriba del fichero, y Suma.Eval reescrita en 3 líneas;
  2  estabilidad iterativa: perfilar el binario YA optimizado con PGO y recompilar, tres generaciones.

    python correr.py
"""
import sys
from pathlib import Path

AQUI = Path(__file__).resolve().parent
sys.path.insert(0, str(AQUI.parent.parent))
from banco_go import Banco  # noqa: E402

COPIA = "cp -r /w /tmp/p && cd /tmp/p && rm -f evaluador sin-pgo con-pgo && "
CUENTA = 'cuenta(){ go build -o /dev/null -pgo=$1 -gcflags=-m . 2>&1 | grep -c "$2"; }; '
# Suma.Eval de 1 línea a 3, con perl (sin pelearse con las comillas de bash -c).
TRES_LINEAS = ("perl -0pi -e 's/func \(s Suma\) Eval\(x \[\]float64\) float64 \{ return s\.A\.Eval\(x\) \+ s\.B\.Eval\(x\) \}/"
               "func (s Suma) Eval(x []float64) float64 {\n\ta := s.A.Eval(x)\n\treturn a + s.B.Eval(x)\n}/' evaluador.go; ")
GENS = ("go build -pgo=off -o sin . && go build -pgo=default.pgo -o gen1 . && ./gen1 -perfil g1.pprof > /dev/null && "
        "go build -pgo=g1.pprof -o gen2 . && ./gen2 -perfil g2.pprof > /dev/null && go build -pgo=g2.pprof -o gen3 . && ")


class BancoRO(Banco):
    def sh(self, orden, mostrar=None, guardar=True, env=None, timeout=1800):
        import subprocess
        cmd = ["docker", "run", "--rm", "-v", f"{self.proyecto}:/w:ro", "-w", "/w",
               "-v", "fab-go-mod:/go/pkg/mod", "-v", "fab-go-cache:/root/.cache/go-build", "golang:1.27.2",
               "bash", "-c", orden]
        r = subprocess.run(cmd, capture_output=True, text=True, encoding="utf-8", errors="replace", timeout=timeout)
        txt = (r.stdout + r.stderr).strip()
        if guardar:
            self.anotar(mostrar or orden, txt)
        return txt


with BancoRO(AQUI, "PGO en Go — segunda ampliación", proyecto="../proyecto") as b:
    b.titulo("1 · Estabilidad de fuente: el código cambia, el perfil no")
    b.sh(COPIA + CUENTA + 'echo "original:                    $(cuenta default.pgo "PGO devirtualizing") desvirtualizadas"; '
         '(for l in a b c d e f g h i j; do echo "// $l"; done; cat /w/evaluador.go) > evaluador.go; '
         'echo "10 líneas añadidas arriba:    $(cuenta default.pgo "PGO devirtualizing") desvirtualizadas"; '
         'cp /w/evaluador.go evaluador.go; ' + TRES_LINEAS +
         'echo "Suma.Eval en 3 líneas:       $(cuenta default.pgo "PGO devirtualizing") desvirtualizadas"; '
         'go build -o /dev/null -pgo=default.pgo -gcflags=-m . 2>&1 | grep "PGO devirtualizing" | sed "s#^./##"',
         mostrar="go build -gcflags=-m -pgo=default.pgo   # el mismo perfil, con el código original y cambiado")
    b.sh(COPIA + TRES_LINEAS + 'sed -n "/func (s Suma)/,/^}/p" evaluador.go', mostrar="# Suma.Eval reescrita en 3 líneas")

    b.titulo("2 · Estabilidad iterativa: perfilar el binario ya optimizado")
    b.sh(COPIA + GENS + CUENTA +
         'echo "gen1 (perfil del binario sin PGO): $(cuenta default.pgo "PGO devirtualizing") desvirtualizadas · $(cuenta default.pgo "inlining call") inlineadas"; '
         'echo "gen2 (perfil de gen1):             $(cuenta g1.pprof "PGO devirtualizing") desvirtualizadas · $(cuenta g1.pprof "inlining call") inlineadas"; '
         'echo "gen3 (perfil de gen2):             $(cuenta g2.pprof "PGO devirtualizing") desvirtualizadas · $(cuenta g2.pprof "inlining call") inlineadas"; '
         'echo ---; for x in sin gen1 gen2 gen3; do echo "$x: $(for i in 1 2 3; do ./$x | grep -o "en .*"; done | tr "\n" " ")"; done',
         mostrar="./gen1 -perfil g1.pprof; go build -pgo=g1.pprof -o gen2 .; ./gen2 -perfil g2.pprof; go build -pgo=g2.pprof -o gen3 .\n"
                 "$ for x in sin gen1 gen2 gen3; do for i in 1 2 3; do ./$x; done; done")
    b.sh(COPIA + GENS + 'for p in default.pgo g1.pprof; do echo "== $p"; go tool pprof -top -nodecount=6 $p 2>/dev/null | sed -n "/flat%/,\$p"; done',
         mostrar="go tool pprof -top default.pgo   vs   go tool pprof -top g1.pprof")
    b.sh(COPIA + 'for p in default.pgo; do go build -o /dev/null -pgo=default.pgo -gcflags=-m . 2>&1 | grep -E "inlining call|devirtualizing" | sort > /tmp/g1.txt; done; '
         'go build -pgo=default.pgo -o gen1 . && ./gen1 -perfil g1.pprof > /dev/null && '
         'go build -o /dev/null -pgo=g1.pprof -gcflags=-m . 2>&1 | grep -E "inlining call|devirtualizing" | sort > /tmp/g2.txt; '
         'diff /tmp/g1.txt /tmp/g2.txt | grep "^[<>]" | sed "s#^< #gen1: #; s#^> #gen2: #; s#./evaluador.go:##"',
         mostrar="diff  <(go build -gcflags=-m -pgo=default.pgo)  <(go build -gcflags=-m -pgo=g1.pprof)   # qué cambia")
