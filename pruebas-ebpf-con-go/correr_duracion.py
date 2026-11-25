# -*- coding: utf-8 -*-
"""eBPF con Go, segunda parte — duración de cada conexión y filtro por proceso. Escribe SALIDA-duracion.md.

proyecto-duracion/: el mismo tracepoint que proyecto/, pero el programa guarda cada socket en un mapa
al nacer (SYN_SENT) y al cerrarse (CLOSE) manda a Go cuánto tardó en conectar y cuánto vivió. Go
escribe dos constantes antes de cargarlo: el PID a filtrar y el espacio de PIDs propio.
Mismo contenedor privilegiado e imagen que correr.py. SALIDA.md (primera parte) no se toca.

    python correr_duracion.py
"""
import sys
from pathlib import Path

AQUI = Path(__file__).resolve().parent
sys.path.insert(0, str(AQUI.parent))
import banco_go  # noqa: E402

banco_go.IMAGEN = "fab-go-ebpf:1.27.2"


class BancoPriv(banco_go.Banco):
    def sh(self, orden, mostrar=None, guardar=True, env=None, timeout=1800, pid_host=False):
        import subprocess
        cmd = ["docker", "run", "--rm", "--privileged"] + (["--pid=host"] if pid_host else []) + [
            "-v", f"{self.proyecto}:/w", "-w", "/w",
            "-v", "fab-go-mod:/go/pkg/mod", "-v", "fab-go-cache:/root/.cache/go-build",
            banco_go.IMAGEN, "bash", "-c", orden]
        r = subprocess.run(cmd, capture_output=True, text=True, encoding="utf-8", errors="replace", timeout=timeout)
        txt = (r.stdout + r.stderr).strip()
        if guardar:
            self.anotar(mostrar or orden, txt)
        return txt


TRACEFS = "mount -t tracefs nodev /sys/kernel/tracing 2>/dev/null; "
SERVIDOR = "go build -o /tmp/servidor ./servidor && (/tmp/servidor &) ; sleep 1; "
# `exec` conserva el PID: el curl elegido espera 2 s (a que el programa eBPF esté cargado) y abre
# una conexión nueva cada medio segundo.
ELEGIDO = ('(sleep 2; exec curl -s -o /dev/null -H "Connection: close" --rate 2/s "127.0.0.1:8080/?ms=[100-500:100]") & '
           'ELEGIDO=$!; echo "pid del curl elegido: $ELEGIDO"; ')
OTROS = "curl -s -o /dev/null -m 3 https://1.1.1.1; curl -s -o /dev/null 127.0.0.1:8080/; "

with BancoPriv(AQUI, "eBPF con Go — duración y filtro por proceso", proyecto="proyecto-duracion",
               salida="SALIDA-duracion.md") as b:
    b.titulo("Compilar")
    b.sh("go generate ./... 2>&1 | grep -iE 'error|warning' ; go build -o duraciones . && "
         "ls -l duraciones_bpfel.o | awk '{print $5, $9}'",
         mostrar="go generate && go build -o duraciones .")

    b.titulo("Cuánto dura cada conexión")
    b.sh(TRACEFS + SERVIDOR + "(./duraciones -durante 8s > /tmp/out.txt 2>/tmp/err.txt &) ; sleep 2; "
         "curl -s 127.0.0.1:8080/ >/dev/null; curl -s '127.0.0.1:8080/?ms=300' >/dev/null; "
         "curl -s '127.0.0.1:8080/?ms=1200' >/dev/null; curl -s -o /dev/null -m 3 https://1.1.1.1; "
         "curl -s -m 2 127.0.0.1:9; echo \"curl :9 código $?\"; sleep 7; cat /tmp/err.txt /tmp/out.txt",
         mostrar="./duraciones -durante 8s   # y curl a /, /?ms=300, /?ms=1200, a 1.1.1.1 y al puerto 9")

    b.titulo("Filtrar por proceso: el PID que da el kernel, sin traducir")
    b.sh(TRACEFS + SERVIDOR + ELEGIDO + "./duraciones -durante 6s -pid $ELEGIDO -sin-ns 2>&1; wait",
         mostrar="./duraciones -durante 6s -pid $ELEGIDO -sin-ns")
    UNO = ("./duraciones -durante 4s -sin-ns 2>/dev/null & sleep 2; "
           "curl -s -o /dev/null 127.0.0.1:8080/ & echo \"pid del curl según el contenedor: $!\"; wait")
    b.sh(TRACEFS + SERVIDOR + UNO, mostrar="./duraciones -sin-ns   # y un curl, con su pid según el contenedor")

    b.titulo("Lo mismo con --pid=host (Docker Desktop)")
    b.sh(TRACEFS + SERVIDOR + UNO, pid_host=True,
         mostrar="docker run --pid=host ... ./duraciones -sin-ns   # y un curl")

    b.titulo("Filtrar por proceso: con el espacio de PIDs propio")
    b.sh(TRACEFS + SERVIDOR + ELEGIDO + "(./duraciones -durante 6s -pid $ELEGIDO > /tmp/out.txt 2>/tmp/err.txt &) ; "
         "sleep 3; " + OTROS + "sleep 4; cat /tmp/err.txt /tmp/out.txt; wait",
         mostrar="./duraciones -durante 6s -pid $ELEGIDO   # y, mientras, curl a 1.1.1.1 y otro al 8080")
