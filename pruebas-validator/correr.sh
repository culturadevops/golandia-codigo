#!/bin/sh
set -e
cd "$(dirname "$0")"
docker build -t validator-lab .
docker run --rm validator-lab > salida-real.txt 2>&1
cat salida-real.txt
