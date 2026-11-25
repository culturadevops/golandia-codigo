#!/bin/sh
set -e
cd "$(dirname "$0")"
docker build -t goreleaser-lab .
docker run --rm goreleaser-lab > salida-real.txt 2>&1
cat salida-real.txt
