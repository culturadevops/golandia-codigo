# -*- coding: utf-8 -*-
"""ejecutar_demos.py — tipos y permisos de Deno frente a Go (vídeo 699). Todo dentro de contenedores con --rm.

Usa la imagen v699-deno:1 que construye investigacion/pruebas_go_vs_deno.py (denoland/deno:2.9.8 + el servidor) y golang:1.27.2.
    python canales/golandia/recursos/pruebas-go-vs-deno/ejecutar_demos.py   → SALIDA-tipos-permisos.md
"""
import re
import subprocess
from datetime import datetime
from pathlib import Path

AQUI = Path(__file__).resolve().parent
DEMOS = str(AQUI / "demos")
DENO, GO = "v699-deno:1", "golang:1.27.2"
ANSI = re.compile(r"\x1b\[[0-9;]*m")
OUT = []


def correr(titulo, imagen, cmd, nota=""):
    args = ["docker", "run", "--rm", "--cpus", "2", "--memory", "1g", "-e", "NO_COLOR=1", "-v", f"{DEMOS}:/d:ro", "--entrypoint", "sh",
            imagen, "-c", cmd]
    r = subprocess.run(args, capture_output=True, text=True, encoding="utf-8", errors="replace", timeout=300)
    salida = ANSI.sub("", (r.stdout + r.stderr)).strip()
    OUT.append(f"### {titulo}\n")
    if nota:
        OUT.append(nota + "\n")
    OUT.append(f"```\n$ {cmd}\n{salida}\n[código de salida: {r.returncode}]\n```\n")
    print(titulo, "→", r.returncode)


def main():
    OUT.append("# SALIDA — tipos y permisos: Deno 2.9.8 frente a Go 1.27.2 (golandia 699)\n")
    OUT.append(f"> Corrida del {datetime.now():%Y-%m-%d %H:%M}. Regenerable con `python ejecutar_demos.py`. Cada bloque es la salida "
               "literal del contenedor (`docker run --rm`).\n")
    OUT.append("## 1 · Permisos\n")
    correr("Deno sin permisos: el servidor no arranca", DENO, "cd /app && timeout 10 deno run servidor.ts")
    correr("Deno con permiso de red solo para su puerto", DENO,
           "deno run --allow-net=0.0.0.0:8080 /d/permisos.ts")
    correr("El binario de deno compile se lleva los permisos con los que lo compilaste", DENO,
           "deno compile --quiet --allow-net=0.0.0.0:8080 -o /tmp/permisos /d/permisos.ts && /tmp/permisos")
    correr("Deno con -A (todos los permisos): lo mismo que Go", DENO, "deno run -A /d/permisos.ts 2>&1 | grep -v evil")
    correr("Go: el binario puede todo lo que pueda su usuario", GO, "cd /tmp && go run /d/go_permisos/main.go")
    OUT.append("## 2 · Tipos\n")
    correr("deno run NO revisa tipos: el error se ejecuta", DENO, "deno run /d/error_tipo.ts")
    correr("deno check sí lo para", DENO, "deno check /d/error_tipo.ts")
    correr("deno compile también revisa tipos antes de compilar", DENO, "deno compile -o /tmp/x /d/error_tipo.ts")
    correr("Go: el mismo error no compila", GO, "cd /tmp && go run /d/go_tipo/main.go")
    OUT.append("## 3 · Anexo: los hilos de deno serve --parallel y el tamaño en disco\n")
    correr("--parallel con el contenedor limitado a 2 CPU", DENO,
           "cd /app && timeout 3 deno serve --parallel --port 8080 --allow-net paralelo.ts; true")
    lim = subprocess.run(["docker", "run", "--rm", "-e", "NO_COLOR=1", "--entrypoint", "sh", DENO, "-c",
                          "cd /app && timeout 3 deno serve --parallel --port 8080 --allow-net paralelo.ts; true"],
                         capture_output=True, text=True, encoding="utf-8", errors="replace")
    OUT.append(f"Sin `--cpus` (la máquina tiene 24 hilos), la misma orden dice:\n\n```\n{ANSI.sub('', lim.stdout + lim.stderr).strip()}\n```\n")
    correr("El límite de memoria de V8 con deno run (contenedor de 1 GiB)", DENO,
           "deno run /d/heap.ts", nota="Para buscar por qué el binario de `deno compile` usó más memoria bajo carga.")
    correr("El mismo límite en el binario de deno compile", DENO,
           "deno compile --quiet -o /tmp/heap /d/heap.ts && /tmp/heap")
    t = subprocess.run(["docker", "images", "--format", "{{.Repository}}:{{.Tag}}  {{.Size}}"], capture_output=True, text=True).stdout
    filas = [l for l in t.splitlines() if l.startswith(("v699-go:1 ", "v699-deno:1 ", "v699-deno-compile:1 "))]
    OUT.append("Tamaño de las imágenes en disco, descomprimidas (`docker images`); la tabla 2 de SALIDA.md es lo que se descarga, "
               "comprimido (`docker image inspect .Size`):\n\n```\n" + "\n".join(filas) + "\n```\n")
    (AQUI / "SALIDA-tipos-permisos.md").write_text("\n".join(OUT) + "\n", encoding="utf-8")
    print("[ok]", AQUI / "SALIDA-tipos-permisos.md")


if __name__ == "__main__":
    main()
