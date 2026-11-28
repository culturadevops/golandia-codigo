#!/bin/bash
# medir.sh <telegraf|vector> <json|nginx> <líneas esperadas en la salida> — corre DENTRO de gl-gra-logs:1 (--cpus 2 --memory 1g).
# Arranca la herramienta como servicio (leyendo el archivo desde el principio), mira cada 0,25 s el tamaño de la salida y
# da por terminado el trabajo cuando la salida tiene las líneas esperadas y lleva 3 s sin crecer. Del proceso se lee, en
# /proc: CPU (utime+stime) y memoria (VmHWM = el pico residente; VmRSS = lo que tiene al final).
set -u
H="$1"; E="$2"; ESP="$3"
rm -rf /salida/* /tmp/vector; mkdir -p /tmp/vector /salida
# los datos se copian a disco del contenedor para que los dos lean del mismo sitio
mkdir -p /datos; [ "$ESP" = 0 ] || cp /base/eventos.json /base/access.log /datos/; sync
CLK=$(getconf CLK_TCK)
ahora() { date +%s.%N; }

OUT=/tmp/log; case "$E" in *stdout) OUT=/salida/out.json;; esac   # A2: la salida es stdout, redirigida al archivo
if [ "$H" = telegraf ]; then
    T0=$(ahora); telegraf --config /banco/telegraf-$E.conf > $OUT 2>/tmp/log &
else
    T0=$(ahora); vector --quiet --config /banco/vector-$E.yaml > $OUT 2>/tmp/log &
fi
PID=$!
if [ "$ESP" = 0 ]; then   # reposo: sin archivo de entrada, 15 s y se mira la memoria
    sleep 15
    read -r ut st < <(awk '{print $14, $15}' /proc/$PID/stat)
    awk -v r="$(awk '/VmRSS/{print $2}' /proc/$PID/status)" -v ut="$ut" -v st="$st" -v clk="$CLK" \
      'BEGIN{printf "REPOSO rss_mib=%.1f cpu_s=%.2f\n", r/1024, (ut+st)/clk}'
    kill $PID; exit 0
fi
ult=0; tult=$T0; tprim=""
while :; do
    sleep 0.25
    t=$(ahora)
    if ! kill -0 $PID 2>/dev/null; then echo "MURIO"; tail -n 20 /tmp/log; exit 1; fi
    s=$(stat -c %s /salida/out.json 2>/dev/null || echo 0)
    [ -z "$tprim" ] && [ "$s" -gt 0 ] && tprim=$t
    if [ "$s" -gt "$ult" ]; then ult=$s; tult=$t; fi
    if [ "$s" -gt 0 ] && awk -v a="$t" -v b="$tult" 'BEGIN{exit !(a-b>3)}'; then
        n=$(wc -l < /salida/out.json)
        [ "$n" -ge "$ESP" ] && break
        awk -v a="$t" -v b="$tult" 'BEGIN{exit !(a-b>30)}' && { echo "PARADO con $n de $ESP líneas"; break; }
    fi
    awk -v a="$t" -v b="$T0" 'BEGIN{exit !(a-b>900)}' && { echo "TIMEOUT interno"; break; }
done
read -r ut st < <(awk '{print $14, $15}' /proc/$PID/stat)
HWM=$(awk '/VmHWM/{print $2}' /proc/$PID/status); RSS=$(awk '/VmRSS/{print $2}' /proc/$PID/status)
HILOS=$(awk '/Threads/{print $2}' /proc/$PID/status)
kill $PID; wait $PID 2>/dev/null
n=$(wc -l < /salida/out.json)
awk -v t0="$T0" -v tp="$tprim" -v tu="$tult" -v ut="$ut" -v st="$st" -v clk="$CLK" -v hwm="$HWM" -v rss="$RSS" -v n="$n" -v h="$HILOS" \
  'BEGIN{printf "RESULTADO lineas=%d seg_total=%.2f seg_primera=%.2f cpu_s=%.2f pico_mib=%.1f final_mib=%.1f hilos=%d\n", n, tu-t0, tp-t0, (ut+st)/clk, hwm/1024, rss/1024, h}'

# ¿hicieron el mismo trabajo? una huella del contenido, igual para los dos aunque cambie la forma del JSON
if [ "${E%%-*}" = json ]; then
    if [ "$H" = telegraf ]; then
        jq -r '[.fields.duracion_ms, .fields.estado, .fields.ruta, .tags.nivel, .tags.entorno] | @tsv' /salida/out.json > /tmp/h
    else
        jq -r '[.duracion_ms, .estado, .ruta, .nivel, .entorno] | @tsv' /salida/out.json > /tmp/h
    fi
else
    if [ "$H" = telegraf ]; then
        jq -r '[.fields.client_ip // .tags.client_ip, (.fields.resp_bytes // 0), (.tags.resp_code // .fields.resp_code), .fields.request, .tags.verb // .fields.verb] | @tsv' /salida/out.json > /tmp/h
    else
        jq -r '(.request | split(" ")) as $r | [.client, (.size // 0), .status, $r[1], $r[0]] | @tsv' /salida/out.json > /tmp/h
    fi
fi
echo "HUELLA $(LC_ALL=C sort /tmp/h | md5sum | cut -c1-12) filas=$(wc -l < /tmp/h) suma_col2=$(awk -F'\t' '{s+=$2} END{printf "%d", s}' /tmp/h)"
head -c 400 /salida/out.json | head -n 1 | sed 's/^/MUESTRA /'
head -n 2 /tmp/h | sed 's/^/FILA /'
