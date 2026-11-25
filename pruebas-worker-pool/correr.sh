#!/bin/sh
set -e
cd "$(dirname "$0")"
docker build -t workerpool-lab .
docker run --rm workerpool-lab > salida-real.txt 2>&1
cat salida-real.txt
