#!/bin/sh
set -e
cd "$(dirname "$0")"
docker build -t wasm-lab .
docker run --rm wasm-lab > salida-real.txt 2>&1
cat salida-real.txt
