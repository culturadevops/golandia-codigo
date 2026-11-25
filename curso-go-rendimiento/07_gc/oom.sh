#!/bin/sh
# Lección 7, segunda parte: un contenedor con 200 MB de memoria (docker run --memory 200m).
# Go no lee ese límite por su cuenta: con GOGC=400 el heap crece hasta que el kernel lo mata.
cd "$(dirname "$0")/.." && . ./comun.sh

paso "cat /sys/fs/cgroup/memory.max"
paso "go build -o \$T/gc ./07_gc"
paso "GOGC=400 \$T/gc; echo \"código de salida: \$?\""
paso "GOGC=400 GOMEMLIMIT=180MiB \$T/gc; echo \"código de salida: \$?\""
