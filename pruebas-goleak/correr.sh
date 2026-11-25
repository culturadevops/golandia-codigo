#!/bin/sh
set -e
cd "$(dirname "$0")"
docker build -t goleak-lab .
docker run --rm goleak-lab > salida-real.txt 2>&1
cat salida-real.txt
