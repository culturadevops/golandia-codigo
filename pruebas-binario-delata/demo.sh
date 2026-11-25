#!/usr/bin/env bash
# demo.sh — la medición del vídeo «Tu binario de Go delata tu código fuente».
#
# El programa de prueba es una herramienta corriente: lee un YAML, usa cobra, y tiene un paquete
# `interno/` con una constante que solo existe en el código privado.
#
# ⛔ Se compila DENTRO de una ruta que parece la de alguien real
#    (/casa/mi-proyecto-privado/app) para poder enseñar si esa ruta acaba dentro del binario.
#
# ⚠️ Las cuatro variantes son las cuatro que la gente prueba de verdad:
#      normal · -trimpath · -ldflags "-s -w" · las dos juntas
#    y de cada una se mira LO MISMO, para que la comparación valga.
set -u

DIR=/casa/mi-proyecto-privado/app
cd "$DIR" || exit 1
SECRETO="proyecto-cascabel-fase-2"

compila() {   # $1 = nombre  $2… = banderas
  local n=$1; shift
  go build "$@" -o "/tmp/$n" . || { echo "no compila: $n"; exit 1; }
}

mira() {      # $1 = nombre legible  $2 = ruta del binario
  local n=$1 b=$2
  echo "── $n ──"
  echo "   tamaño:                 $(stat -c %s "$b") bytes"
  echo -n "   ¿lista de dependencias? "
  if go version -m "$b" >/dev/null 2>&1 && [ "$(go version -m "$b" | grep -c '^	dep')" -gt 0 ]; then
    echo "SÍ · $(go version -m "$b" | grep -c '^	dep') dependencias con versión y hash"
  else
    echo "no"
  fi
  echo -n "   ¿nombre del módulo?     "
  go version -m "$b" 2>/dev/null | awk '/^\tpath/{print $2}' | head -1 | sed 's/^$/no/'
  echo -n "   ¿rutas de mi máquina?   "
  local r
  r=$(strings -a "$b" | grep -c "/casa/mi-proyecto-privado" || true)
  [ "$r" -gt 0 ] && echo "SÍ · $r apariciones de /casa/mi-proyecto-privado" || echo "no"
  echo -n "   ¿la constante privada?  "
  strings -a "$b" | grep -q "$SECRETO" && echo "SÍ · «$SECRETO»" || echo "no"
  # ⛔ CORRECCIÓN: la primera versión contaba «símbolos que empiezan por el nombre del módulo».
  #    Eso medía cosas DISTINTAS en cada lado: sin -trimpath las rutas de fichero son absolutas
  #    (/casa/...) y no empezaban por el módulo; con -trimpath sí. Salía «1 contra 3» y parecía
  #    que -trimpath añadía información. No añade nada: cambia cómo se escribe lo mismo.
  #    Ahora se pregunta por las DOS cosas por separado, y la misma pregunta a los cuatro.
  echo -n "   ¿el nombre de mi función? "
  strings -a "$b" | grep -q "interno\.Saludo" && echo "SÍ · interno.Saludo" || echo "no"
  echo -n "   ¿mis nombres de fichero?  "
  strings -a "$b" | grep -q "secreto\.go" && echo "SÍ · secreto.go" || echo "no"
  echo -n "   ¿se ve que es Go?       "
  strings -a "$b" | grep -q "go1\.[0-9]" && echo "SÍ" || echo "no"
  echo
}

echo "======================================================================"
echo " EL PROGRAMA: una herramienta normal. 3 dependencias y un paquete interno."
echo "======================================================================"
echo "   compilado en: $DIR"
echo "   la constante privada: $SECRETO"
echo

compila normal
compila trimpath -trimpath
compila strip   -ldflags "-s -w"
compila ambos   -trimpath -ldflags "-s -w"

echo "======================================================================"
echo " LAS CUATRO VARIANTES, MIRANDO LO MISMO EN CADA UNA"
echo "======================================================================"
mira "1 · go build              (lo que hace todo el mundo)" /tmp/normal
mira "2 · -trimpath" /tmp/trimpath
mira "3 · -ldflags \"-s -w\"      (lo que la gente llama «despojar»)" /tmp/strip
mira "4 · -trimpath -ldflags \"-s -w\"" /tmp/ambos

echo "======================================================================"
echo " 🥇 LA SALIDA COMPLETA DEL MÁS DESPOJADO DE LOS CUATRO"
echo "======================================================================"
echo "\$ go version -m binario-despojado"
go version -m /tmp/ambos | sed 's/^/   /'

echo
echo "======================================================================"
echo " Y LA RUTA, EN EL BINARIO NORMAL"
echo "======================================================================"
strings -a /tmp/normal | grep -oE "/casa/mi-proyecto-privado[^ \"]*" | sort -u | head -6 | sed 's/^/   /'

echo
echo "======================================================================"
echo " ⛔ ¿SE PUEDE QUITAR LA LISTA DE DEPENDENCIAS?"
echo "======================================================================"
echo "   probando -buildvcs=false:"
go build -trimpath -buildvcs=false -ldflags "-s -w" -o /tmp/sinvcs . 2>/dev/null
echo -n "     dependencias que quedan: "; go version -m /tmp/sinvcs 2>/dev/null | grep -c '^	dep' || echo 0
echo "   probando quitar la sección con el enlazador externo:"
go build -trimpath -ldflags "-s -w -buildid=" -o /tmp/sinid . 2>/dev/null
echo -n "     dependencias que quedan: "; go version -m /tmp/sinid 2>/dev/null | grep -c '^	dep' || echo 0

echo
echo "======================================================================"
echo " LO QUE UNA TRAZA DE PÁNICO ENSEÑA EN CADA CASO"
echo "======================================================================"
cat > /tmp/panico.go <<'GO'
package main

func main() { panic("boom") }
GO
mkdir -p /tmp/p && cp /tmp/panico.go /tmp/p/main.go
( cd /tmp/p && cat > go.mod <<'GOMOD'
module miempresa.interno/panico

go 1.24
GOMOD
  go build -o /tmp/pan-normal . 2>/dev/null
  go build -trimpath -o /tmp/pan-trim . 2>/dev/null )
echo "   sin -trimpath:"
/tmp/pan-normal 2>&1 | grep -oE "/[a-z/._-]+\.go:[0-9]+" | head -2 | sed 's/^/     /'
echo "   con -trimpath:"
/tmp/pan-trim 2>&1 | grep -oE "[a-z0-9/._-]+\.go:[0-9]+" | head -2 | sed 's/^/     /'

echo
echo "== fin =="
