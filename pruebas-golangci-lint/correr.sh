#!/bin/sh
set -e
cd "$(dirname "$0")"
docker build -t lint-lab .
docker run --rm lint-lab > salida-real.txt 2>&1
cat salida-real.txt
