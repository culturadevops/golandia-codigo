#!/bin/sh
set -e
cd "$(dirname "$0")"
docker build -t generics-lab .
docker run --rm generics-lab > salida-real.txt 2>&1
cat salida-real.txt
