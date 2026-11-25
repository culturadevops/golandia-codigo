#!/bin/sh
# Tamaño del binario: hola mundo y el servidor de la serie (el mismo main.go), con go build y con tinygo build.
cd /tmp && cp -r /d/demos/hola h && cp -r /srvgo s
cd /tmp/h
CGO_ENABLED=0 go build -ldflags='-s -w' -o /tmp/out/hola-go . 2>/dev/null || { mkdir -p /tmp/out; CGO_ENABLED=0 go build -ldflags='-s -w' -o /tmp/out/hola-go .; }
tinygo build -no-debug -o /tmp/out/hola-tinygo .
GOOS=wasip1 GOARCH=wasm go build -ldflags='-s -w' -o /tmp/out/hola-go.wasm .
tinygo build -no-debug -target=wasip1 -o /tmp/out/hola-tinygo.wasm .
cd /tmp/s
CGO_ENABLED=0 go build -ldflags='-s -w' -o /tmp/out/servidor-go .
tinygo build -no-debug -o /tmp/out/servidor-tinygo .
tinygo build -no-debug -opt=2 -o /tmp/out/servidor-tinygo-opt2 .
echo "### tinygo version"; tinygo version
echo "### ls -l (bytes)"; cd /tmp/out && ls -l | awk 'NR>1 {printf "%10d  %s\n", $5, $9}'
echo "### ./hola-go && ./hola-tinygo"; ./hola-go; ./hola-tinygo
