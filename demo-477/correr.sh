#!/usr/bin/env bash
# Demo del vídeo 477 «Tu agente de Go en producción (medido)» — se EJECUTA y su salida va a SALIDA.md.
# El agente de soporte en Go (agente/main.go, solo librería estándar) sobre Ollama local: traza JSON · evaluaciones de 1, 3 y 5
# llamadas (v1 y v2 con buscar_pedidos) · las mismas evaluaciones en serie y con 4 goroutines · inyección sin y con guardas.
# La tienda es de mentira: ningún dinero real.
set -u
AQUI="$(cd "$(dirname "$0")" && pwd)"
T=$(mktemp -d)
(cd "$AQUI/agente" && go build -o "$T/agente.exe" .)
A="$T/agente.exe"
echo '$ go version'; go version | cut -d' ' -f3
echo '$ ollama --version'; ollama --version 2>&1 | tail -1
echo '$ ollama list | grep "qwen2.5:"'; ollama list | grep "qwen2.5:" | awk '{print $1, $3, $4}'
echo
echo '$ ./agente traza qwen2.5:7b'
"$A" traza qwen2.5:7b
echo
echo '$ ./agente pasos qwen2.5:1.5b qwen2.5:7b'
"$A" pasos qwen2.5:1.5b qwen2.5:7b
echo
echo '$ ./agente pasos -v2 qwen2.5:1.5b qwen2.5:7b   # + buscar_pedidos(lista)'
"$A" pasos -v2 qwen2.5:1.5b qwen2.5:7b
echo
echo '$ ./agente pasos -v2 -paralelo 4 qwen2.5:7b   # las mismas 30 evaluaciones, 4 a la vez'
"$A" pasos -v2 -paralelo 4 qwen2.5:7b
echo
echo '$ ./agente inyeccion qwen2.5:7b'
"$A" inyeccion qwen2.5:7b
cd / && rm -rf "$T"
