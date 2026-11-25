#!/bin/bash
# Dentro de golang:1.27.2: compila y corre el banco del monitor de uptime (solo librería estándar).
set -e
cd /trabajo/monitor && go vet ./... && go build -o /tmp/monitor .
echo '$ go version'; go version
for m in secuencial ingenuo estados sin-timeout rondas; do
  echo; echo "\$ ./monitor $m"; /tmp/monitor $m
done
echo; echo '$ DISCORD_WEBHOOK_URL= ./monitor entorno; echo "salida: $?"'
set +e; DISCORD_WEBHOOK_URL= /tmp/monitor entorno; echo "salida: $?"
