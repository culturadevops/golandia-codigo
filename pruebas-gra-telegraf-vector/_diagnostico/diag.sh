#!/bin/bash
# diag.sh — ¿dónde se le va el tiempo a Vector? cada variante con el mismo millón de líneas; reloj y CPU con GNU time.
mkdir -p /datos /salida /tmp/vector; cp /base/eventos.json /datos/; sync
for v in v4-stdin-blackhole v3-stdin-remap-file; do
  rm -rf /salida/* /tmp/vector/*
  /usr/bin/time -f "$v seg=%e cpu=%U+%S rss_kb=%M" vector --quiet --config /diag/$v.yaml < /datos/eventos.json > /dev/null 2>/tmp/err; tail -1 /tmp/err
  [ -f /salida/out.json ] && echo "   salida: $(wc -l < /salida/out.json) líneas"
done
for v in v1-file-blackhole v2-file-remap-blackhole; do
  rm -rf /tmp/vector/*
  ( vector --config /diag/$v.yaml > /tmp/o 2>&1 & echo $! > /tmp/pid; wait ) &
  sleep 1; P=$(cat /tmp/pid); T0=$(date +%s.%N)
  # el blackhole imprime cada segundo cuántos eventos lleva: esperamos al total
  for i in $(seq 1 120); do sleep 1; grep -q -E 'events=(900000|1000000)\b' /tmp/o && break; done
  T1=$(date +%s.%N); read ut st < <(awk '{print $14, $15}' /proc/$P/stat)
  echo "$v seg~$(awk -v a=$T1 -v b=$T0 'BEGIN{printf "%.1f", a-b}') cpu=$(( (ut+st) / $(getconf CLK_TCK) ))s $(grep -o 'events=[0-9]*' /tmp/o | tail -1)"
  kill $P; sleep 1
done
