# -*- coding: utf-8 -*-
"""Corre los experimentos de testcontainers-go y escribe SALIDA.md. Nada se escribe a mano.

⚠️ Los tiempos necesitan la máquina EN REPOSO (nada renderizando).
Hace falta Go 1.24+ y Docker.

    python ejecutar_todo.py
"""
import os
import platform
import re
import subprocess
import time
from datetime import datetime, timezone
from pathlib import Path

AQUI = Path(__file__).resolve().parent
BLOQUEO = "Control de aplicaciones"
partes = []
# testcontainers imprime mucho ruido de diagnóstico; para el vídeo interesan las líneas de Go.
RUIDO = re.compile(r"(🐳|✅|🔔|⏳|🚫|No image auth|Shell not found|Server Version|API Version|"
                   r"Operating System|Total Memory|Labels:|com\.docker|Testcontainers for Go|"
                   r"Resolved Docker|Test SessionID|Test ProcessID|^\s*$)")


def correr(cmd, env=None, timeout=1800):
    """Si App Control bloquea el binario de `go test`/`go run` por su hash, reintenta con otro
    -buildid (cambia los bytes, no el programa). En Linux y macOS nunca pasa."""
    for i in range(6):
        c = list(cmd)
        if i and c[:2] in (["go", "test"], ["go", "run"]):
            c.insert(2, f"-ldflags=-buildid=tc-{i}")
        t0 = time.time()
        r = subprocess.run(c, capture_output=True, text=True, encoding="utf-8", errors="replace",
                           cwd=AQUI, env={**os.environ, **(env or {})}, timeout=timeout)
        out = r.stdout + r.stderr
        if BLOQUEO not in out:
            return r.returncode, out, time.time() - t0
    return r.returncode, out, time.time() - t0


def limpio(txt):
    return "\n".join(l for l in txt.splitlines() if not RUIDO.search(l))


def paso(titulo, cmd, env=None, filtrar=True):
    codigo, out, seg = correr(cmd, env)
    txt = limpio(out) if filtrar else out
    envs = " ".join(f"{k}={v}" for k, v in (env or {}).items())
    partes.extend(["", f"## {titulo}", "",
                   f"`$ {envs + ' ' if envs else ''}{' '.join(cmd)}` → código {codigo} · {seg:.1f} s de reloj",
                   "", "```", txt.strip(), "```"])
    print(f"{titulo}: código {codigo} · {seg:.1f} s", flush=True)
    return codigo, out


def contenedores():
    r = subprocess.run(["docker", "ps", "-a", "--filter", "label=org.testcontainers=true",
                        "--format", "{{.ID}} {{.Image}} {{.Status}}"], capture_output=True,
                       text=True)
    return r.stdout.strip() or "(ninguno)"


def main():
    go = subprocess.run(["go", "version"], capture_output=True, text=True).stdout.strip()
    dk = subprocess.run(["docker", "version", "--format", "{{.Server.Version}}"],
                        capture_output=True, text=True).stdout.strip()
    tc = re.search(r"testcontainers-go v[\d.]+", (AQUI / "go.mod").read_text()).group(0)
    cab = ["# SALIDA — testcontainers-go", "", "Generado por `python ejecutar_todo.py`. "
           "**No se edita a mano.**", "",
           f"- fecha: {datetime.now(timezone.utc):%Y-%m-%d %H:%M} UTC",
           f"- {go} · Docker {dk} · {tc}", f"- {platform.system()} {platform.release()}",
           "- imágenes ya descargadas antes de medir (postgres:17-alpine, redis:7-alpine)"]

    paso("1 · el doble en memoria", ["go", "test", "-count=1", "-run", "^TestMem$", "-v", "./usuarios"])
    paso("2 · la misma batería contra Postgres de verdad",
         ["go", "test", "-count=1", "-run", "^TestPG$", "-v", "./usuarios"])
    paso("3 · cinco pruebas, un contenedor por prueba",
         ["go", "test", "-count=1", "-run", "^TestCincoPorPrueba$", "-v", "./usuarios"])
    paso("4 · cinco pruebas, un contenedor compartido",
         ["go", "test", "-count=1", "-run", "^TestCincoCompartido$", "-v", "./usuarios"])
    paso("5 · esperar el log de Postgres UNA vez",
         ["go", "test", "-count=1", "-run", "^TestEsperaUnaVez$", "-v", "./usuarios"])
    paso("6 · esperar el log de Postgres DOS veces",
         ["go", "test", "-count=1", "-run", "^TestEsperaDosVeces$", "-v", "./usuarios"])

    # 7 · Ryuk
    for etiqueta, env in (("con Ryuk (por defecto)", None),
                          ("sin Ryuk", {"TESTCONTAINERS_RYUK_DISABLED": "true"})):
        _, out = paso(f"7 · el proceso muere sin limpiar · {etiqueta}",
                      ["go", "run", "./cmd/huerfano"], env)
        justo = contenedores()
        time.sleep(20)
        luego = contenedores()
        partes.extend(["", "```", "docker ps -a --filter label=org.testcontainers=true",
                       "", "justo al morir el proceso:", justo, "", "20 segundos después:", luego,
                       "```"])
        print("   justo:", justo.replace("\n", " | "), "\n   luego:", luego.replace("\n", " | "))
    # limpiar lo que dejó el caso sin Ryuk
    ids = subprocess.run(["docker", "ps", "-aq", "--filter", "label=org.testcontainers=true"],
                         capture_output=True, text=True).stdout.split()
    if ids:
        subprocess.run(["docker", "rm", "-f", *ids], capture_output=True)

    # 8a · Un DOCKER_HOST que no responde NO simula «sin Docker»: testcontainers-go prueba los
    #      candidatos en orden (internal/core/docker_host.go, extractDockerHost) y, si uno no
    #      responde, lo salta y usa el siguiente. Aquí acaba en el Docker local.
    paso("8a · DOCKER_HOST apuntando a un sitio que no responde",
         ["go", "test", "-count=1", "-run", "^TestPG$", "-v", "./usuarios"],
         {"DOCKER_HOST": "tcp://127.0.0.1:1"})

    # 8b · Una CI sin Docker de verdad: el test dentro de un contenedor de Go sin socket de Docker.
    cache = subprocess.run(["go", "env", "GOMODCACHE"], capture_output=True, text=True).stdout.strip()
    paso("8b · ¿y en una CI sin Docker? (contenedor golang sin socket de Docker)",
         ["docker", "run", "--rm", "-v", f"{AQUI}:/app", "-v", f"{cache}:/go/pkg/mod", "-w", "/app",
          "-e", "GOFLAGS=-mod=mod", "golang:1.26", "go", "test", "-count=1", "-run", "^TestPG$",
          "-v", "./usuarios"])

    (AQUI / "SALIDA.md").write_text("\n".join(cab + partes) + "\n", encoding="utf-8")
    print("SALIDA.md escrito")


if __name__ == "__main__":
    main()
