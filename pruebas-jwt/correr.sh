#!/bin/sh
set -e
cd "$(dirname "$0")"
docker build -t jwt-lab .
docker run --rm jwt-lab > salida-real.txt 2>&1
cat salida-real.txt
