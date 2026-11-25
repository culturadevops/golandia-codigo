#!/usr/bin/env bash
# Demo del vídeo 474 «Proyecto de IA 2 en Go» — se EJECUTA y su salida va a SALIDA.md.
# medir-llm (Go, solo librería estándar) contra tres modelos locales de Ollama, con GPU y forzando solo CPU (num_gpu=0).
# Y, para comparar el cliente, el medidor en Python del vídeo 472 sobre el mismo modelo (mismo prompt y opciones).
set -u
AQUI="$(cd "$(dirname "$0")" && pwd)"
T=$(mktemp -d)
(cd "$AQUI/medir-llm" && go build -o "$T/medir-llm.exe" .)
echo '$ go version'; go version | cut -d' ' -f3
echo '$ ollama --version'; ollama --version 2>&1 | tail -1
echo '$ nvidia-smi --query-gpu=name,memory.total --format=csv,noheader'; nvidia-smi --query-gpu=name,memory.total --format=csv,noheader
echo '$ ls -l medir-llm | awk "{print \$5}"   # el binario, en bytes'; ls -l "$T/medir-llm.exe" | awk '{print $5}'
echo
echo '$ ./medir-llm -rondas 5 qwen2.5-coder:1.5b qwen2.5-coder:3b qwen2.5-coder:7b'
"$T/medir-llm.exe" -rondas 5 qwen2.5-coder:1.5b qwen2.5-coder:3b qwen2.5-coder:7b
echo
echo '$ ./medir-llm -rondas 5 -cpu qwen2.5-coder:1.5b qwen2.5-coder:3b qwen2.5-coder:7b'
"$T/medir-llm.exe" -rondas 5 -cpu qwen2.5-coder:1.5b qwen2.5-coder:3b qwen2.5-coder:7b
echo
echo '$ # el mismo modelo con el cliente de Python (vídeo 472), para comparar clientes'
echo '$ python medir_llm.py qwen2.5-coder:7b --rondas 5'
PYTHONUTF8=1 python "$AQUI/../../../cultura-devops/recursos/demo-472/medir_llm.py" qwen2.5-coder:7b --rondas 5
echo '$ ./medir-llm -rondas 5 qwen2.5-coder:7b'
"$T/medir-llm.exe" -rondas 5 qwen2.5-coder:7b
cd / && rm -rf "$T"
