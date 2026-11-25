# -*- coding: utf-8 -*-
"""ejecutar_todo.py — «¿Go o Rust? Lo medí» (golandia 698, 2026-10-29). Regenera SALIDA.md de esta carpeta.

El MISMO banco que «¿Go o X?» (651-654): reutiliza las funciones de investigacion/pruebas_go_java_revancha.py y
investigacion/pruebas_go_vs_x.py sin copiarlas (mismo servidor de Go, mismo contrato /suma y /trabajo, servidores con
--cpus 2 y 1 GiB, `hey -c 16` con 4 CPU desde otro contenedor, la comprobación de «misma respuesta» antes de medir, el
arranque en mediana de 8, la latencia de una petición sola con `hey -c 1 -q 5`) y el cómputo puro de
investigacion/pruebas_go_vs_x_computo.py (50 de calentamiento + 300 medidas, campo «ns» del propio proceso).

Lo único que cambia son los NOMBRES, porque hay otros agentes con Docker a la vez (encargo 2026-10-29): contenedor
v698-srv, red v698-red, puerto 9698 e imágenes v698-*. Se cambian envolviendo `sh` y `urlopen`, no editando el banco.

Añadido de este vídeo: «lo que cuesta escribirlo» — líneas, dependencias, compilación limpia e incremental (con --cpus 2)
y tamaño del binario.

    python canales/golandia/recursos/pruebas-go-vs-rust/ejecutar_todo.py
"""
import re
import statistics
import subprocess
import sys
import urllib.request
from datetime import datetime
from pathlib import Path

AQUI = Path(__file__).resolve().parent
INV = AQUI.parent.parent / "investigacion"
sys.path.insert(0, str(INV))
import pruebas_go_java_revancha as R  # noqa: E402
import pruebas_go_vs_x as X  # noqa: E402
import pruebas_go_vs_x_computo as K  # noqa: E402

# ── nombres propios (v698) sin tocar el banco ─────────────────────────────────────────────
R.RED = "v698-red"
_sh, _urlopen = R.sh, urllib.request.urlopen
CAMBIOS = {"srv": "v698-srv", "golandia-revancha-hey:1": "v698-hey:1", "9300:8080": "9698:8080"}


def sh(args, **kw):
    a = [CAMBIOS.get(x, x).replace("http://srv:8080", "http://v698-srv:8080") for x in args]
    return _sh(a, **kw)


def urlopen(url, *a, **kw):
    return _urlopen(url.replace("127.0.0.1:9300", "127.0.0.1:9698"), *a, **kw)


R.sh = sh
urllib.request.urlopen = urlopen

GO_IMG = "v698-govs-go:1"
RUST_IMG = "v698-govs-rust:1"
MUSL_IMG = "v698-govs-rust-musl:1"
GO_VER, RUST_VER = "golang:1.27.2", "rust:1.99.1-trixie"
VARIANTES = [("go", GO_IMG, {}), ("rust (glibc)", RUST_IMG, {}), ("rust (musl)", MUSL_IMG, {})]
BASES = {"go": "scratch + binario estático",
         "rust (glibc)": "debian:trixie-slim + binario (Rust 1.99.1, axum 0.8.9, tokio 1.54.1)",
         "rust (musl)": "scratch + binario estático musl (Rust 1.99.1, axum 0.8.9, tokio 1.54.1)"}
w = R.w


def construir():
    """Como X.construir: `docker build --no-cache --target build` cronometrado, luego la imagen final con caché."""
    df_go = AQUI / "Dockerfile.go"
    df_go.write_text(X.GO_DF, encoding="utf-8")
    filas = []
    for nombre, df, ctx, tag in [("go", df_go, X.GO_SRC, GO_IMG),
                                 ("rust (glibc)", AQUI / "Dockerfile.rust", AQUI, RUST_IMG),
                                 ("rust (musl)", AQUI / "Dockerfile.rust-musl", AQUI, MUSL_IMG)]:
        import time
        t0 = time.perf_counter()
        r = R.sh(["docker", "build", "-q", "--no-cache", "--target", "build", "-f", str(df), "-t", tag + "-etapa", str(ctx)])
        seg = time.perf_counter() - t0
        assert r.returncode == 0, r.stderr[-2000:]
        r = R.sh(["docker", "build", "-q", "-f", str(df), "-t", tag, str(ctx)])
        assert r.returncode == 0, r.stderr[-2000:]
        filas.append((nombre, seg))
    r = R.sh(["docker", "build", "-q", "-f", str(R.BASE / "Dockerfile.hey"), "-t", "v698-hey:1", str(R.BASE)])
    assert r.returncode == 0, r.stderr[-1500:]
    return filas


# ── lo que cuesta escribirlo ──────────────────────────────────────────────────────────────
REPES = 3
GO_ESCRIBIR = r"""set -e
cp -r /src /tmp/w && cd /tmp/w
t() { a=$(date +%s.%N); "$@" > /tmp/log 2>&1 || { cat /tmp/log >&2; exit 1; }; b=$(date +%s.%N); echo "$a $b"; }
echo "T limpia $(t env CGO_ENABLED=0 go build -ldflags='-s -w' -o /tmp/s .)"
sed -i 's/"lang":"go","n":%d,"suma":0/"lang":"go!","n":%d,"suma":0/' main.go && grep -q 'go!' main.go
echo "T incremental $(t env CGO_ENABLED=0 go build -ldflags='-s -w' -o /tmp/s .)"
echo "B $(stat -c %s /tmp/s)"
"""
RUST_ESCRIBIR = r"""set -e
mkdir /tmp/w && cp -r /src/Cargo.toml /src/Cargo.lock /src/src /tmp/w/ && cd /tmp/w
t() { a=$(date +%s.%N); "$@" > /tmp/log 2>&1 || { cat /tmp/log >&2; exit 1; }; b=$(date +%s.%N); echo "$a $b"; }
echo "T descarga $(t cargo fetch --locked)"
echo "T limpia $(t cargo build --release --locked --offline)"
sed -i 's/"lang":"rust","n":{ms}/"lang":"rust!","n":{ms}/' src/main.rs && grep -q 'rust!' src/main.rs
echo "T incremental $(t cargo build --release --locked --offline)"
echo "B $(stat -c %s target/release/servidor)"
echo "T limpia-debug $(t cargo build --locked --offline)"
sed -i 's/"lang":"rust!","n":{ms}/"lang":"rust?","n":{ms}/' src/main.rs && grep -q 'rust?' src/main.rs
echo "T incremental-debug $(t cargo build --locked --offline)"
"""


def escribir_una(imagen, src, guion):
    r = R.sh(["docker", "run", "--rm", "--name", "v698-escribir", "--cpus", "2", "-v", f"{R.win(src)}:/src:ro",
              "--entrypoint", "sh", imagen, "-c", guion])
    assert r.returncode == 0, r.stdout[-1500:] + r.stderr[-1500:]
    t, b = {}, None
    for linea in r.stdout.splitlines():
        p = linea.split()
        if p[0] == "T":
            t[p[1]] = float(p[3]) - float(p[2])
        elif p[0] == "B":
            b = int(p[1])
    return t, b


def lineas(f):
    """Líneas de código: sin blancos ni comentarios de línea (el mismo criterio con que salieron las 90 de Go en 651/654)."""
    return sum(1 for l in f.read_text(encoding="utf-8").splitlines() if l.strip() and not l.strip().startswith("//"))


def cuesta_escribirlo():
    w("## 9 · Lo que cuesta escribirlo\n")
    go_src, rs_src = X.GO_SRC / "main.go", AQUI / "src" / "main.rs"
    paquetes = len(re.findall(r"^\[\[package\]\]", (AQUI / "Cargo.lock").read_text(encoding="utf-8"), re.M))
    w("| | go | rust |")
    w("|---|---|---|")
    w(f"| líneas de código (sin blancos ni comentarios) | {lineas(go_src)} | {lineas(rs_src)} |")
    w(f"| líneas totales del archivo | {len(go_src.read_text(encoding='utf-8').splitlines())} | "
      f"{len(rs_src.read_text(encoding='utf-8').splitlines())} |")
    w(f"| dependencias externas | 0 (solo librería estándar) | 2 directas (axum, tokio) · {paquetes - 1} paquetes en Cargo.lock |")
    w()
    w(f"Compilación dentro de `{GO_VER}` y `{RUST_VER}` con `--cpus 2`, contenedor nuevo cada vez (cachés vacías), "
      f"mediana de {REPES}. «Incremental»: cambiar una cadena en el manejador de /espera y volver a compilar.\n")
    res = {"go": [], "rust": []}
    bins = {}
    for _ in range(REPES):
        t, b = escribir_una(GO_VER, X.GO_SRC, GO_ESCRIBIR)
        res["go"].append(t)
        bins["go"] = b
        t, b = escribir_una(RUST_VER, AQUI, RUST_ESCRIBIR)
        res["rust"].append(t)
        bins["rust"] = b

    def med(lang, k):
        return statistics.median(x[k] for x in res[lang])

    def todas(lang, k):
        d = 2 if k.startswith("incremental") else 1
        return " / ".join(f"{x[k]:.{d}f}" for x in res[lang])

    w("| paso | go (s) | rust (s) | las 3 corridas go · rust |")
    w("|---|---|---|---|")
    w(f"| bajar dependencias | — | {med('rust', 'descarga'):.1f} | — · {todas('rust', 'descarga')} |")
    w(f"| compilación limpia (release) | {med('go', 'limpia'):.1f} | {med('rust', 'limpia'):.1f} | "
      f"{todas('go', 'limpia')} · {todas('rust', 'limpia')} |")
    w(f"| incremental (release) | {med('go', 'incremental'):.2f} | {med('rust', 'incremental'):.2f} | "
      f"{todas('go', 'incremental')} · {todas('rust', 'incremental')} |")
    w(f"| compilación limpia (debug) | (Go no tiene modo aparte) | {med('rust', 'limpia-debug'):.1f} | "
      f"— · {todas('rust', 'limpia-debug')} |")
    w(f"| incremental (debug) | — | {med('rust', 'incremental-debug'):.2f} | — · {todas('rust', 'incremental-debug')} |")
    w()
    # el binario musl, tal cual queda en la etapa de compilación de la imagen (la imagen final es scratch: sin `stat`)
    musl = int(R.sh(["docker", "run", "--rm", "--entrypoint", "stat", MUSL_IMG + "-etapa", "-c", "%s", "/servidor"]).stdout.strip())
    w("| binario | bytes | MiB |")
    w("|---|---|---|")
    w(f"| go (`-ldflags='-s -w'`) | {bins['go']} | {bins['go'] / 1048576:.2f} |")
    w(f"| rust glibc (`strip = true`) | {bins['rust']} | {bins['rust'] / 1048576:.2f} |")
    w(f"| rust musl (estático, el de la imagen scratch) | {musl} | {musl / 1048576:.2f} |")
    w()


def computo():
    w(f"## 8 · Cómputo puro (reloj del propio proceso, sin red)\n")
    w(f"Como `investigacion/pruebas_go_vs_x_computo.py`: contenedor con {R.CPUS} CPU y {R.MEM}, {K.CAL} peticiones de "
      f"calentamiento y {K.N} medidas, de una en una; el tiempo es el campo «ns» que mide cada servidor alrededor de su cálculo.\n")
    w(f"| variante | /suma?n={R.N_SUMA} mediana | p99 | /trabajo?n={R.N_TRABAJO} mediana | p99 |")
    w("|---|---|---|---|---|")
    for nombre, img, env in VARIANTES:
        t0 = R.arrancar(nombre, img, env, puerto=9300)
        R.esperar_200(9300, t0)
        s50, s99 = K.medir(f"/suma?n={R.N_SUMA}")
        t50, t99 = K.medir(f"/trabajo?n={R.N_TRABAJO}")
        R.sh(["docker", "rm", "-f", "srv"])
        w(f"| {nombre} | {s50:.3f} ms | {s99:.3f} ms | {t50:.3f} ms | {t99:.3f} ms |")
    w()


def anexo_hilos():
    w("## Anexo · los núcleos que ve tokio dentro del contenedor\n")
    for cpus in ("2", None):
        args = ["docker", "run", "-d", "--name", "srv"] + (["--cpus", cpus] if cpus else []) + [RUST_IMG]
        R.sh(["docker", "rm", "-f", "srv"])
        R.sh(args)
        import time
        time.sleep(1.5)
        log = R.sh(["docker", "logs", "srv"])
        linea = (log.stderr + log.stdout).strip().splitlines()[0]
        R.sh(["docker", "rm", "-f", "srv"])
        w(f"- `docker run {'--cpus ' + cpus + ' ' if cpus else ''}{RUST_IMG}` → `{linea}`")
    w("\n`hilos` es `std::thread::available_parallelism()`, el número con el que `#[tokio::main]` crea sus hilos de trabajo.\n")


def otros_contenedores():
    r = _sh(["docker", "ps", "--format", "{{.Names}}"])
    return [n for n in r.stdout.split() if not n.startswith("v698-")]


def main():
    R.OUT.clear()
    R.sh(["docker", "network", "create", R.RED])
    w("# SALIDA — ¿Go o Rust? (golandia 698), con el banco de la revancha\n")
    w(f"> Corrida del {datetime.now():%Y-%m-%d %H:%M}. Regenerable con `python canales/golandia/recursos/pruebas-go-vs-rust/"
      "ejecutar_todo.py`. ⛔ Ninguna cifra de blog: todo salió de aquí.\n")
    ajenos = otros_contenedores()
    filas = construir()
    ver_go = R.sh(["docker", "run", "--rm", GO_VER, "go", "version"]).stdout.strip()
    ver_rs = R.sh(["docker", "run", "--rm", RUST_VER, "rustc", "--version"]).stdout.strip()
    w(f"- go: `{ver_go}` (net/http, solo librería estándar)")
    w(f"- rust: `{ver_rs}` — axum 0.8.9 sobre tokio 1.54.1 (versiones fijadas con `=` en Cargo.toml y Cargo.lock)")
    w(f"- Docker `{R.sh(['docker', 'version', '--format', '{{.Server.Version}}']).stdout.strip()}`. Servidores con {R.CPUS} CPU y "
      f"{R.MEM}; `hey -c {R.C}` con 4 CPU desde otro contenedor.")
    w(f"- otros contenedores encendidos al empezar (de otros agentes): {len(ajenos)} {ajenos[:6]}\n")
    w("## 1 · Construir la imagen desde cero (`docker build --no-cache`, dependencias incluidas)\n")
    w("| lenguaje | qué se cronometra | segundos |")
    w("|---|---|---|")
    que = {"go": "go build", "rust (glibc)": "cargo build --release (baja y compila axum, tokio y 50 paquetes más)",
           "rust (musl)": "lo mismo, en rust:1.99.1-alpine"}
    for nombre, seg in filas:
        w(f"| {nombre} | {que[nombre]} | {seg:.1f} |")
    w()
    w("## 2 · Lo que despliegas\n")
    w("| imagen | base | tamaño |")
    w("|---|---|---|")
    for nombre, tag, _ in VARIANTES:
        s = int(R.sh(["docker", "image", "inspect", "-f", "{{.Size}}", tag]).stdout.strip())
        w(f"| {nombre} | {BASES[nombre]} | {s / 1048576:.1f} MiB |")
    w("\n`docker image inspect -f {{.Size}}`, como en 651-654. Con el almacén de imágenes de containerd (Docker 29) es el tamaño "
      "comprimido, lo que se descarga: por eso la imagen de Go (2,4 MiB) es más chica que su binario (sección 9).\n")
    R.VARIANTES[:] = VARIANTES
    R.mismo_resultado()
    X.arranque_y_memoria()
    X.carga_max("trabajo", R.N_TRABAJO, "4 · Carga con memoria: /trabajo (objetos, mapa, orden, JSON), a tope")
    X.carga_max("suma", R.N_SUMA, "5 · Carga de CPU pura: /suma, a tope")
    X.ritmo_fijo("trabajo", R.N_TRABAJO, 0, "6 · Latencia de una petición sola: /trabajo")
    X.ritmo_fijo("suma", R.N_SUMA, 0, "7 · Latencia de una petición sola: /suma")
    computo()
    cuesta_escribirlo()
    anexo_hilos()
    w(f"- otros contenedores encendidos al terminar: {len(otros_contenedores())}")
    (AQUI / "SALIDA.md").write_text("\n".join(R.OUT) + "\n", encoding="utf-8")
    print(f"\n[ok] {AQUI / 'SALIDA.md'}")


if __name__ == "__main__":
    try:
        main()
    finally:  # limpieza: contenedor y red propios (las imágenes v698-* se borran a mano al cerrar el vídeo)
        _sh(["docker", "rm", "-f", "v698-srv", "v698-escribir"])
        _sh(["docker", "network", "rm", "v698-red"])
