#!/bin/bash
# El mismo ssh, N a la vez con xargs -P. Uso: paralelo.sh hosts.txt N
uno() {
  s=$(timeout 10 ssh -o BatchMode=yes -o ConnectTimeout=3 admin@"$1" revisar 2>&1)
  echo "$1 $? $s"
}
export -f uno
xargs -P "$2" -I{} bash -c 'uno {}' < "$1"
