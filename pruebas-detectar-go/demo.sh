#!/usr/bin/env bash
# demo.sh — la medición del vídeo «Por qué el malware se escribe en Go (y cómo se detecta)».
#
# ⛔ DEFENSIVO. Aquí no se escribe nada ofensivo ni se analiza ninguna muestra real. Lo que se
#    analiza son MIS PROPIOS binarios, compilados en este mismo banco. La pregunta es de formato
#    de ejecutable, no de payload.
#
# DOS BLOQUES:
#   A · POR QUÉ lo eligen: se mide la compilación cruzada de verdad, desde una sola máquina.
#   B · CÓMO se detecta: cuántas señales independientes lo delatan, y cuáles sobreviven a -s -w.
#       Con un binario de C compilado con gcc al lado, de contraste.
set -u

cd /lab || exit 1

# ─────────────────────────────────────────────────────────────────────
echo "======================================================================"
echo " A · POR QUÉ. Lo que Go le da a cualquiera que reparta un binario."
echo "======================================================================"
echo
echo "-- A1 · compilación cruzada, desde ESTA máquina (linux/amd64) --"
PLATS="linux/amd64 linux/arm64 windows/amd64 windows/386 darwin/amd64 darwin/arm64 freebsd/amd64 linux/386 linux/arm windows/arm64"
ok=0; fallo=0
t0=$(date +%s%N)
for p in $PLATS; do
  os=${p%%/*}; arch=${p##*/}
  if ( cd /lab/prog && CGO_ENABLED=0 GOOS=$os GOARCH=$arch go build -o "/tmp/h-$os-$arch" . ) 2>/dev/null; then
    ok=$((ok+1))
  else
    fallo=$((fallo+1)); echo "     falló: $p"
  fi
done
t1=$(date +%s%N)
echo "   plataformas compiladas: $ok · fallidas: $fallo"
echo "   tiempo total: $(( (t1-t0)/1000000 )) ms"
echo "   una orden por plataforma. Ni un compilador cruzado instalado."
echo
ls -l /tmp/h-* 2>/dev/null | awk '{printf "     %-28s %s bytes\n", $9, $5}' | head -12

echo
echo "-- A2 · ¿qué necesita la máquina de destino? --"
echo -n "   ldd sobre el binario de Go:  "
ldd /tmp/h-linux-amd64 2>&1 | head -1
gcc -O2 -o /tmp/c-hola /lab/prog-c/hola.c 2>/dev/null
echo    "   ldd sobre el binario de C:"
ldd /tmp/c-hola 2>&1 | sed 's/^/     /' | head -4

echo
echo "-- A3 · tamaño: el precio, que también es una señal --"
echo -n "   Go (hola mundo):  "; stat -c %s /tmp/h-linux-amd64
echo -n "   C  (hola mundo):  "; stat -c %s /tmp/c-hola

# ─────────────────────────────────────────────────────────────────────
echo
echo "======================================================================"
echo " B · CÓMO SE DETECTA. Señales independientes, y cuáles aguantan."
echo "======================================================================"

( cd /lab/prog && go build -o /tmp/go-normal . )
( cd /lab/prog && go build -ldflags "-s -w" -o /tmp/go-strip . )

senal() {   # $1 = etiqueta  $2 = binario
  local n=$1 b=$2
  local s=0
  echo "── $n ──"

  echo -n "   1. la cadena de versión (go1.x)          "
  if strings -a "$b" | grep -qE "^go1\.[0-9]+"; then
    echo "SÍ · $(strings -a "$b" | grep -oE "^go1\.[0-9.]+" | head -1)"; s=$((s+1))
  else echo "no"; fi

  echo -n "   2. la información de compilación         "
  if go version -m "$b" 2>/dev/null | grep -q "^	dep\|^	mod"; then
    echo "SÍ · legible con go version -m"; s=$((s+1))
  else echo "no"; fi

  echo -n "   3. la sección .go.buildinfo              "
  if readelf -S "$b" 2>/dev/null | grep -q "go.buildinfo"; then echo "SÍ"; s=$((s+1)); else echo "no"; fi

  echo -n "   4. secciones propias del lenguaje        "
  local sec
  sec=$(readelf -S "$b" 2>/dev/null | grep -oE "\.go\.[a-z]+|\.noptrdata|\.typelink|\.itablink|\.gopclntab" | sort -u | tr '\n' ' ')
  if [ -n "$sec" ]; then echo "SÍ · $sec"; s=$((s+1)); else echo "no"; fi

  echo -n "   5. mensajes del runtime                  "
  local n5
  n5=$(strings -a "$b" | grep -cE "goroutine|runtime\.|GOMAXPROCS|schedule: |chan receive" || true)
  if [ "$n5" -gt 0 ]; then echo "SÍ · $n5 cadenas"; s=$((s+1)); else echo "no"; fi

  echo -n "   6. nombres de funciones del programa     "
  if strings -a "$b" | grep -q "main\.main\|runtime\.main"; then echo "SÍ"; s=$((s+1)); else echo "no"; fi

  echo -n "   7. nombres de paquetes de la biblioteca  "
  local n7
  n7=$(strings -a "$b" | grep -cE "^(net/http|encoding/json|crypto/tls|os/exec)" || true)
  if [ "$n7" -gt 0 ]; then echo "SÍ · $n7"; s=$((s+1)); else echo "no"; fi

  echo "   >>> $s de 7 señales"
  echo
}

for b in /tmp/go-normal /tmp/go-strip /tmp/c-hola; do
  [ -f "$b" ] || { echo "⛔ NO EXISTE $b — el banco para aquí en vez de decir «0 de 7»"; exit 1; }
done

senal "binario NORMAL"                  /tmp/go-normal
senal "binario con -ldflags \"-s -w\""  /tmp/go-strip
senal "y de contraste: un binario de C" /tmp/c-hola

echo "======================================================================"
echo " B2 · LO QUE UN DEFENSOR SACA DEL BINARIO DESPOJADO"
echo "======================================================================"
echo "\$ go version -m binario-despojado"
go version -m /tmp/go-strip 2>/dev/null | head -8 | sed 's/^/   /'

echo
echo "======================================================================"
echo " B3 · ¿Y DEL BINARIO DE C?"
echo "======================================================================"
echo -n "   ¿dice qué librerías usa?  "; ldd /tmp/c-hola >/dev/null 2>&1 && echo "las enlazadas, sí. Las versiones NO" || echo "no"
echo -n "   ¿dice el compilador?      "; strings -a /tmp/c-hola | grep -qE "GCC:" && echo "SÍ · $(strings -a /tmp/c-hola | grep -oE 'GCC: \(.*\)' | head -1)" || echo "no"
echo -n "   ¿dice sus dependencias?   "; echo "no hay lista: hay que deducirlas"

echo
echo "== fin =="
