# -*- coding: utf-8 -*-
"""¿Go si eres administrador de sistemas? — banco ejecutable (golandia, vídeo 388). Escribe SALIDA.md.

  1  50 servidores simulados (3 colgados): revisar /salud con Bash + curl, uno detrás de otro
  2  lo mismo con Go, todos a la vez (una goroutine por servidor)
  3  el binario: tamaño, y sin dependencias (CGO_ENABLED=0 → ldd: not a dynamic executable)
  4  compilación cruzada: Linux arm64 y Windows desde la misma máquina; cuántos destinos hay
  5  el binario en una imagen que NO tiene Go (httpd:2.4-alpine): funciona igual

    python correr.py
"""
import subprocess
import sys
from pathlib import Path

AQUI = Path(__file__).resolve().parent
sys.path.insert(0, str(AQUI.parent))
from banco_go import Banco  # noqa: E402

SIM = "go build -o /tmp/sim ./cmd/simulador && (/tmp/sim &) && sleep 1 && "

with Banco(AQUI, "Go para administradores de sistemas") as b:
    b.titulo("La versión de Go")
    b.sh("go version")

    b.titulo("1 · Bash + curl, uno detrás de otro")
    b.sh(SIM + "bash revisar.sh", mostrar="bash revisar.sh")

    b.titulo("1b · Bash también puede ir en paralelo")
    b.sh(SIM + "bash paralelo.sh", mostrar="bash paralelo.sh   # xargs -P 50 curl -m 2")

    b.titulo("2 · Go, todos a la vez")
    b.sh(SIM + "go run ./cmd/revisar", mostrar="go run ./cmd/revisar")

    b.titulo("2b · Como mucho 10 a la vez")
    b.sh(SIM + "go build -o /tmp/r ./cmd/revisar && /tmp/r -max 10", mostrar="./revisar -max 10")

    b.titulo("2c · Para otra herramienta, y para cron")
    b.sh(SIM + "go build -o /tmp/r ./cmd/revisar && /tmp/r -json 2>/dev/null | head -c 160; echo ' …'",
         mostrar="./revisar -json 2>/dev/null | head -c 160")
    b.sh(SIM + "go build -o /tmp/r ./cmd/revisar && /tmp/r > /dev/null 2>&1; echo \"código de salida: $?\"",
         mostrar="./revisar > /dev/null 2>&1; echo \"código de salida: $?\"")

    b.titulo("3 · El binario")
    b.sh("CGO_ENABLED=0 go build -o revisar ./cmd/revisar && ls -lh revisar | awk '{print $5, $9}' && ldd revisar",
         mostrar="CGO_ENABLED=0 go build -o revisar ./cmd/revisar && ls -lh revisar && ldd revisar")
    b.sh("CGO_ENABLED=0 go build -ldflags='-s -w' -o revisar-min ./cmd/revisar && ls -lh revisar-min | awk '{print $5, $9}'",
         mostrar="go build -ldflags='-s -w' -o revisar-min ./cmd/revisar && ls -lh revisar-min")

    b.titulo("4 · Compilación cruzada")
    b.sh("CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o revisar-arm64 ./cmd/revisar && "
         "CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o revisar.exe ./cmd/revisar && "
         "ls -lh revisar-arm64 revisar.exe | awk '{print $5, $9}'",
         mostrar="GOOS=linux GOARCH=arm64 go build -o revisar-arm64 ./cmd/revisar && "
                 "GOOS=windows GOARCH=amd64 go build -o revisar.exe ./cmd/revisar && ls -lh revisar-arm64 revisar.exe")
    b.sh("go tool dist list | wc -l", mostrar="go tool dist list | wc -l")

    b.titulo("5 · En un servidor sin Go")
    b.sh("CGO_ENABLED=0 go build -o sim ./cmd/simulador", guardar=False)
    cmd = ["docker", "run", "--rm", "-v", f"{AQUI / 'proyecto'}:/w", "-w", "/w", "httpd:2.4-alpine", "sh", "-c",
           "which go || echo 'go: no está instalado'; (./sim &); sleep 1; ./revisar"]
    r = subprocess.run(cmd, capture_output=True, text=True, encoding="utf-8", errors="replace")
    b.anotar("docker run httpd:2.4-alpine  →  which go; ./revisar", (r.stdout + r.stderr).strip())
    b.sh("rm -f revisar revisar-min revisar-arm64 revisar.exe sim", guardar=False)
