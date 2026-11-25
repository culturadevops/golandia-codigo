#!/bin/sh
cd /d/demos
echo "### ruby hilos.rb"; ruby hilos.rb 2>&1; echo "[código de salida: $?]"
echo "### ruby --yjit hilos.rb"; ruby --yjit hilos.rb 2>&1; echo "[código de salida: $?]"
