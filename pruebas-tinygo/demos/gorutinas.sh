#!/bin/sh
# 2 goroutines con cálculo puro: ¿usan los 2 núcleos? El mismo archivo con go build y con tinygo build.
cp -r /d/demos/gorutinas /tmp/g && cd /tmp/g
go build -o /tmp/con-go . && tinygo build -no-debug -o /tmp/con-tinygo .
echo "### ./con-go"; /tmp/con-go; echo "[código de salida: $?]"
echo "### ./con-tinygo"; /tmp/con-tinygo; echo "[código de salida: $?]"
