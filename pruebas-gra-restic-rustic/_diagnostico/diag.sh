#!/bin/bash
# diag.sh — ¿por qué rustic gasta el triple de CPU en el backup inicial? Variantes de su configuración contra restic.
# Mismo origen (/base/datos, solo lectura), repo nuevo en cada variante; GNU time.
export RESTIC_PASSWORD=x RUSTIC_PASSWORD=x RUSTIC_NO_PROGRESS=true RESTIC_CACHE_DIR=/tmp/rc RUSTIC_CACHE_DIR=/tmp/uc
cat /base/datos/binarios/* /base/datos/logs/* > /dev/null; find /base/datos -type f -exec cat {} + > /dev/null   # caché caliente para todos
v() { local n="$1"; shift; rm -rf /repo/* /tmp/rc /tmp/uc; "$@" init > /dev/null 2>&1 || true; }
corre() {  # corre <nombre> <init...> -- <backup...>
  local n="$1"; shift; local ini=(); while [ "$1" != "--" ]; do ini+=("$1"); shift; done; shift
  rm -rf /repo/* /tmp/rc /tmp/uc
  "${ini[@]}" > /tmp/i 2>&1 || { echo "$n init falló"; tail -3 /tmp/i; }
  /usr/bin/time -f "$n seg=%e user=%U sys=%S rss_kb=%M" "$@" > /tmp/o 2>/tmp/t; tail -1 /tmp/t
  echo "   repo $(du -sm /repo | cut -f1) MiB · $(grep -h -E 'stored|Added to the repo' /tmp/o /tmp/t | head -1 | sed 's/^ *//')"
}
for vuelta in 1 2; do
  echo "## vuelta $vuelta"
  corre "restic (auto, por defecto)" env RESTIC_REPOSITORY=/repo restic init -- env RESTIC_REPOSITORY=/repo restic backup --host b /base/datos
  corre "rustic (por defecto)" env RUSTIC_REPOSITORY=/repo rustic init -- env RUSTIC_REPOSITORY=/repo rustic backup --host b /base/datos
  corre "rustic --set-extra-verify false" env RUSTIC_REPOSITORY=/repo rustic init --set-extra-verify false -- env RUSTIC_REPOSITORY=/repo rustic backup --host b /base/datos
  corre "rustic extra-verify false + compresión 1" env RUSTIC_REPOSITORY=/repo rustic init --set-extra-verify false --set-compression 1 -- env RUSTIC_REPOSITORY=/repo rustic backup --host b /base/datos
  corre "restic --compression max" env RESTIC_REPOSITORY=/repo restic init -- env RESTIC_REPOSITORY=/repo restic backup --compression max --host b /base/datos
done
echo "## lo que guarda rustic en su config por defecto"
rm -rf /repo/*; RUSTIC_REPOSITORY=/repo rustic init > /dev/null 2>&1; RESTIC_REPOSITORY=/repo RESTIC_PASSWORD=x restic cat config
