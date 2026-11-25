#!/usr/bin/env bash
# demo.sh — la medición del vídeo 232 «Fyne o Wails: lo medí».
#
# LA MISMA APP en los dos: una ventana con un contador, un botón que suma y una lista de
# cinco elementos. Ni una función de más en ninguno.
#
# ⛔ MISMA IMAGEN BASE (golang:1.25-bookworm) para los dos lados, y MISMA CACHE DE MODULOS
#    en un volumen compartido. Solo cambian los paquetes de sistema que exige cada uno.
#
# ⚠️ CINCO CORRECCIONES DE LAS PRIMERAS CORRIDAS, todas mías:
#  1. Git Bash convierte `-w /app` en `C:/Program Files/Git/app`. docker fallaba y el script
#     salía con código 0: "pasó" sin compilar nada. → MSYS_NO_PATHCONV=1.
#  2. Fyne NO compila con lo que dice su propia documentación. Falta Wayland.
#  3. Wails 2.15 exige Go >= 1.25; con 1.24 ni resuelve el módulo.
#  4. ⛔ EL GRANDE: `go build` a secas sobre un proyecto de Wails compila SIN el frontend de
#     escritorio. Sale un binario de 4,9 MB que enlaza solo libc, no abre ninguna ventana
#     y TERMINA CON CODIGO 0. Yo lo di por bueno y publiqué "Wails pesa 6 veces menos":
#     estaba comparando una app contra un programa vacío. Hacen falta las etiquetas
#     `desktop,production`. Por eso este banco COMPRUEBA que la ventana se sostiene ANTES
#     de dar ningún número.
#  5. El "compilar en frío" incluía descargar los módulos (26 contra 6): eso medía mi red.
#     Ahora la descarga se mide aparte y la compilación se cronometra con los módulos ya ahí.
set -u
export MSYS_NO_PATHCONV=1
D=$(cd "$(dirname "$0")" && pwd -W)
TAGS_W="desktop,production,webkit2_41"
docker volume create gomod >/dev/null

fyne()  { docker run --rm -v gomod:/go/pkg/mod -v "$D/app-fyne:/app"  -w /app banco-fyne  bash -c "$1"; }
wails() { docker run --rm -v gomod:/go/pkg/mod -v "$D/app-wails:/app" -w /app banco-wails bash -c "$1"; }

echo "== 0 · ⛔ ¿ABREN LOS DOS UNA VENTANA DE VERDAD? (si no, no se compara nada) =="
comprueba() {
  local q=$1 tags=$2 nom=$3
  $q "
    go build ${tags:+-tags $tags} -o /tmp/c . >/dev/null 2>&1 || { echo '   $nom: NO COMPILA'; exit 1; }
    n=\$(ldd /tmp/c | wc -l)
    Xvfb :99 -screen 0 1024x768x24 >/dev/null 2>&1 &
    sleep 2
    DISPLAY=:99 /tmp/c >/tmp/run.log 2>&1 &
    P=\$!; sleep 8
    if kill -0 \$P 2>/dev/null; then
      echo \"   $nom: ventana VIVA tras 8 s · enlaza \$n librerias del sistema\"
    else
      echo \"   $nom: ⛔ LA VENTANA NO SE SOSTUVO (salio con \$?). Su salida:\"; head -5 /tmp/run.log
      exit 1
    fi"
}
comprueba fyne  ""        "Fyne " || exit 1
comprueba wails "$TAGS_W" "Wails" || exit 1

echo
echo "== 1 · LINEAS DE CODIGO para la misma app =="
echo -n "   Fyne  (main.go):                      "; grep -vcE '^\s*(//|$)' app-fyne/main.go
W1=$(grep -vcE '^\s*(//|$)' app-wails/main.go)
W2=$(grep -vcE '^\s*(//|$)' app-wails/app.go)
W3=$(grep -vcE '^\s*$' app-wails/frontend/dist/index.html)
echo "   Wails (main.go + app.go + index.html): $W1 + $W2 + $W3 = $((W1+W2+W3))"
echo "   ⚠️ y Wails necesita ademas wails.json y la carpeta frontend/dist"

echo
echo "== 2 · PAQUETES DE SISTEMA que exige cada uno (simulado sobre la base limpia) =="
for lado in "Fyne :libgl1-mesa-dev xorg-dev libwayland-dev libxkbcommon-dev wayland-protocols" \
            "Wails:libgtk-3-dev libwebkit2gtk-4.1-dev"; do
  n=${lado%%:*}; pk=${lado#*:}
  docker run --rm golang:1.25-bookworm bash -c "
    apt-get update -qq >/dev/null 2>&1
    inst=\$(apt-get install -s $pk 2>/dev/null | grep -c '^Inst ')
    mb=\$(apt-get install --print-uris -y $pk 2>/dev/null | grep -oE \"' [0-9]+ \" | tr -d \"' \" \
          | awk '{s+=\$1} END {printf \"%.0f\", s/1048576}')
    echo \"   $n: \$inst paquetes · \${mb:-?} MB de descarga\"
  "
done

echo
echo "== 3 · MODULOS EN EL GRAFO =="
echo -n "   Fyne : "; fyne  'go list -m all 2>/dev/null | wc -l'
echo -n "   Wails: "; wails 'go list -m all 2>/dev/null | wc -l'

echo
echo "== 4 · COMPILACION (modulos YA descargados: aqui no se mide la red) =="
echo -n "   Fyne : "; fyne  'go mod download >/dev/null 2>&1; go clean -cache >/dev/null 2>&1; s=$(date +%s%N); go build -o /tmp/c . || exit 1; e=$(date +%s%N); echo "$(( (e-s)/1000000 )) ms"'
echo -n "   Wails: "; wails "go mod download >/dev/null 2>&1; go clean -cache >/dev/null 2>&1; s=\$(date +%s%N); go build -tags $TAGS_W -o /tmp/c . || exit 1; e=\$(date +%s%N); echo \"\$(( (e-s)/1000000 )) ms\""

echo
echo "== 5 · 🥇 TAMANO DEL BINARIO =="
echo -n "   Fyne : "; fyne  'go build -o /tmp/c . && stat -c "%s bytes" /tmp/c'
echo -n "   Wails: "; wails "go build -tags $TAGS_W -o /tmp/c . && stat -c '%s bytes' /tmp/c"

echo
echo "== 6 · 🥇 LIBRERIAS DEL SISTEMA que necesita EN MARCHA =="
echo -n "   Fyne : "; fyne  'go build -o /tmp/c . >/dev/null 2>&1; ldd /tmp/c | wc -l'
fyne  'go build -o /tmp/c . >/dev/null 2>&1; ldd /tmp/c | grep -oE "lib(GL|X11|xkbcommon|wayland-client)\.so[0-9.]*" | sort -u | sed "s/^/      /"'
echo -n "   Wails: "; wails "go build -tags $TAGS_W -o /tmp/c . >/dev/null 2>&1; ldd /tmp/c | wc -l"
wails "go build -tags $TAGS_W -o /tmp/c . >/dev/null 2>&1; ldd /tmp/c | grep -oE 'lib(webkit2gtk[a-z0-9.-]*|gtk-3|javascriptcoregtk[a-z0-9.-]*|soup-3.0)\.so[0-9.]*' | sort -u | sed 's/^/      /'"

echo
echo "== 7 · ⛔ COMPILACION CRUZADA A WINDOWS desde Linux (las dos con cgo) =="
echo -n "   Fyne : "; fyne  'if GOOS=windows GOARCH=amd64 go build -o /tmp/c.exe . 2>/tmp/e; then echo "cruza"; else echo "NO cruza: $(head -3 /tmp/e | tail -1 | cut -c1-90)"; fi'
echo -n "   Wails: "; wails "if GOOS=windows GOARCH=amd64 go build -tags $TAGS_W -o /tmp/c.exe . 2>/tmp/e; then echo 'cruza'; else echo \"NO cruza: \$(head -3 /tmp/e | tail -1 | cut -c1-90)\"; fi"

echo
echo "== 8 · 🥇 MEMORIA CON LA VENTANA ABIERTA (Xvfb, mismo metodo, ARBOL COMPLETO) =="
mide_ram() {
  local q=$1 tags=$2 nom=$3
  $q "
    go build ${tags:+-tags $tags} -o /tmp/c . >/dev/null 2>&1
    Xvfb :99 -screen 0 1024x768x24 >/dev/null 2>&1 &
    sleep 2
    DISPLAY=:99 /tmp/c >/dev/null 2>&1 &
    P=\$!; sleep 8
    kill -0 \$P 2>/dev/null || { echo '   $nom: murio'; exit 1; }
    tot=0
    for p in \$(pgrep -P \$P 2>/dev/null) \$P; do
      r=\$(awk '/VmRSS/{print \$2}' /proc/\$p/status 2>/dev/null); tot=\$((tot+\${r:-0}))
    done
    echo \"   $nom: \$((tot/1024)) MB · arbol de \$(( \$(pgrep -P \$P 2>/dev/null | wc -l) + 1 )) procesos\""
}
mide_ram fyne  ""        "Fyne "
mide_ram wails "$TAGS_W" "Wails"

echo
echo "== 9 · TAMANO DEL ENTORNO DE COMPILACION =="
docker images --format '   {{.Repository}}: {{.Size}}' | grep banco-
echo
echo "== fin =="
