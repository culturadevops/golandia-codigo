# -*- coding: utf-8 -*-
"""demos.py — lo de «¿Go o Kotlin Native?» (golandia 879) que no es cronómetro → SALIDA-demos.md.

1. ¿Cuántos núcleos ve cada uno dentro de un contenedor con --cpus 2? (la línea LISTO_EN_MS que imprime cada servidor)
2. ldd de cada binario: ¿de qué depende?
3. La versión de GraalVM y su recolector.
4. Las tareas de Gradle que compilan a nativo (debug y release).

Sin candado: no mide tiempos. Usa las imágenes gl-879-* que deja el banco.

    python recursos/pruebas-go-vs-kotlin-native/demos.py
"""
import subprocess
import time
from datetime import datetime
from pathlib import Path

AQUI = Path(__file__).resolve().parent
OUT = []


def w(s=""):
    print(s, flush=True)
    OUT.append(s)


def sh(args):
    r = subprocess.run(args, capture_output=True, text=True, encoding="utf-8", errors="replace", timeout=900)
    return (r.stdout + r.stderr).rstrip()


def bloque(lineas):
    w("```")
    for l in lineas:
        w(l)
    w("```\n")


w("# SALIDA-demos — ¿Go o Kotlin Native? (golandia 879)\n")
w(f"> Corrida del {datetime.now():%Y-%m-%d %H:%M}. Sin cronómetro de carga.\n")

w("## 1 · ¿Cuántos núcleos ve cada uno? (contenedor con --cpus 2)\n")
lin = []
for img in ("gl-879-kotlin-jvm:1", "gl-879-kotlin-graal:1", "gl-879-kotlin-native:1"):
    sh(["docker", "rm", "-f", "gl-879-demo"])
    sh(["docker", "run", "-d", "--name", "gl-879-demo", "--cpus", "2", "--memory", "1g", "--ulimit", "core=1", img])
    time.sleep(4)
    log = [l for l in sh(["docker", "logs", "gl-879-demo"]).splitlines() if "LISTO_EN_MS" in l]
    lin += [f"$ docker run --cpus 2 {img}", *log]
    sh(["docker", "rm", "-f", "gl-879-demo"])
lin += ["$ nproc   # en un contenedor de 2 CPU",
        sh(["docker", "run", "--rm", "--cpus", "2", "--entrypoint", "nproc", "gl-879-kotlin-native:1"])]
bloque(lin)

w("## 2 · ¿De qué depende cada binario? (ldd)\n")
for img, ruta in (("gl-879-kotlin-native:1", "/servidor"), ("gl-879-kotlin-graal:1", "/servidor")):
    bloque([f"$ ldd {ruta}   # {img}", sh(["docker", "run", "--rm", "--entrypoint", "ldd", img, ruta])])

w("## 3 · GraalVM: versión y recolector\n")
bloque(["$ native-image --version",
        sh(["docker", "run", "--rm", "--entrypoint", "native-image", "ghcr.io/graalvm/native-image-community:25", "--version"])])

w("## 4 · Las tareas de Gradle que enlazan el binario nativo\n")
bloque(["$ gradle tasks --all | grep -i 'link.*LinuxX64'",
        sh(["docker", "run", "--rm", "--entrypoint", "bash", "gl-879-kotlin-native:1-etapa", "-c",
            "cd /src && gradle --no-daemon -q tasks --all 2>/dev/null | grep -i -E '^(link|run).*LinuxX64'"])])

(AQUI / "SALIDA-demos.md").write_text("\n".join(OUT) + "\n", encoding="utf-8")
