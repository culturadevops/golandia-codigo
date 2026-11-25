#!/bin/bash
# Mediciones con candado: Pyronaut (completo), RPi --tiempos, Azure --tiempos. Sin renders: espera 10 min seguidos sin
# render antes de empezar y, si aparece uno durante una medición, la repite (hasta 3 veces) tras otra espera.
LOCK=<proyecto>/_logs/banco.lock
R=<proyecto>/canales/golandia/recursos
S=<proyecto>/canales/golandia/recursos/pruebas-go-o-pyronaut
VIGIA=$S/vigia-renders.txt
renders() { powershell -NoProfile -c "@(Get-CimInstance Win32_Process -Filter \"Name='node.exe'\" | Where-Object { \$_.CommandLine -match 'remotion' }).Count" | tr -d '\r'; }
esperar_sin_render() { local z=0 n=${1:-12}; while [ $z -lt $n ]; do if [ "$(renders)" = "0" ]; then z=$((z+1)); else z=0; echo "$(date +%T) render activo, espero"; fi; sleep 15; done; }
medir() {  # medir <nombre> <carpeta> <log> <args...>
  local nombre=$1 dir=$2 log=$3; shift 3
  for intento in 1 2 3; do
    esperar_sin_render 40
    : > $VIGIA
    ( while [ -f $S/midiendo ]; do echo "$(date +%T) $(renders)" >> $VIGIA; sleep 15; done ) &
    touch $S/midiendo
    echo "$(date +%T) $nombre (intento $intento)"
    (cd $dir && python correr.py "$@" > $log 2>&1; echo "rc=$?" >> $log)
    rm -f $S/midiendo; sleep 16
    local malas=$(awk '$2!=0' $VIGIA | wc -l)
    echo "$(date +%T) $nombre terminado; muestras con render: $malas"
    [ "$malas" = "0" ] && return 0
  done
  echo "$(date +%T) $nombre: SIEMPRE hubo renders, no vale"
}
export PYTHONUTF8=1
esperar_sin_render 40
until mkdir "$LOCK" 2>/dev/null; do echo "$(date +%T) candado ocupado, espero"; sleep 30; done
trap 'rm -f $S/midiendo; rmdir "$LOCK"' EXIT
medir pyronaut $R/pruebas-go-o-pyronaut _log-correr.txt
medir rpi $R/pruebas-go-raspberry-pi-emulada _log-tiempos.txt --tiempos
medir azure $R/pruebas-go-azure-functions _log-tiempos.txt --tiempos
echo "$(date +%T) fin"
