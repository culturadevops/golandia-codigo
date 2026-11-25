#!/bin/sh
# Compilar el servidor de Go en frío (sin caché de compilación) y tras cambiar una línea.
cp -r /srvgo /tmp/g && cd /tmp/g
export GOCACHE=/tmp/cache
t0=$(date +%s%N); go build -o /tmp/s . 2>&1; c=$?; t1=$(date +%s%N)
echo "### go build (en frío, sin caché): $(( (t1 - t0) / 1000000 )) ms [código de salida: $c]"
sed -i 's/"lang":"go"/"lang":"go2"/' main.go
t0=$(date +%s%N); go build -o /tmp/s . 2>&1; c=$?; t1=$(date +%s%N)
echo "### go build (tras cambiar una línea): $(( (t1 - t0) / 1000000 )) ms [código de salida: $c]"
