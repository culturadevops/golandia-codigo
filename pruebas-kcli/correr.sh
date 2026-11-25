#!/bin/sh
set -e
cd "$(dirname "$0")"
docker build -t kcli-lab .
docker run --rm kcli-lab > salida-real.txt 2>&1
cat salida-real.txt
