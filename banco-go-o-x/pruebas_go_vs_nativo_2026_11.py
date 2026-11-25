# -*- coding: utf-8 -*-
"""pruebas_go_vs_nativo_2026_11.py — la tanda «¿Go o X compilado a nativo?» del 2026-11-18 (golandia 879 Kotlin Native y 880
TypeScript con Bun AOT y scriptc), con el MISMO banco que 865 (pruebas_go_vs_lote_2026_11.py): mismo servidor de Go, mismo
contrato, servidores con 2 CPU y 1 GiB, `hey -c 16` con 4 CPU desde otro contenedor, la comprobación de que todos devuelven lo
mismo ANTES de medir (si no, se para), latencia de una petición sola y cómputo puro con el reloj de cada proceso.

No se copian las funciones del banco: se importan de pruebas_go_vs_lote_2026_11 (y este de pruebas_go_vs_x y de la revancha).
Lo que cambia:
  · todos los nombres de Docker llevan el prefijo gl-<id>- (contenedor, red, imágenes, puerto 9000+id), porque en esta máquina
    corren a la vez otros bancos («srv», «revancha», v8xx-, jk-a-, jk-b-);
  · la red se crea con subred fija (Docker se queda sin rangos por las redes de otros proyectos);
  · antes de medir se espera a que la máquina esté en reposo (carga de /proc/loadavg leída desde un contenedor alpine) y se
    apunta en la SALIDA qué contenedores ajenos había y cuánta CPU usaban;
  · la salida va a SALIDA-corrida-<n>.md: se corre dos veces y se comparan.

⛔ Solo se ejecuta con el candado de medición tomado (`mkdir <proyecto>/_logs/banco.lock`); el guion lo comprueba.

    python investigacion/pruebas_go_vs_nativo_2026_11.py kotlin-native 1   → recursos/pruebas-go-vs-kotlin-native/SALIDA-corrida-1.md
"""
import re
import sys
import time
from datetime import datetime
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
import pruebas_go_vs_lote_2026_11 as L  # noqa: E402

R, GX = L.R, L.GX
REPOSO_CARGA = 3.0      # carga de 1 minuto (24 hilos en la VM de Docker) por debajo de la cual se mide
REPOSO_MAX_MIN = 20     # si no baja en este tiempo, se mide igual y queda escrito

L.X["kotlin-native"] = dict(
    id=879, nombre="Kotlin Native", carpeta="pruebas-go-vs-kotlin-native", subred="10.87.9.0/24",
    construir=[("kotlin JVM", "gradle jarCompleto (Kotlin 2.4.21, Ktor 3.6.1 CIO)", "Dockerfile.kotlin-jvm", "kotlin-jvm", "build"),
               ("kotlin GraalVM", "gradle jarCompleto + native-image (GraalVM CE 25)", "Dockerfile.kotlin-graal", "kotlin-graal",
                "build"),
               ("kotlin/native release", "gradle linkReleaseExecutableLinuxX64", "Dockerfile.kotlin-native", "kotlin-native", "build"),
               ("kotlin/native debug", "gradle linkDebugExecutableLinuxX64", "Dockerfile.kotlin-native-debug", "kotlin-native-debug",
                "build")],
    variantes=[("kotlin JVM", "kotlin-jvm", {}), ("kotlin GraalVM", "kotlin-graal", {}), ("kotlin/native", "kotlin-native", {}),
               ("kotlin/native, HILOS=2", "kotlin-native", {"HILOS": "2"})],
    bases={"kotlin-jvm": "eclipse-temurin:25-jre + servidor.jar (Ktor y Kotlin dentro)",
           "kotlin-graal": "debian:trixie-slim + binario de GraalVM native-image",
           "kotlin-native": "debian:trixie-slim + binario de Kotlin/Native (release)",
           "kotlin-native-debug": "debian:trixie-slim + binario de Kotlin/Native (debug)"},
    binarios=[("kotlin-jvm", "/servidor.jar"), ("kotlin-graal", "/servidor-graal"), ("kotlin-native", "/servidor"),
              ("kotlin-native-debug", "/servidor-debug")],
    version=["docker", "run", "--rm", "--entrypoint", "bash", "ghcr.io/graalvm/native-image-community:25", "-c",
             "echo 'Kotlin 2.4.21 · Ktor 3.6.1 (CIO) · Gradle 9.9.1 (gradle:9.9.1-jdk21) · JVM: eclipse-temurin:25-jre'; "
             "native-image --version | head -1"])


# 880: el MISMO src/servidor.ts (node:http) con bun --compile (estable 1.4.4), bun --compile --aot (canary: la estable no lo
# trae) y scriptc 0.4.14. Los tres son un solo hilo de JavaScript/TypeScript: no se usa node:cluster (ver README).
L.X["ts-nativo"] = dict(
    id=880, nombre="TypeScript compilado a nativo", carpeta="pruebas-go-vs-ts-nativo", subred="10.88.0.0/24",
    construir=[("bun --compile", "bun build --compile (Bun 1.4.4)", "Dockerfile.bun", "bun", "build"),
               ("bun --compile --aot", "bun build --compile --aot (Bun 1.4.5-canary.1)", "Dockerfile.bun-aot", "bun-aot", "build"),
               ("scriptc", "apt install clang + descarga de scriptc 0.4.14 + scriptc build", "Dockerfile.scriptc", "scriptc",
                "build")],
    variantes=[("bun --compile", "bun", {}), ("bun --compile --aot", "bun-aot", {}), ("scriptc", "scriptc", {})],
    bases={"bun": "debian:trixie-slim + ejecutable de bun --compile",
           "bun-aot": "debian:trixie-slim + ejecutable de bun --compile --aot",
           "scriptc": "debian:trixie-slim + binario de scriptc"},
    binarios=[("bun", "/servidor"), ("bun-aot", "/servidor"), ("scriptc", "/servidor")],
    version=[["docker", "run", "--rm", "--entrypoint", "bun", "gl-880-bun:1-etapa", "--revision"],
             ["docker", "run", "--rm", "--entrypoint", "bun", "gl-880-bun-aot:1-etapa", "--revision"],
             ["docker", "run", "--rm", "--entrypoint", "scriptc", "gl-880-scriptc:1-etapa", "--version"]])


def preparar(cfg):
    p = f"gl-{cfg['id']}"
    L.PUERTO = 9000 + cfg["id"]
    L.MAPA.update({"srv": f"{p}-srv", "golandia-revancha-hey:1": f"{p}-hey:1", "9300:8080": f"{L.PUERTO}:8080"})
    R.RED = f"{p}-red"
    R.sh = L.sh
    L.urllib.request.urlopen = L.urlopen


def img(cfg, k):
    return f"gl-{cfg['id']}-{k}:1"


L.preparar, L.img = preparar, img


def carga_1min():
    r = R.sh(["docker", "run", "--rm", "alpine:3", "cat", "/proc/loadavg"])
    return float(r.stdout.split()[0]), r.stdout.strip()


def ajenos():
    """Contenedores que no son de este banco y la CPU que gastan en este momento."""
    r = R.sh(["docker", "stats", "--no-stream", "--format", "{{.Name}}\t{{.CPUPerc}}"])
    filas = []
    for l in r.stdout.splitlines():
        if "\t" in l and not l.startswith("gl-"):
            n, c = l.split("\t")
            filas.append((n, c))
    return filas


def esperar_reposo(cfg):
    t0 = time.time()
    while True:
        c, linea = carga_1min()
        if c < REPOSO_CARGA or time.time() - t0 > REPOSO_MAX_MIN * 60:
            break
        print(f"[reposo] carga {c} ≥ {REPOSO_CARGA}: espero 30 s", flush=True)
        time.sleep(30)
    R.w("## Estado de la máquina antes de medir\n")
    R.w(f"- /proc/loadavg (desde un contenedor alpine, VM de Docker con 24 hilos): `{linea}`"
        + ("" if c < REPOSO_CARGA else f" ⚠️ no bajó de {REPOSO_CARGA} en {REPOSO_MAX_MIN} min"))
    fil = ajenos()
    if fil:
        R.w("- contenedores ajenos en marcha (CPU en ese momento): " + ", ".join(f"{n} {c}" for n, c in fil))
    else:
        R.w("- contenedores ajenos en marcha: ninguno")
    R.w()


def main():
    x, corrida = sys.argv[1], sys.argv[2]
    cfg = L.X[x]
    base = L.REC / cfg["carpeta"]
    if not L.CANDADO.is_dir():
        raise SystemExit(f"⛔ sin candado: toma {L.CANDADO} antes de medir (mkdir)")
    preparar(cfg)
    R.OUT.clear()
    R.sh(["docker", "rm", "-f", L.MAPA["srv"]])
    R.sh(["docker", "network", "rm", R.RED])
    r = R.sh(["docker", "network", "create", "--subnet", cfg["subred"], R.RED])
    assert r.returncode == 0, r.stderr
    R.w(f"# SALIDA — ¿Go o {cfg['nombre']}? (golandia {cfg['id']}), corrida {corrida}, con el banco de la serie\n")
    R.w(f"> Corrida del {datetime.now():%Y-%m-%d %H:%M}, con el candado `_logs/banco.lock` tomado. Regenerable con "
        f"`python investigacion/pruebas_go_vs_nativo_2026_11.py {x} {corrida}` (mismo banco que 865). ⛔ Ninguna cifra de blog: "
        "todo salió de aquí.\n")
    filas = L.construir(cfg, base)
    ver_go = R.sh(["docker", "run", "--rm", "golang:1.27.2", "go", "version"]).stdout.strip()
    cmds = cfg["version"] if isinstance(cfg["version"][0], list) else [cfg["version"]]
    lineas = []
    for c in cmds:
        v = R.sh(c)
        lineas += [l.strip() for l in (v.stdout + v.stderr).strip().splitlines()[:3]]
    ver_x = " · ".join(lineas)
    R.w(f"- go: `{ver_go}` (net/http, solo librería estándar)")
    R.w(f"- {x}: `{ver_x}`")
    R.w(f"- Docker `{R.sh(['docker', 'version', '--format', '{{.Server.Version}}']).stdout.strip()}`. Servidores con {R.CPUS} CPU y "
        f"{R.MEM}; `hey -c {R.C}` con 4 CPU desde otro contenedor.\n")
    R.w("## 1 · Construir la imagen desde cero (`docker build --no-cache`, dependencias incluidas)\n")
    R.w("| lenguaje | qué se cronometra | segundos |")
    R.w("|---|---|---|")
    for nombre, que, seg in filas:
        R.w(f"| {nombre} | {que} | {seg:.1f} |")
    R.w()
    R.w("## 2 · Lo que despliegas\n")
    R.w("| imagen | base | tamaño |")
    R.w("|---|---|---|")
    R.w(f"| go | scratch + binario estático | {L.tam_mib(img(cfg, 'go')):.1f} MiB |")
    for k, b in cfg["bases"].items():
        R.w(f"| {k} | {b} | {L.tam_mib(img(cfg, k)):.1f} MiB |")
    bins = [f"Go, {L.binario_mib(img(cfg, 'go') + '-etapa', '/servidor'):.1f} MiB"]
    for k, ruta in cfg["binarios"]:
        bins.append(f"{k}, {L.binario_mib(img(cfg, k) + '-etapa', ruta):.2f} MiB")
    R.w(f"\nEl artefacto que se copia a la imagen: {'; '.join(bins)}.\n")
    esperar_reposo(cfg)
    R.VARIANTES[:] = [("go", img(cfg, "go"), {})] + [(n, img(cfg, k), env) for n, k, env in cfg["variantes"]]
    R.mismo_resultado()
    GX.arranque_y_memoria()
    GX.carga_max("trabajo", R.N_TRABAJO, "4 · Carga con memoria: /trabajo (objetos, mapa, orden, JSON), a tope")
    GX.carga_max("suma", R.N_SUMA, "5 · Carga de CPU pura: /suma, a tope")
    GX.ritmo_fijo("trabajo", R.N_TRABAJO, 0, "6 · Latencia de una petición sola: /trabajo")
    GX.ritmo_fijo("suma", R.N_SUMA, 0, "7 · Latencia de una petición sola: /suma")
    L.computo()
    c, linea = carga_1min()
    R.w(f"Al terminar: /proc/loadavg `{linea}`; ajenos: " + (", ".join(f"{n} {c}" for n, c in ajenos()) or "ninguno") + "\n")
    (base / f"SALIDA-corrida-{corrida}.md").write_text("\n".join(R.OUT) + "\n", encoding="utf-8")
    R.sh(["docker", "rm", "-f", "srv"])
    R.sh(["docker", "network", "rm", R.RED])
    print(f"\n[ok] {base / f'SALIDA-corrida-{corrida}.md'}")


if __name__ == "__main__":
    main()
