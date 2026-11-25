#!/bin/sh
set -e
cd "$(dirname "$0")"
docker build -t delve-lab .
docker run --rm --cap-add=SYS_PTRACE --security-opt seccomp=unconfined delve-lab > salida-real.txt 2>&1
cat salida-real.txt
