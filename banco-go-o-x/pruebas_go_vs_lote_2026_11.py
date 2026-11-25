# -*- coding: utf-8 -*-
"""pruebas_go_vs_lote_2026_11.py — la tanda «¿Go o X?» del 2026-11-05 (golandia 803-808: C, Ruby, Dart, Scala, Gleam y TinyGo), con el
MISMO banco que 651-654 y 699 (pruebas_go_vs_x.py + pruebas_go_vs_x_computo.py): mismo servidor de Go y mismo contrato, servidores
con 2 CPU y 1 GiB, `hey -c 16` con 4 CPU desde otro contenedor, la comprobación de que todos devuelven lo mismo ANTES de medir
(si no, se para), latencia a ritmo fijo y cómputo puro con el reloj de cada proceso.

Igual que pruebas_go_vs_deno.py, no se copian las funciones: se importan. Solo cambian los NOMBRES de Docker (contenedor, red,
puerto e imágenes con prefijo v<id>-), porque en esta máquina corren otros bancos a la vez con «srv», «revancha» y el puerto 9300.

⛔ Solo se ejecuta con el candado de medición tomado (`mkdir <proyecto>/_logs/banco.lock`); el guion lo comprueba.

    python investigacion/pruebas_go_vs_lote_2026_11.py c|ruby|dart|scala|gleam|tinygo   → recursos/pruebas-go-vs-<x>/SALIDA.md
"""
import json
import statistics
import sys
import time
import urllib.request
from datetime import datetime
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
import pruebas_go_vs_x as GX  # noqa: E402

R = GX.R
REC = GX.AQUI
HEY_CTX = REC / "pruebas-go-java-revancha"
CANDADO = Path(r"<proyecto>\_logs\banco.lock")
GO_DF = REC / "pruebas-go-vs-comun" / "Dockerfile.go"

X = {
    "c": dict(id=803, nombre="C", carpeta="pruebas-go-vs-c",
              construir=[("c", "gcc -O2 + apt install libmicrohttpd-dev", "Dockerfile.c", "c", "build")],
              variantes=[("c (2 hilos)", "c", {"HILOS": "2"}), ("c (16 hilos)", "c", {"HILOS": "16"})],
              bases={"c": "debian:trixie-slim + libmicrohttpd12 + binario (gcc 16.2 -O2)"},
              binarios=[("c", "/servidor")],
              version=["docker", "run", "--rm", "gcc:16.2.0", "gcc", "--version"]),
    "ruby": dict(id=804, nombre="Ruby", carpeta="pruebas-go-vs-ruby",
                 construir=[("ruby", "bundle install de puma + rack", "Dockerfile.ruby", "ruby", None)],
                 variantes=[("ruby (1 proceso)", "ruby", {"WORKERS": "0"}),
                            ("ruby + YJIT (1 proceso)", "ruby", {"WORKERS": "0", "RUBY_YJIT_ENABLE": "1"}),
                            ("ruby + YJIT (2 procesos)", "ruby", {"WORKERS": "2", "RUBY_YJIT_ENABLE": "1"}),
                            ("ruby + ZJIT (1 proceso)", "ruby", {"WORKERS": "0", "RUBYOPT": "--zjit"})],
                 bases={"ruby": "ruby:4.0.9-slim + Puma 7 + Rack 3 (sin Rails)"},
                 binarios=[],
                 version=["docker", "run", "--rm", "--entrypoint", "ruby", "v804-ruby:1", "--version"]),
    "dart": dict(id=805, nombre="Dart", carpeta="pruebas-go-vs-dart",
                 construir=[("dart compile exe", "dart pub get + dart compile exe (AOT)", "Dockerfile.dart", "dart", "build"),
                            ("dart run", "dart pub get (JIT, sin compilar)", "Dockerfile.dart-jit", "dart-jit", None)],
                 variantes=[("dart AOT (1 isolate)", "dart", {"ISOLATES": "1"}), ("dart AOT (2 isolates)", "dart", {"ISOLATES": "2"}),
                            ("dart run JIT (2 isolates)", "dart-jit", {"ISOLATES": "2"})],
                 bases={"dart": "scratch + /runtime de dart + binario de dart compile exe",
                        "dart-jit": "dart:3.13.7 (SDK completo) + código fuente"},
                 binarios=[("dart", "/app/servidor")],
                 version=["docker", "run", "--rm", "dart:3.13.7", "dart", "--version"]),
    "scala": dict(id=806, nombre="Scala", carpeta="pruebas-go-vs-scala",
                  construir=[("scala", "scala-cli package --assembly (Scala 3.10.0, descarga el compilador)", "Dockerfile.scala",
                              "scala", "build")],
                  variantes=[("scala (hilos virtuales)", "scala", {"EJECUTOR": "virtual"}),
                             ("scala (hilos de plataforma)", "scala", {"EJECUTOR": "plataforma"})],
                  bases={"scala": "eclipse-temurin:25-jre + servidor.jar (con la librería de Scala dentro)"},
                  binarios=[("scala", "/servidor.jar")],
                  version=["docker", "run", "--rm", "--entrypoint", "java", "v806-scala:1", "-version"]),
    # 865: el MISMO Servidor.scala (HTTP a mano sobre ServerSocket, porque Scala Native no trae jdk.httpserver) en la JVM y en
    # Scala Native 0.5.12, con release-full y con el modo por defecto de scala-cli (debug).
    "scala-native": dict(id=865, nombre="Scala Native", carpeta="pruebas-go-vs-scala-native",
                         construir=[("scala JVM", "scala-cli package --assembly (Scala 3.10.0, Java 25)", "Dockerfile.scala-jvm",
                                     "scala-jvm", "build"),
                                    ("scala native", "scala-cli package --native --native-mode release-full",
                                     "Dockerfile.scala-native", "scala-native", "build"),
                                    ("scala native + arreglo", "release-full con SCALANATIVE_GC_TRAP_BASED_YIELDPOINTS=0",
                                     "Dockerfile.scala-native-sintrap", "scala-native-sintrap", "build"),
                                    ("scala native (modo por defecto)", "scala-cli package --native (sin --native-mode)",
                                     "Dockerfile.scala-native-defecto", "scala-native-defecto", "build")],
                         # La carga se mide con el binario release ARREGLADO (sintrap): el release tal cual se cuelga bajo carga
                         # con «Unhandled signal 11» del GC (lo mide aparte extra.py, igual que el modo por defecto, que a ~8
                         # pet/s no da percentiles). «scala native» ve los 24 núcleos del host; «HILOS=2», los 2 de verdad.
                         variantes=[("scala JVM", "scala-jvm", {}), ("scala native", "scala-native-sintrap", {}),
                                    ("scala native, HILOS=2", "scala-native-sintrap", {"HILOS": "2"})],
                         bases={"scala-jvm": "eclipse-temurin:25-jre + servidor.jar",
                                "scala-native": "debian:trixie-slim + binario de Scala Native (release-full)",
                                "scala-native-sintrap": "debian:trixie-slim + binario release-full con el arreglo",
                                "scala-native-defecto": "debian:trixie-slim + binario de Scala Native (modo por defecto)"},
                         binarios=[("scala-jvm", "/servidor.jar"), ("scala-native", "/servidor"), ("scala-native-sintrap", "/servidor"),
                                   ("scala-native-defecto", "/servidor-defecto")],
                         version=["docker", "run", "--rm", "--entrypoint", "bash", "virtuslab/scala-cli:1.18.0", "-c",
                                  "scala-cli version --cli | head -1; echo 'Scala 3.10.0 · Scala Native 0.5.12 (Immix) · "
                                  "JVM: eclipse-temurin:25-jre'"]),
    "gleam": dict(id=807, nombre="Gleam", carpeta="pruebas-go-vs-gleam",
                  construir=[("gleam", "gleam deps download + gleam export erlang-shipment", "Dockerfile.gleam", "gleam", "build")],
                  variantes=[("gleam (mist, BEAM)", "gleam", {})],
                  bases={"gleam": "erlang:29-alpine + erlang-shipment (mist 6)"},
                  binarios=[],
                  version=["docker", "run", "--rm", "--entrypoint", "gleam", "ghcr.io/gleam-lang/gleam:v1.19.2-erlang-alpine",
                           "--version"]),
    # TinyGo compila EL MISMO main.go del servidor de Go (contexto = GX.GO_SRC): no hay otro código que escribir.
    "tinygo": dict(id=808, nombre="TinyGo", carpeta="pruebas-tinygo", ctx="go",
                   construir=[("tinygo", "tinygo build -no-debug (por defecto, -opt=z)", "Dockerfile.tinygo-srv", "tinygo", "build"),
                              ("tinygo -opt=2", "tinygo build -no-debug -opt=2", "Dockerfile.tinygo-srv-opt2", "tinygo-opt2",
                               "build")],
                   variantes=[("tinygo (-opt=z)", "tinygo", {}), ("tinygo -opt=2", "tinygo-opt2", {})],
                   bases={"tinygo": "scratch + binario de tinygo (-opt=z)", "tinygo-opt2": "scratch + binario de tinygo -opt=2"},
                   binarios=[("tinygo", "/servidor"), ("tinygo-opt2", "/servidor")],
                   version=["docker", "run", "--rm", "tinygo/tinygo:0.44.0", "tinygo", "version"]),
}

_sh, _urlopen = R.sh, urllib.request.urlopen
PUERTO = 0
MAPA = {}


def sh(args, **kw):
    return _sh([MAPA.get(a, a.replace("http://srv:8080", f"http://{MAPA['srv']}:8080")) for a in args], **kw)


def urlopen(url, *a, **kw):
    return _urlopen(url.replace("127.0.0.1:9300", f"127.0.0.1:{PUERTO}"), *a, **kw)


def preparar(cfg):
    global PUERTO
    p = f"v{cfg['id']}"
    PUERTO = 9000 + cfg["id"]
    MAPA.update({"srv": f"{p}-srv", "golandia-revancha-hey:1": f"{p}-hey:1", "9300:8080": f"{PUERTO}:8080"})
    R.RED = f"{p}-red"
    R.sh = sh
    urllib.request.urlopen = urlopen


def img(cfg, k):
    return f"v{cfg['id']}-{k}:1"


def construir(cfg, base):
    GO_DF.write_text(GX.GO_DF, encoding="utf-8")
    r = R.sh(["docker", "build", "-q", "-f", str(HEY_CTX / "Dockerfile.hey"), "-t", MAPA["golandia-revancha-hey:1"], str(HEY_CTX)])
    assert r.returncode == 0, r.stderr[-2000:]
    filas = []
    todo = [("go", "go build", GO_DF, GX.GO_SRC, img(cfg, "go"), "build")]
    ctx = GX.GO_SRC if cfg.get("ctx") == "go" else base
    todo += [(n, q, base / df, ctx, img(cfg, k), et) for n, q, df, k, et in cfg["construir"]]
    for nombre, que, df, ctx, tag, etapa in todo:
        args = ["docker", "build", "-q", "--no-cache", "-f", str(df), "-t", tag]
        if etapa:
            args = ["docker", "build", "-q", "--no-cache", "--target", etapa, "-f", str(df), "-t", tag + "-etapa"]
        t0 = time.perf_counter()
        r = R.sh(args + [str(ctx)])
        seg = time.perf_counter() - t0
        assert r.returncode == 0, r.stderr[-2000:]
        if etapa:
            r = R.sh(["docker", "build", "-q", "-f", str(df), "-t", tag, str(ctx)])
            assert r.returncode == 0, r.stderr[-2000:]
        filas.append((nombre, que, seg))
    return filas


def tam_mib(tag):
    return int(R.sh(["docker", "image", "inspect", "-f", "{{.Size}}", tag]).stdout.strip()) / 1048576


def binario_mib(etapa_tag, ruta):
    r = R.sh(["docker", "run", "--rm", "--entrypoint", "stat", etapa_tag, "-c", "%s", ruta])
    return int(r.stdout.strip()) / 1048576


def computo():
    """El método de pruebas_go_vs_x_computo.py: 50 de calentamiento, 300 medidas de una en una, campo «ns» de cada servidor."""
    cal, n = 50, 300

    def medir(ruta):
        ns = []
        for i in range(cal + n):
            j = json.loads(urllib.request.urlopen(f"http://127.0.0.1:9300{ruta}", timeout=60).read())
            if i >= cal:
                ns.append(j["ns"])
        return statistics.median(ns) / 1e6, sorted(ns)[int(n * 0.99) - 1] / 1e6

    R.w("## 8 · Cómputo puro: el reloj del propio proceso, sin red\n")
    R.w(f"{cal} peticiones de calentamiento y {n} medidas, de una en una; el tiempo es el campo «ns» que mide cada servidor "
        "alrededor de su propio cálculo.\n")
    R.w(f"| variante | /suma?n={R.N_SUMA} mediana | p99 | /trabajo?n={R.N_TRABAJO} mediana | p99 |")
    R.w("|---|---|---|---|---|")
    for nombre, imagen, env in R.VARIANTES:
        t0 = R.arrancar(nombre, imagen, env, puerto=9300)
        R.esperar_200(9300, t0)
        s50, s99 = medir(f"/suma?n={R.N_SUMA}")
        t50, t99 = medir(f"/trabajo?n={R.N_TRABAJO}")
        R.sh(["docker", "rm", "-f", "srv"])
        R.w(f"| {nombre} | {s50:.3f} ms | {s99:.3f} ms | {t50:.3f} ms | {t99:.3f} ms |")
    R.w()


def main():
    x = sys.argv[1]
    cfg = X[x]
    base = REC / cfg["carpeta"]
    if "--humo" in sys.argv:  # sin medir: solo comprueba que las imágenes ya construidas responden lo mismo, con 1 CPU
        preparar(cfg)
        R.CPUS = "1"
        R.sh(["docker", "network", "create", R.RED])
        R.sh(["docker", "build", "-q", "-f", str(HEY_CTX / "Dockerfile.hey"), "-t", MAPA["golandia-revancha-hey:1"], str(HEY_CTX)])
        GO_DF.write_text(GX.GO_DF, encoding="utf-8")
        R.sh(["docker", "build", "-q", "-f", str(GO_DF), "-t", img(cfg, "go"), str(GX.GO_SRC)])
        R.VARIANTES[:] = [("go", img(cfg, "go"), {})] + [(n, img(cfg, k), env) for n, k, env in cfg["variantes"]]
        R.mismo_resultado()
        R.sh(["docker", "rm", "-f", "srv"])
        R.sh(["docker", "network", "rm", R.RED])
        return
    if not CANDADO.is_dir():
        raise SystemExit(f"⛔ sin candado: toma {CANDADO} antes de medir (mkdir)")
    preparar(cfg)
    R.OUT.clear()
    R.sh(["docker", "network", "create", R.RED])
    R.w(f"# SALIDA — ¿Go o {cfg['nombre']}? (golandia {cfg['id']}), con el banco de la revancha\n")
    R.w(f"> Corrida del {datetime.now():%Y-%m-%d %H:%M}, con el candado `_logs/banco.lock` tomado. Regenerable con "
        f"`python investigacion/pruebas_go_vs_lote_2026_11.py {x}` (mismo banco que 651-654 y 699). ⛔ Ninguna cifra de blog: "
        "todo salió de aquí.\n")
    filas = construir(cfg, base)
    ver_go = R.sh(["docker", "run", "--rm", "golang:1.27.2", "go", "version"]).stdout.strip()
    v = R.sh(cfg["version"])
    ver_x = " · ".join(l.strip() for l in (v.stdout + v.stderr).strip().splitlines()[:3])
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
    R.w(f"| go | scratch + binario estático | {tam_mib(img(cfg, 'go')):.1f} MiB |")
    for k, b in cfg["bases"].items():
        R.w(f"| {k} | {b} | {tam_mib(img(cfg, k)):.1f} MiB |")
    bins = [f"Go, {binario_mib(img(cfg, 'go') + '-etapa', '/servidor'):.1f} MiB"]
    for k, ruta in cfg["binarios"]:
        bins.append(f"{k}, {binario_mib(img(cfg, k) + '-etapa', ruta):.2f} MiB")
    R.w(f"\nEl artefacto que se copia a la imagen: {'; '.join(bins)}.\n")
    R.VARIANTES[:] = [("go", img(cfg, "go"), {})] + [(n, img(cfg, k), env) for n, k, env in cfg["variantes"]]
    R.mismo_resultado()
    GX.arranque_y_memoria()
    GX.carga_max("trabajo", R.N_TRABAJO, "4 · Carga con memoria: /trabajo (objetos, mapa, orden, JSON), a tope")
    GX.carga_max("suma", R.N_SUMA, "5 · Carga de CPU pura: /suma, a tope")
    GX.ritmo_fijo("trabajo", R.N_TRABAJO, 0, "6 · Latencia de una petición sola: /trabajo")
    GX.ritmo_fijo("suma", R.N_SUMA, 0, "7 · Latencia de una petición sola: /suma")
    computo()
    (base / "SALIDA.md").write_text("\n".join(R.OUT) + "\n", encoding="utf-8")
    R.sh(["docker", "rm", "-f", "srv"])
    R.sh(["docker", "network", "rm", R.RED])
    print(f"\n[ok] {base / 'SALIDA.md'}")


if __name__ == "__main__":
    main()
