#!/bin/bash
# El bucle de siempre: un ssh detrás de otro. Uso: bucle.sh hosts.txt
for h in $(cat "$1"); do
  s=$(timeout 10 ssh -o BatchMode=yes -o ConnectTimeout=3 admin@"$h" revisar 2>&1)
  echo "$h $? $s"
done
