# -*- coding: utf-8 -*-
"""pruebas_go_java_revancha.py — la revancha de Go contra Java que pidieron en los comentarios del vídeo 148 (8mCKytldeG4).
Guarda la salida LITERAL en `recursos/pruebas-go-java-revancha/SALIDA.md`.

Lo que pidieron o corrigieron (comentarios-2026-10-15.md):
  · «la revancha con GraalVM» (dos comentarios)           → variante java-nativo (GraalVM native-image)
  · «la imagen de java está deprecada, usa JLink»          → variante java-jlink (JRE mínimo sobre debian slim)
  · «¿por qué hilos virtuales para CPU?»                   → misma imagen con EJECUTOR=plataforma (un hilo por núcleo)
  · «son aplicaciones hola mundo»                          → /trabajo: objetos, mapa, orden y JSON (reserva memoria)
Y un cambio de método: la carga la genera `hey` en su propio contenedor, con los servidores limitados a 2 CPU y 1 GiB,
porque en el 148 la prueba de 300 conexiones la limitaba el cliente (Python), no el servidor.

    python investigacion/pruebas_go_java_revancha.py
"""
import io
import json
import re
import statistics
import subprocess
import sys
import threading
import time
import urllib.request
from datetime import datetime
from pathlib import Path

sys.stdout = io.TextIOWrapper(sys.stdout.buffer, encoding="utf-8")
BASE = Path(__file__).resolve().parent.parent / "recursos" / "pruebas-go-java-revancha"
BIN = BASE / "_bin"
RED = "revancha"
CPUS, MEM = "2", "1g"
N_TRABAJO = 20_000
N_SUMA = 200_000
C, SEG_FRIO, SEG_CALIENTE = 16, 10, 20
REPES_FRIO = 8
IMG = {
    "go": "golandia-revancha-go:1",
    "java-jre": "golandia-revancha-java-jre:1",
    "java-jlink": "golandia-revancha-java-jlink:1",
    "java-nativo": "golandia-revancha-java-nativo:1",
}
VARIANTES = [  # (nombre, imagen, entorno)
    ("go", IMG["go"], {}),
    ("java-jre (virtuales)", IMG["java-jre"], {"EJECUTOR": "virtual"}),
    ("java-jre (plataforma)", IMG["java-jre"], {"EJECUTOR": "plataforma"}),
    ("java-jlink", IMG["java-jlink"], {"EJECUTOR": "virtual"}),
    ("java-nativo (GraalVM)", IMG["java-nativo"], {"EJECUTOR": "virtual"}),
]
OUT = []


def sh(args, **kw):
    return subprocess.run(args, capture_output=True, text=True, encoding="utf-8", errors="replace", **kw)


def w(s=""):
    print(s, flush=True)
    OUT.append(s)


def win(p):
    return str(p).replace("/", "\\")


def compilar():
    BIN.mkdir(exist_ok=True)
    pasos = [
        ("go build", "golang:1.26", BASE / "go",
         "cd /src && time -p env CGO_ENABLED=0 go build -ldflags='-s -w' -o /bin_out/servidor . 2>&1"),
        ("javac + jar", "eclipse-temurin:25-jdk", BASE / "java",
         "cd /src && time -p sh -c 'javac -d /tmp/c Servidor.java && jar --create --file /bin_out/servidor.jar --main-class Servidor -C /tmp/c .' 2>&1"),
        ("jlink (JRE mínimo)", "eclipse-temurin:25-jdk", BASE / "java",
         # Solo se cronometra aquí: el JRE de la imagen se genera dentro del docker build (trae enlaces simbólicos que una
         # carpeta de Windows no conserva).
         "rm -rf /bin_out/jre && time -p jlink --add-modules java.base,jdk.httpserver --strip-debug --no-man-pages --no-header-files "
         "--compress zip-6 --output /tmp/jre 2>&1"),
        ("native-image (GraalVM)", "ghcr.io/graalvm/native-image-community:25", BASE / "java",
         "cd /bin_out && time -p native-image --no-fallback -jar servidor.jar -o servidor-nativo > /bin_out/native-image.log 2>&1; tail -3 /bin_out/native-image.log"),
    ]
    w("## 1 · Compilar (dentro de la imagen oficial de cada uno, `time -p`, caché vacía)\n")
    w("| paso | imagen | segundos |")
    w("|---|---|---|")
    for nombre, imagen, src, cmd in pasos:
        r = sh(["docker", "run", "--rm", "--entrypoint", "bash", "-v", f"{win(src)}:/src", "-v", f"{win(BIN)}:/bin_out", imagen, "-c", cmd])
        m = re.search(r"real\s+([\d.]+)", r.stdout + r.stderr)
        if not m:
            raise SystemExit(f"falló {nombre}:\n{r.stdout[-2000:]}\n{r.stderr[-2000:]}")
        w(f"| {nombre} | `{imagen}` | {float(m.group(1)):.1f} |")
    # hey: el generador de carga, compilado igual, en su propia imagen
    if not (BIN / "hey").exists():
        r = sh(["docker", "run", "--rm", "-v", f"{win(BIN)}:/bin_out", "-e", "CGO_ENABLED=0", "-e", "GOBIN=/bin_out", "golang:1.26",
                "go", "install", "github.com/rakyll/hey@v0.1.4"])
        assert (BIN / "hey").exists(), r.stderr[-1500:]
    w()


DOCKERFILES = {
    "go": "FROM scratch\nCOPY _bin/servidor /servidor\nEXPOSE 8080\nENTRYPOINT [\"/servidor\"]\n",
    "java-jre": "FROM eclipse-temurin:25-jre\nCOPY _bin/servidor.jar /app/servidor.jar\nEXPOSE 8080\nENTRYPOINT [\"java\", \"-jar\", \"/app/servidor.jar\"]\n",
    "java-jlink": "FROM eclipse-temurin:25-jdk AS jre\nRUN jlink --add-modules java.base,jdk.httpserver --strip-debug --no-man-pages "
                  "--no-header-files --compress zip-6 --output /jre\n"
                  "FROM debian:trixie-slim\nCOPY --from=jre /jre /opt/jre\nCOPY _bin/servidor.jar /app/servidor.jar\nEXPOSE 8080\n"
                  "ENTRYPOINT [\"/opt/jre/bin/java\", \"-jar\", \"/app/servidor.jar\"]\n",
    "java-nativo": "FROM debian:trixie-slim\nCOPY _bin/servidor-nativo /servidor\nEXPOSE 8080\nENTRYPOINT [\"/servidor\"]\n",
    "hey": "FROM scratch\nCOPY _bin/hey /hey\nENTRYPOINT [\"/hey\"]\n",
}


def imagenes():
    for k, df in DOCKERFILES.items():
        f = BASE / f"Dockerfile.{k}"
        f.write_text(df, encoding="utf-8")
        tag = IMG.get(k, "golandia-revancha-hey:1")
        r = sh(["docker", "build", "-q", "-f", str(f), "-t", tag, str(BASE)])
        assert r.returncode == 0, r.stderr[-1500:]
    w("## 2 · Lo que despliegas\n")
    w("| imagen | base | descarga (comprimida) |")
    w("|---|---|---|")
    bases = {"go": "scratch", "java-jre": "eclipse-temurin:25-jre", "java-jlink": "debian:trixie-slim + JRE de jlink",
             "java-nativo": "debian:trixie-slim + binario nativo"}
    for k, tag in IMG.items():
        b = int(sh(["docker", "image", "inspect", "-f", "{{.Size}}", tag]).stdout.strip())
        w(f"| {k} | {bases[k]} | {b / 1048576:.1f} MiB |")
    jre = int(sh(["docker", "run", "--rm", "--entrypoint", "du", IMG["java-jlink"], "-sb", "/opt/jre"]).stdout.split()[0])
    w(f"\nEl JRE de jlink (java.base + jdk.httpserver) pesa {jre / 1048576:.1f} MiB; el binario de GraalVM, "
      f"{(BIN / 'servidor-nativo').stat().st_size / 1048576:.1f} MiB; el de Go, {(BIN / 'servidor').stat().st_size / 1048576:.1f} MiB.\n")


def arrancar(nombre, imagen, env, puerto=None):
    sh(["docker", "rm", "-f", "srv"])
    # --ulimit core=1: si el servidor cae (p. ej. Scala Native con signal 11), WSL no escribe un volcado de hasta 17 GB en
    # %LOCALAPPDATA%\Temp\wsl-crashes. core=0 NO basta: el core_pattern es una tubería y el kernel solo la corta con 1.
    args = ["docker", "run", "-d", "--name", "srv", "--network", RED, "--cpus", CPUS, "--memory", MEM, "--ulimit", "core=1"]
    for k, v in env.items():
        args += ["-e", f"{k}={v}"]
    if puerto:
        args += ["-p", f"{puerto}:8080"]
    t0 = time.perf_counter()
    r = sh(args + [imagen])
    assert r.returncode == 0, r.stderr
    return t0


def esperar_200(puerto, t0):
    while True:
        try:
            urllib.request.urlopen(f"http://127.0.0.1:{puerto}/suma?n=1", timeout=1).read()
            return (time.perf_counter() - t0) * 1000
        except Exception:
            if time.perf_counter() - t0 > 30:
                raise SystemExit("no arrancó")
            time.sleep(0.005)


def mem_mib():
    s = sh(["docker", "stats", "--no-stream", "--format", "{{.MemUsage}}", "srv"]).stdout.split("/")[0].strip()
    v = float(re.match(r"[\d.]+", s).group(0))
    return v / 1024 if "KiB" in s else v * 1024 if "GiB" in s else v


def hey(url, seg):
    r = sh(["docker", "run", "--rm", "--network", RED, "--cpus", "4", "golandia-revancha-hey:1", "-z", f"{seg}s", "-c", str(C), url])
    t = r.stdout
    rps = float(re.search(r"Requests/sec:\s+([\d.]+)", t).group(1))
    lat = {q: float(re.search(rf"\s{q}% in ([\d.]+) secs", t).group(1)) * 1000 for q in (50, 99)}
    codigos = re.findall(r"\[(\d+)\]\s+(\d+) responses", t)
    assert codigos and all(c == "200" for c, _ in codigos), f"códigos raros: {codigos}"
    return rps, lat


def mismo_resultado():
    w("## 0 · ¿Hacen lo mismo? (antes de medir)\n")
    salidas = {}
    for nombre, imagen, env in VARIANTES:
        t0 = arrancar(nombre, imagen, env, puerto=9300)
        esperar_200(9300, t0)
        j = json.loads(urllib.request.urlopen(f"http://127.0.0.1:9300/trabajo?n={N_TRABAJO}").read())
        s = json.loads(urllib.request.urlopen(f"http://127.0.0.1:9300/suma?n={N_SUMA}").read())
        salidas[nombre] = (j["total"], [x["id"] for x in j["primeros"]], s["suma"])
        sh(["docker", "rm", "-f", "srv"])
    ref = salidas["go"]
    for nombre, v in salidas.items():
        ok = v == ref
        w(f"- {nombre}: total {v[0]}, primeros ids {v[1][:5]}…, suma {v[2]} → {'igual que Go' if ok else '⛔ DISTINTO'}")
        if not ok:
            raise SystemExit("las variantes no devuelven lo mismo: no se compara")
    w()


def arranque_y_memoria():
    w(f"## 3 · Arranque: contenedor nuevo → primera respuesta 200 (mediana de {REPES_FRIO}) y memoria en reposo\n")
    w("| variante | 1ª respuesta (docker incluido) | «listo» (su propio reloj) | memoria en reposo |")
    w("|---|---|---|---|")
    for nombre, imagen, env in VARIANTES:
        ts, listos, mems = [], [], []
        for _ in range(REPES_FRIO):
            t0 = arrancar(nombre, imagen, env, puerto=9300)
            ts.append(esperar_200(9300, t0))
            log = sh(["docker", "logs", "srv"]).stderr + sh(["docker", "logs", "srv"]).stdout
            m = re.search(r"LISTO_EN_MS ([\d.]+)", log)
            if m:
                listos.append(float(m.group(1)))
            time.sleep(1)
            mems.append(mem_mib())
        sh(["docker", "rm", "-f", "srv"])
        w(f"| {nombre} | {statistics.median(ts):.0f} ms | {statistics.median(listos):.1f} ms | {statistics.median(mems):.1f} MiB |")
    w()


def carga(ruta, n, titulo):
    w(f"## {titulo}\n")
    w(f"`hey -c {C}` desde otro contenedor; servidor con {CPUS} CPU y {MEM} de memoria. «En frío»: los primeros {SEG_FRIO} s tras "
      f"arrancar. «Templado»: los {SEG_CALIENTE} s siguientes.\n")
    w("| variante | peticiones/s en frío | peticiones/s templado | p50 | p99 | memoria máx. bajo carga |")
    w("|---|---|---|---|---|---|")
    for nombre, imagen, env in VARIANTES:
        t0 = arrancar(nombre, imagen, env, puerto=9300)
        esperar_200(9300, t0)
        url = f"http://srv:8080/{ruta}?n={n}"
        frio, _ = hey(url, SEG_FRIO)
        pico = [0.0]
        fin = threading.Event()

        def muestrear():
            while not fin.is_set():
                try:
                    pico[0] = max(pico[0], mem_mib())
                except Exception:
                    pass

        th = threading.Thread(target=muestrear)
        th.start()
        rps, lat = hey(url, SEG_CALIENTE)
        fin.set()
        th.join()
        sh(["docker", "rm", "-f", "srv"])
        w(f"| {nombre} | {frio:.0f} | {rps:.0f} | {lat[50]:.1f} ms | {lat[99]:.1f} ms | {pico[0]:.0f} MiB |")
    w()


def main():
    sh(["docker", "network", "create", RED])
    ver = {
        "docker": sh(["docker", "version", "--format", "{{.Server.Version}}"]).stdout.strip(),
        "go": sh(["docker", "run", "--rm", "golang:1.26", "go", "version"]).stdout.strip(),
        "java": sh(["docker", "run", "--rm", "eclipse-temurin:25-jdk", "java", "-version"]).stderr.splitlines()[0],
        "graal": sh(["docker", "run", "--rm", "--entrypoint", "native-image", "ghcr.io/graalvm/native-image-community:25",
                     "--version"]).stdout.splitlines()[0],
    }
    w("# SALIDA — la revancha de Go contra Java (vídeo 148), medida en esta máquina\n")
    w(f"> Corrida del {datetime.now():%Y-%m-%d %H:%M}. Regenerable con `python investigacion/pruebas_go_java_revancha.py`. "
      "⛔ Ninguna cifra de blog: todo salió de aquí.\n")
    for k, v in ver.items():
        w(f"- {k}: `{v}`")
    w(f"- Máquina: 24 hilos (Windows 11 + Docker Desktop). Servidores limitados a {CPUS} CPU y {MEM}; `hey` con 4 CPU.")
    w(f"- /trabajo?n={N_TRABAJO}: crea {N_TRABAJO} objetos, los mete en un mapa, los ordena y arma un JSON a mano (igual en los dos).")
    w(f"- /suma?n={N_SUMA}: el trabajo de CPU del vídeo 148 (multiplicar y módulo).\n")
    compilar()
    imagenes()
    mismo_resultado()
    arranque_y_memoria()
    carga("trabajo", N_TRABAJO, "4 · Carga con memoria: /trabajo (objetos, mapa, orden, JSON)")
    carga("suma", N_SUMA, "5 · Carga de CPU pura: /suma (aquí se ve si los hilos virtuales molestan)")
    (BASE / "SALIDA.md").write_text("\n".join(OUT) + "\n", encoding="utf-8")
    print("\n[ok] SALIDA.md escrita")


if __name__ == "__main__":
    main()
