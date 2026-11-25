#!/bin/sh
# Compilar el servidor en frío (sin servidor de compilación, con el compilador ya descargado) y tras cambiar una línea.
cp /d/src/Servidor.scala /tmp/ && cd /tmp
scala-cli compile --server=false Servidor.scala > /dev/null 2>&1   # descarga el compilador: no se cronometra
rm -rf .scala-build
t0=$(date +%s%N); scala-cli compile --server=false Servidor.scala > /dev/null 2>&1; c=$?; t1=$(date +%s%N)
echo "### scala-cli compile (en frío, sin caché del proyecto): $(( (t1 - t0) / 1000000 )) ms [código de salida: $c]"
sed -i 's/"lang":"scala"/"lang":"scala3"/' Servidor.scala
t0=$(date +%s%N); scala-cli compile --server=false Servidor.scala > /dev/null 2>&1; c=$?; t1=$(date +%s%N)
echo "### scala-cli compile (tras cambiar una línea): $(( (t1 - t0) / 1000000 )) ms [código de salida: $c]"
scala-cli compile Servidor.scala > /dev/null 2>&1
sed -i 's/"lang":"scala3"/"lang":"scala"/' Servidor.scala
t0=$(date +%s%N); scala-cli compile Servidor.scala > /dev/null 2>&1; c=$?; t1=$(date +%s%N)
echo "### scala-cli compile con el servidor Bloop ya caliente (tras cambiar una línea): $(( (t1 - t0) / 1000000 )) ms [código de salida: $c]"
