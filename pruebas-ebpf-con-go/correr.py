# -*- coding: utf-8 -*-
"""eBPF con Go — banco ejecutable (golandia, vídeo 384). Escribe SALIDA.md.

Un programa eBPF en C (proyecto/conexiones.c) enganchado al tracepoint sock/inet_sock_set_state, y
un programa en Go con cilium/ebpf que lo carga, lee cada conexión TCP nueva por un ring buffer y al
final lee el mapa que el kernel fue llenando. Corre en un contenedor PRIVILEGIADO con la imagen
imagen/Dockerfile (golang:1.27.2 + clang + libbpf-dev), sobre el kernel de WSL2 de Docker Desktop.

    python correr.py
"""
import subprocess
import sys
from pathlib import Path

AQUI = Path(__file__).resolve().parent
sys.path.insert(0, str(AQUI.parent))
import banco_go  # noqa: E402

IMAGEN = "fab-go-ebpf:1.27.2"
banco_go.IMAGEN = IMAGEN


class BancoPriv(banco_go.Banco):
    def sh(self, orden, mostrar=None, guardar=True, env=None, timeout=1800, privilegiado=True):
        cmd = ["docker", "run", "--rm"] + (["--privileged"] if privilegiado else []) + [
            "-v", f"{self.proyecto}:/w", "-w", "/w",
            "-v", "fab-go-mod:/go/pkg/mod", "-v", "fab-go-cache:/root/.cache/go-build", IMAGEN, "bash", "-c", orden]
        r = subprocess.run(cmd, capture_output=True, text=True, encoding="utf-8", errors="replace", timeout=timeout)
        txt = (r.stdout + r.stderr).strip()
        if guardar:
            self.anotar(mostrar or orden, txt)
        return txt


TRACEFS = "mount -t tracefs nodev /sys/kernel/tracing 2>/dev/null; "
SERVIDOR = "go build -o /tmp/servidor ./servidor && (/tmp/servidor &) && sleep 1 && "

with BancoPriv(AQUI, "eBPF con Go") as b:
    b.titulo("El kernel: ¿soporta eBPF?")
    b.sh("uname -r && ls -l /sys/kernel/btf/vmlinux | awk '{print $5, $9}' && "
         "zcat /proc/config.gz | grep -E '^CONFIG_(BPF_SYSCALL|BPF_JIT|DEBUG_INFO_BTF)='",
         mostrar="uname -r && ls -l /sys/kernel/btf/vmlinux && zcat /proc/config.gz | grep BPF")

    b.titulo("Sin escribir nada: el tracepoint en crudo")
    b.sh(TRACEFS + "E=/sys/kernel/tracing/events/sock/inet_sock_set_state; echo > /sys/kernel/tracing/trace; echo 1 > $E/enable; "
         + SERVIDOR + "curl -s 127.0.0.1:8080 > /dev/null; sleep 1; echo 0 > $E/enable; "
         "grep 'dport=8080' /sys/kernel/tracing/trace | sed 's/.*inet_sock_set_state: //' | "
         "sed -E 's/ saddrv6=[^ ]+ daddrv6=[^ ]+//; s/ protocol=IPPROTO_TCP//' | head -3",
         mostrar="echo 1 > /sys/kernel/tracing/events/sock/inet_sock_set_state/enable\n$ curl 127.0.0.1:8080\n$ cat /sys/kernel/tracing/trace")

    b.titulo("Compilar: bpf2go + go build")
    b.sh("go version && grep cilium go.mod && head -3 main.go | tail -1 && "
         "go generate ./... 2>&1 | grep -iE 'error|warning' ; go build -o conexiones . && "
         "ls -l conexiones_bpfel.o conexiones_bpfel.go conexiones | awk '{print $5, $9}'",
         mostrar="go generate   # bpf2go compila conexiones.c con clang\n$ go build -o conexiones .")

    b.titulo("El verificador del kernel")
    b.sh(TRACEFS + "mkdir -p /tmp/malo && cp go.mod go.sum main.go /tmp/malo/ && "
         "sed -e 's|__u16 puerto = saliente ? dport : sport;|__u16 puerto = saliente ? a->dport : a->sport;|' "
         "-e 's|e->sport = sport;|e->sport = a->sport;|' -e 's|e->dport = dport;|e->dport = a->dport;|' conexiones.c > /tmp/malo/conexiones.c && "
         "cd /tmp/malo && go generate ./... > /dev/null 2>&1 && go build -o malo . && ./malo -durante 1s 2>&1 | head -1 | cut -c1-200",
         mostrar="__u16 puerto = saliente ? a->dport : a->sport;   # elegir directamente sobre el contexto\n$ ./conexiones")

    b.titulo("Ver las conexiones de la máquina")
    b.sh(TRACEFS + SERVIDOR + "(./conexiones -durante 6s > /tmp/out.txt 2>/tmp/err.txt &) && sleep 2 && "
         "for i in 1 2 3; do curl -s 127.0.0.1:8080 > /dev/null; done; curl -s -o /dev/null -m 3 https://1.1.1.1; "
         "sleep 6; cat /tmp/err.txt /tmp/out.txt",
         mostrar="./conexiones -durante 6s   # y, mientras, 3 curl al servidor local y 1 a 1.1.1.1")

    b.titulo("Una conexión que falla")
    b.sh(TRACEFS + "(./conexiones -durante 4s > /tmp/out.txt 2>/dev/null &) && sleep 1 && "
         "curl -s -m 2 127.0.0.1:9; echo \"curl: código $?\"; sleep 4; grep -E ':9$|curl:|puerto 9 ' /tmp/out.txt",
         mostrar="./conexiones -durante 4s   # y curl a 127.0.0.1:9, donde no escucha nadie")

    b.titulo("Otro contenedor, que no sabe nada")
    otro = subprocess.Popen(["docker", "run", "--rm", "curlimages/curl:8.11.1", "sh", "-c",
                             "sleep 4; curl -s -o /dev/null -m 3 https://1.1.1.1; curl -s -o /dev/null -m 3 http://example.com"],
                            stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
    b.sh(TRACEFS + "./conexiones -durante 10s 2>/dev/null | grep -E 'curl|---|puerto (80|443) '",
         mostrar="./conexiones -durante 10s   # mientras OTRO contenedor hace curl a 1.1.1.1 y example.com")
    otro.wait()

    b.titulo("Sin privilegios")
    b.sh("./conexiones -durante 2s 2>&1 | head -2", privilegiado=False,
         mostrar="docker run (sin --privileged) ... ./conexiones")
