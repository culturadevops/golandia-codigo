# -*- coding: utf-8 -*-
"""demos.py — las demos SIN cronómetro de «¿Go o Scala Native?» (golandia 865) → SALIDA-demos.md.

No mide velocidad (eso es SALIDA.md, con el candado): enseña qué hace scala-cli en cada modo, de qué depende el binario,
y que el mismo código fuente corre en la JVM y en nativo.

    python demos.py
"""
import subprocess
from datetime import datetime
from pathlib import Path

AQUI = Path(__file__).resolve().parent
IMG = "virtuslab/scala-cli:1.18.0"
OUT = []


def sh(cmd):
    r = subprocess.run(cmd, capture_output=True, text=True, encoding="utf-8", errors="replace")
    return (r.stdout + r.stderr).rstrip(), r.returncode


def bloque(titulo, comando_visible, script):
    salida, codigo = sh(["docker", "run", "--rm", "-v", f"{AQUI / 'src'}:/src:ro", "--entrypoint", "bash", IMG, "-c", script])
    OUT.extend([f"## {titulo}", "", "```", f"$ {comando_visible}", salida, f"[codigo de salida: {codigo}]", "```", ""])


PREP = "cp /src/Servidor.scala /tmp/ && cd /tmp && "
FILTRO = " 2>&1 | grep -E 'Linking \\(|Optimizing|Compiling to native|Linking native|Total' ; "

OUT += ["# SALIDA-demos — ¿Go o Scala Native? (golandia 865)", "",
        f"> Corrida del {datetime.now():%Y-%m-%d %H:%M}. Sin cronómetro de carga: los tiempos «Total» son los que imprime "
        "scala-cli y se dan solo como orientación (la comparación de build es la de SALIDA.md, con el candado).", ""]
bloque("1 · Sin --native-mode: ¿en qué modo compila?",
       "scala-cli --power package --native Servidor.scala -o servidor",
       PREP + "scala-cli --power package --server=false --scala 3.10.0 --native --native-version 0.5.12 Servidor.scala -o /tmp/s"
       + FILTRO + "ls -l /tmp/s | awk '{print $5, $9}'")
bloque("2 · Con --native-mode release-full",
       "scala-cli --power package --native --native-mode release-full Servidor.scala -o servidor",
       PREP + "scala-cli --power package --server=false --scala 3.10.0 --native --native-version 0.5.12 --native-mode release-full "
       "Servidor.scala -o /tmp/s" + FILTRO + "ls -l /tmp/s | awk '{print $5, $9}'; echo '$ ldd servidor'; ldd /tmp/s")
bloque("3 · El mismo archivo, en la JVM",
       "scala-cli --power package --assembly Servidor.scala -o servidor.jar",
       PREP + "scala-cli --power package --server=false --scala 3.10.0 --jvm 25 --assembly --preamble=false Servidor.scala "
       "-o /tmp/s.jar 2>&1 | tail -1; ls -l /tmp/s.jar | awk '{print $5, $9}'")
bloque("4 · Líneas del servidor",
       "grep -cv '^ *//' Servidor.scala",
       "grep -cv '^ *//' /src/Servidor.scala; grep -c '' /src/Servidor.scala")
# 5 · ¿Cuántos núcleos cree tener cada uno? Mismo contenedor de 2 CPU; cada servidor imprime Runtime.availableProcessors.
OUT.extend(["## 5 · ¿Cuántos núcleos ve cada uno? (contenedor con --cpus 2)", "", "```"])
for img in ("v865-scala-jvm:1", "v865-scala-native-sintrap:1"):
    sh(["docker", "rm", "-f", "v865-demo"])
    sh(["docker", "run", "-d", "--name", "v865-demo", "--cpus", "2", "--memory", "1g", img])
    import time
    time.sleep(3)
    log, _ = sh(["docker", "logs", "v865-demo"])
    OUT.extend([f"$ docker run --cpus 2 {img}", log.strip().splitlines()[-1]])
    sh(["docker", "rm", "-f", "v865-demo"])
OUT.extend(["$ nproc   # en el mismo contenedor de 2 CPU", sh(["docker", "run", "--rm", "--cpus", "2", "debian:trixie-slim", "nproc"])[0],
            "```", ""])
(AQUI / "SALIDA-demos.md").write_text("\n".join(OUT) + "\n", encoding="utf-8")
print("\n".join(OUT))
