#!/bin/sh
cp -r /d/demos/go_hilos /tmp/g && cd /tmp/g && go build -o /tmp/hilos . && echo "### ./hilos" && /tmp/hilos; echo "[código de salida: $?]"
