# -*- coding: utf-8 -*-
"""pruebas_go_rust_revancha.py — «¿Go o Rust? La revancha» (golandia 958, orden del dueño 2026-11-24: «una revancha Go contra
Rust, que piden los comentarios»), con el MISMO banco de la serie «¿Go o X? Lo medí» (879 → 902-905): mismo servidor de Go,
mismo contrato /suma y /trabajo, servidores con 2 CPU y 1 GiB, `hey -c 16` con 4 CPU desde otro contenedor, la comprobación de
que todos devuelven lo mismo ANTES de medir (si no, se para), espera a que la máquina esté en reposo, latencia de una petición
sola y cómputo puro con el reloj de cada proceso.

No se copian las funciones del banco: se importan de pruebas_go_vs_nativo_2026_11 (y este de pruebas_go_vs_lote_2026_11,
pruebas_go_vs_x y pruebas_go_java_revancha). El servidor de Rust es el main.rs del 698 (axum 0.8.9 sobre tokio 1.54.1) sin tocar
la lógica. Lo que el 698 NO midió y aquí sí:
  · perfil «tope»: release + LTO completo + codegen-units=1 + panic=abort;
  · el modo debug (cargo build / cargo run sin --release);
  · el repartidor de memoria: musl con mimalloc y con jemalloc (el arreglo de la trampa del 698) y glibc con mimalloc;
  · otro framework: actix-web 4 en vez de axum;
  · Go también afinado: GOGC=400 (menos recolección a cambio de más memoria);
  · muchas conexiones esperando a la vez (/espera?ms=1000 con 1.000 y 4.000 clientes): gorrutinas contra tareas de tokio;
  · compilación limpia e incremental con y sin LTO, y los binarios de cada variante.

Cambios de operación respecto al banco viejo (encargo 2026-11-24):
  · TODO lo de Docker con prefijo gl-rr- (contenedores, red, imágenes); puerto 9488;
  · TODO `docker run` lleva --rm, --ulimit core=1, nombre gl-rr-… y un TIMEOUT duro (una corrida se colgó 4 h el 23-nov): si
    vence, se borra el contenedor por su nombre, se guarda la SALIDA parcial y se para;
  · urlopen siempre con timeout.

⛔ Solo se ejecuta con el candado de medición tomado (`mkdir <proyecto>/_logs/banco.lock`); el guion lo comprueba.

    python investigacion/pruebas_go_rust_revancha.py <corrida>   → recursos/pruebas-go-rust-revancha/SALIDA-corrida-<corrida>.md
"""
import re
import statistics
import subprocess
import sys
import threading
import time
import urllib.request
from datetime import datetime
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
import pruebas_go_vs_nativo_2026_11 as N  # noqa: E402

L, R, GX = N.L, N.R, N.L.GX
BASE = L.REC / "pruebas-go-rust-revancha"
P = "gl-rr"
PUERTO = 9488
SUBRED = "10.94.88.0/24"
RUST_IMG, RUST_ALP = "rust:1.99.1-trixie", "rust:1.99.1-alpine"

CFG = dict(
    id="rr", nombre="Rust (la revancha)", carpeta="pruebas-go-rust-revancha", subred=SUBRED,
    construir=[
        ("rust (698)", "cargo build --release (axum, glibc)", "Dockerfile.rust", "rust", "build"),
        ("rust tope", "cargo build --profile tope (LTO fat, codegen-units=1, panic=abort)", "Dockerfile.rust-tope", "rust-tope", "build"),
        ("rust + mimalloc", "cargo build --release --features mimalloc (glibc)", "Dockerfile.rust-mimalloc", "rust-mimalloc", "build"),
        ("rust musl (698)", "cargo build --release en rust:1.99.1-alpine", "Dockerfile.rust-musl", "rust-musl", "build"),
        ("rust musl + mimalloc", "lo mismo con --features mimalloc", "Dockerfile.rust-musl-mimalloc", "rust-musl-mimalloc", "build"),
        ("rust musl + jemalloc", "apk add make + --features jemalloc (compila jemalloc)", "Dockerfile.rust-musl-jemalloc",
         "rust-musl-jemalloc", "build"),
        ("actix-web", "cargo build --release (actix-web 4, glibc)", "Dockerfile.actix", "actix", "build"),
        ("rust debug", "cargo build (sin --release)", "Dockerfile.rust-debug", "rust-debug", "build"),
    ],
    variantes=[
        ("go, GOGC=400", "go", {"GOGC": "400"}),
        ("rust (698)", "rust", {}),
        ("rust tope (LTO)", "rust-tope", {}),
        ("rust + mimalloc", "rust-mimalloc", {}),
        ("rust musl (698)", "rust-musl", {}),
        ("rust musl + mimalloc", "rust-musl-mimalloc", {}),
        ("rust musl + jemalloc", "rust-musl-jemalloc", {}),
        ("actix-web", "actix", {}),
        ("rust debug", "rust-debug", {}),
    ],
    bases={"rust": "debian:trixie-slim + binario release (el del 698)",
           "rust-tope": "debian:trixie-slim + binario con LTO fat + codegen-units=1 + panic=abort",
           "rust-mimalloc": "debian:trixie-slim + binario release con mimalloc",
           "rust-musl": "scratch + binario estático musl (el del 698)",
           "rust-musl-mimalloc": "scratch + binario estático musl con mimalloc",
           "rust-musl-jemalloc": "scratch + binario estático musl con jemalloc",
           "actix": "debian:trixie-slim + binario release de actix-web",
           "rust-debug": "debian:trixie-slim + binario debug (sin --release)"},
    binarios=[(k, "/servidor") for k in ("rust", "rust-tope", "rust-mimalloc", "rust-musl", "rust-musl-mimalloc",
                                         "rust-musl-jemalloc", "actix", "rust-debug")],
)

# ── operación: nombres gl-rr-, --rm, core=1 y timeout duro en cada docker run ─────────────────────────────────────────
_n = [0]
_sh_mapeado = None   # L.sh (traduce srv → gl-rr-srv, puerto, hey)


class Colgado(Exception):
    pass


def _timeout(a):
    if a[:2] == ["docker", "build"]:
        return 1800
    if a[:2] == ["docker", "run"]:
        if "-d" in a:
            return 120
        if "-z" in a:
            seg = int(str(a[a.index("-z") + 1]).rstrip("s"))
            return seg + 120
        return 1200   # compilaciones de «lo que cuesta escribirlo», versiones, stat…
    return 180


def sh(args, **kw):
    a = [str(x) for x in args]
    nombre = None
    if a[:2] == ["docker", "run"]:
        extra = []
        if "--rm" not in a:
            extra.append("--rm")
        if "--ulimit" not in a:
            extra += ["--ulimit", "core=1"]
        if "--name" in a:
            nombre = a[a.index("--name") + 1]
        else:
            _n[0] += 1
            nombre = f"{P}-tmp-{_n[0]}"
            extra += ["--name", nombre]
        a = a[:2] + extra + a[2:]
        nombre = L.MAPA.get(nombre, nombre)
    kw.setdefault("timeout", _timeout(a))
    try:
        return _sh_mapeado(a, **kw)
    except subprocess.TimeoutExpired:
        if nombre:
            subprocess.run(["docker", "rm", "-f", nombre], capture_output=True, timeout=120)
        raise Colgado(f"⛔ timeout de {kw['timeout']} s en: {' '.join(a)[:300]}")


_urlopen_mapeado = None


def urlopen(url, *a, **kw):
    if not a:
        kw.setdefault("timeout", 60)
    return _urlopen_mapeado(url, *a, **kw)


def img(cfg, k):
    return f"{P}-{k}:1"


def preparar(cfg):
    global _sh_mapeado, _urlopen_mapeado
    L.PUERTO = PUERTO
    L.MAPA.update({"srv": f"{P}-srv", "golandia-revancha-hey:1": f"{P}-hey:1", "9300:8080": f"{PUERTO}:8080"})
    R.RED = f"{P}-red"
    _sh_mapeado = L.sh
    R.sh = sh
    _urlopen_mapeado = L.urlopen
    urllib.request.urlopen = urlopen


N.img = L.img = img
N.preparar = L.preparar = preparar


def ajenos():
    r = R.sh(["docker", "stats", "--no-stream", "--format", "{{.Name}}\t{{.CPUPerc}}"])
    return [tuple(l.split("\t")) for l in r.stdout.splitlines() if "\t" in l and not l.startswith(P + "-")]


N.ajenos = ajenos


def cpu_windows():
    try:
        r = subprocess.run(["powershell", "-NoProfile", "-Command", "(Get-CimInstance Win32_Processor | Measure-Object "
                            "-Property LoadPercentage -Average).Average"], capture_output=True, text=True, timeout=60)
        return r.stdout.strip() + " %"
    except Exception as e:  # noqa: BLE001
        return f"(no se pudo leer: {e})"


# ── secciones nuevas ──────────────────────────────────────────────────────────────────────────────────────────────────
def lineas_listo():
    """Lo que cada servidor dice de sí mismo al arrancar (hilos que ve, repartidor, perfil)."""
    R.w("## 3b · Lo que dice cada servidor al arrancar (con --cpus 2)\n")
    for nombre, imagen, env in R.VARIANTES:
        t0 = R.arrancar(nombre, imagen, env, puerto=9300)
        R.esperar_200(9300, t0)
        r = R.sh(["docker", "logs", "srv"])
        m = re.search(r"LISTO_EN_MS[^\n]*", r.stderr + r.stdout)
        R.sh(["docker", "rm", "-f", "srv"])
        R.w(f"- {nombre} → `{m.group(0).strip() if m else '(sin línea LISTO_EN_MS)'}`")
    R.w()


def hey_texto(url, seg, c):
    r = R.sh(["docker", "run", "--network", R.RED, "--cpus", "4", "golandia-revancha-hey:1", "-z", f"{seg}s", "-c", str(c),
              "-t", "30", url])
    return r.stdout


def muchas_conexiones(variantes, ms, clientes, seg):
    R.w(f"## 9 · Muchas conexiones esperando a la vez: /espera?ms={ms}\n")
    R.w(f"Cada petición duerme {ms} ms sin gastar CPU (`time.Sleep` en Go; `tokio::time::sleep` en axum y en actix). Con N "
        f"clientes el techo teórico es N × {1000 // ms} peticiones/s: aquí no se mide cálculo, se mide cuánto cuesta tener N "
        f"conexiones abiertas a la vez. `hey -c N -z {seg}s -t 30` desde otro contenedor con 4 CPU, tras {R.SEG_FRIO} s de "
        "calentamiento con los mismos clientes; servidor con 2 CPU y 1 GiB. Memoria máxima con `docker stats` durante la medida.\n")
    R.w("| variante | clientes | peticiones/s | p50 | p99 | errores | memoria máx. |")
    R.w("|---|---|---|---|---|---|---|")
    for nombre, imagen, env in variantes:
        for c in clientes:
            t0 = R.arrancar(nombre, imagen, env, puerto=9300)
            R.esperar_200(9300, t0)
            url = f"http://srv:8080/espera?ms={ms}"
            hey_texto(url, R.SEG_FRIO, c)
            pico = [0.0]
            fin = threading.Event()

            def muestrear():
                while not fin.is_set():
                    try:
                        pico[0] = max(pico[0], R.mem_mib())
                    except Exception:  # noqa: BLE001
                        pass

            th = threading.Thread(target=muestrear)
            th.start()
            t = hey_texto(url, seg, c)
            fin.set()
            th.join()
            R.sh(["docker", "rm", "-f", "srv"])
            rps = float(re.search(r"Requests/sec:\s+([\d.]+)", t).group(1))
            lat = {q: float(re.search(rf"\s{q}% in ([\d.]+) secs", t).group(1)) * 1000 for q in (50, 99)}
            ok = sum(int(n) for cod, n in re.findall(r"\[(\d+)\]\s+(\d+) responses", t) if cod == "200")
            otros = [(cod, n) for cod, n in re.findall(r"\[(\d+)\]\s+(\d+) responses", t) if cod != "200"]
            err = re.search(r"Error distribution:\n((?:\s+\[\d+\][^\n]*\n?)+)", t)
            nerr = sum(int(x) for x in re.findall(r"\[(\d+)\]", err.group(1))) if err else 0
            nerr += sum(int(n) for _, n in otros)
            R.w(f"| {nombre} | {c} | {rps:.0f} | {lat[50]:.1f} ms | {lat[99]:.1f} ms | {nerr} (de {ok + nerr}) | {pico[0]:.0f} MiB |")
            if err:
                R.w(f"|  ↳ errores | | | | | `{' · '.join(l.strip() for l in err.group(1).strip().splitlines())[:200]}` | |")
    R.w()


REPES = 3
GO_ESCRIBIR = r"""set -e
cp -r /src /tmp/w && cd /tmp/w
t() { a=$(date +%s.%N); "$@" > /tmp/log 2>&1 || { cat /tmp/log >&2; exit 1; }; b=$(date +%s.%N); echo "$a $b"; }
echo "T limpia $(t env CGO_ENABLED=0 go build -ldflags='-s -w' -o /tmp/s .)"
sed -i 's/"lang":"go","n":%d,"suma":0/"lang":"go!","n":%d,"suma":0/' main.go && grep -q 'go!' main.go
echo "T incremental $(t env CGO_ENABLED=0 go build -ldflags='-s -w' -o /tmp/s .)"
echo "B go $(stat -c %s /tmp/s)"
"""
RUST_ESCRIBIR = r"""set -e
mkdir /tmp/w && cp -r /src/Cargo.toml /src/Cargo.lock /src/src /tmp/w/ && cd /tmp/w
t() { a=$(date +%s.%N); "$@" > /tmp/log 2>&1 || { cat /tmp/log >&2; exit 1; }; b=$(date +%s.%N); echo "$a $b"; }
echo "T descarga $(t cargo fetch --locked)"
echo "T limpia $(t cargo build --release --locked --offline)"
sed -i 's/"lang":"rust","n":{ms}/"lang":"rust!","n":{ms}/' src/main.rs && grep -q 'rust!' src/main.rs
echo "T incremental $(t cargo build --release --locked --offline)"
echo "T limpia-tope $(t cargo build --profile tope --locked --offline)"
sed -i 's/"lang":"rust!","n":{ms}/"lang":"rust?","n":{ms}/' src/main.rs && grep -q 'rust?' src/main.rs
echo "T incremental-tope $(t cargo build --profile tope --locked --offline)"
echo "T limpia-debug $(t cargo build --locked --offline)"
sed -i 's/"lang":"rust?","n":{ms}/"lang":"rust:","n":{ms}/' src/main.rs && grep -q 'rust:' src/main.rs
echo "T incremental-debug $(t cargo build --locked --offline)"
"""


def escribir_una(imagen, src, guion, nombre):
    r = R.sh(["docker", "run", "--name", nombre, "--cpus", "2", "-v", f"{R.win(src)}:/src:ro", "--entrypoint", "sh",
              imagen, "-c", guion])
    assert r.returncode == 0, r.stdout[-1500:] + r.stderr[-1500:]
    t = {}
    for linea in r.stdout.splitlines():
        p = linea.split()
        if p and p[0] == "T":
            t[p[1]] = float(p[3]) - float(p[2])
    return t


def lineas(f):
    return sum(1 for l in f.read_text(encoding="utf-8").splitlines() if l.strip() and not l.strip().startswith("//"))


def paquetes(lock):
    return len(re.findall(r"^\[\[package\]\]", lock.read_text(encoding="utf-8"), re.M)) - 1


def cuesta_escribirlo():
    R.w("## 10 · Lo que cuesta compilarlo\n")
    go_src, ax, ac = GX.GO_SRC / "main.go", BASE / "axum" / "src" / "main.rs", BASE / "actix" / "src" / "main.rs"
    R.w("| | go | rust axum | rust actix-web |")
    R.w("|---|---|---|---|")
    R.w(f"| líneas de código (sin blancos ni comentarios) | {lineas(go_src)} | {lineas(ax)} (el 698: {lineas(L.REC / 'pruebas-go-vs-rust' / 'src' / 'main.rs')}; lo demás es el repartidor) | {lineas(ac)} |")
    R.w(f"| paquetes en Cargo.lock | 0 (solo librería estándar) | {paquetes(BASE / 'axum' / 'Cargo.lock')} (incluye mimalloc y "
        f"jemalloc opcionales) | {paquetes(BASE / 'actix' / 'Cargo.lock')} |")
    R.w()
    R.w(f"Compilación dentro de `golang:1.27.2` y `{RUST_IMG}` con `--cpus 2`, contenedor nuevo cada vez (cachés vacías), "
        f"mediana de {REPES}. «Incremental»: cambiar una cadena en el manejador de /espera y volver a compilar. «tope» = LTO fat + "
        "codegen-units=1 + panic=abort.\n")
    res = {"go": [], "rust": []}
    for i in range(REPES):
        res["go"].append(escribir_una("golang:1.27.2", GX.GO_SRC, GO_ESCRIBIR, f"{P}-escribir-go-{i}"))
        res["rust"].append(escribir_una(RUST_IMG, BASE / "axum", RUST_ESCRIBIR, f"{P}-escribir-rust-{i}"))

    def med(lang, k):
        return statistics.median(x[k] for x in res[lang])

    def todas(lang, k):
        d = 2 if k.startswith("incremental") else 1
        return " / ".join(f"{x[k]:.{d}f}" for x in res[lang])

    R.w(f"| paso | segundos (mediana) | las {REPES} corridas |")
    R.w("|---|---|---|")
    R.w(f"| go: compilación limpia | {med('go', 'limpia'):.1f} | {todas('go', 'limpia')} |")
    R.w(f"| go: incremental | {med('go', 'incremental'):.2f} | {todas('go', 'incremental')} |")
    R.w(f"| rust: bajar dependencias | {med('rust', 'descarga'):.1f} | {todas('rust', 'descarga')} |")
    for k, txt in [("limpia", "rust release: compilación limpia"), ("incremental", "rust release: incremental"),
                   ("limpia-tope", "rust tope (LTO): compilación limpia"), ("incremental-tope", "rust tope (LTO): incremental"),
                   ("limpia-debug", "rust debug: compilación limpia"), ("incremental-debug", "rust debug: incremental")]:
        d = 2 if k.startswith("incremental") else 1
        R.w(f"| {txt} | {med('rust', k):.{d}f} | {todas('rust', k)} |")
    R.w()


def versiones():
    ver_go = R.sh(["docker", "run", "golang:1.27.2", "go", "version"]).stdout.strip()
    rv = R.sh(["docker", "run", RUST_IMG, "sh", "-c", "rustc --version; cargo --version"]).stdout.strip().splitlines()
    ra = R.sh(["docker", "run", RUST_ALP, "sh", "-c", "rustc -vV | grep -E '^host'"]).stdout.strip()

    def ver(lock, nombre):
        m = re.search(rf'name = "{nombre}"\nversion = "([^"]+)"', lock.read_text(encoding="utf-8"))
        return m.group(1) if m else "?"

    la, lc = BASE / "axum" / "Cargo.lock", BASE / "actix" / "Cargo.lock"
    R.w(f"- go: `{ver_go}` (net/http, solo librería estándar)")
    R.w(f"- rust: `{' · '.join(rv)}` · alpine `{ra}`")
    R.w(f"- crates (Cargo.lock): axum {ver(la, 'axum')} · tokio {ver(la, 'tokio')} · mimalloc {ver(la, 'mimalloc')} · "
        f"tikv-jemallocator {ver(la, 'tikv-jemallocator')} · actix-web {ver(lc, 'actix-web')} (su tokio {ver(lc, 'tokio')})")
    R.w(f"- Docker `{R.sh(['docker', 'version', '--format', '{{.Server.Version}}']).stdout.strip()}`. Servidores con {R.CPUS} CPU y "
        f"{R.MEM}; `hey -c {R.C}` con 4 CPU desde otro contenedor.\n")


def main():
    global REPES
    corrida = sys.argv[1]
    cfg = CFG
    humo = "--humo" in sys.argv   # prueba del guion: imágenes con caché, 2 variantes, tiempos cortos; NO es una medida
    if humo:
        corrida = "humo"
        R.SEG_FRIO, R.SEG_CALIENTE, R.REPES_FRIO, REPES = 2, 3, 1, 1
        cfg = dict(CFG, construir=CFG["construir"][:1], variantes=CFG["variantes"][1:2], bases={"rust": CFG["bases"]["rust"]},
                   binarios=CFG["binarios"][:1])
    if not L.CANDADO.is_dir():
        raise SystemExit(f"⛔ sin candado: toma {L.CANDADO} antes de medir (mkdir)")
    preparar(cfg)
    R.OUT.clear()
    R.sh(["docker", "rm", "-f", "srv"])
    R.sh(["docker", "network", "rm", R.RED])
    r = R.sh(["docker", "network", "create", "--subnet", cfg["subred"], R.RED])
    assert r.returncode == 0, r.stderr
    destino = BASE / ("_humo/SALIDA-humo.md" if humo else f"SALIDA-corrida-{corrida}.md")
    destino.parent.mkdir(exist_ok=True)
    try:
        R.w(f"# SALIDA — ¿Go o Rust? La revancha (golandia 958), corrida {corrida}, con el banco de la serie\n")
        R.w(f"> Corrida del {datetime.now():%Y-%m-%d %H:%M}, con el candado `_logs/banco.lock` tomado. Regenerable con "
            f"`python investigacion/pruebas_go_rust_revancha.py {corrida}` (mismo banco que 879, 865 y 902-905; el 698 es "
            "`recursos/pruebas-go-vs-rust/SALIDA.md`). ⛔ Ninguna cifra de blog: todo salió de aquí.\n")
        if humo:
            for _n2, _q, df, k, _e in cfg["construir"]:
                R.sh(["docker", "build", "-q", "--target", "build", "-f", str(BASE / df), "-t", img(cfg, k) + "-etapa", str(BASE)])
            L.GO_DF.write_text(GX.GO_DF, encoding="utf-8")
            R.sh(["docker", "build", "-q", "--target", "build", "-f", str(L.GO_DF), "-t", img(cfg, "go") + "-etapa", str(GX.GO_SRC)])
            R.sh(["docker", "build", "-q", "-f", str(L.GO_DF), "-t", img(cfg, "go"), str(GX.GO_SRC)])
            R.sh(["docker", "build", "-q", "-f", str(L.HEY_CTX / "Dockerfile.hey"), "-t", L.MAPA["golandia-revancha-hey:1"],
                  str(L.HEY_CTX)])
            filas = []
        else:
            filas = L.construir(cfg, BASE)
        versiones()
        R.w("## 1 · Construir la imagen desde cero (`docker build --no-cache`, dependencias incluidas)\n")
        R.w("| variante | qué se cronometra | segundos |")
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
        R.w("\n| binario (lo que se copia a la imagen) | bytes | MiB |")
        R.w("|---|---|---|")
        b = L.binario_mib(img(cfg, "go") + "-etapa", "/servidor")
        R.w(f"| go (`-ldflags='-s -w'`) | {round(b * 1048576)} | {b:.2f} |")
        for k, ruta in cfg["binarios"]:
            b = L.binario_mib(img(cfg, k) + "-etapa", ruta)
            R.w(f"| {k} | {round(b * 1048576)} | {b:.2f} |")
        R.w()
        R.w(f"- CPU de Windows (Win32_Processor.LoadPercentage) antes de medir: {cpu_windows()}")
        if not humo:
            N.esperar_reposo(cfg)
        R.VARIANTES[:] = [("go", img(cfg, "go"), {})] + [(n, img(cfg, k), env) for n, k, env in cfg["variantes"]]
        R.mismo_resultado()
        GX.arranque_y_memoria()
        lineas_listo()
        GX.carga_max("trabajo", R.N_TRABAJO, "4 · Carga con memoria: /trabajo (objetos, mapa, orden, JSON), a tope")
        GX.carga_max("suma", R.N_SUMA, "5 · Carga de CPU pura: /suma, a tope")
        if not humo:   # en humo, hey con 15 peticiones no imprime percentiles: estas dos y el cómputo son del banco tal cual
            GX.ritmo_fijo("trabajo", R.N_TRABAJO, 0, "6 · Latencia de una petición sola: /trabajo")
            GX.ritmo_fijo("suma", R.N_SUMA, 0, "7 · Latencia de una petición sola: /suma")
            L.computo()
        vm = [v for v in R.VARIANTES if v[0] in ("go", "rust (698)", "actix-web")]
        muchas_conexiones(vm, 1000, (1000, 4000), R.SEG_CALIENTE)
        cuesta_escribirlo()
        c, linea = N.carga_1min()
        R.w(f"Al terminar: /proc/loadavg `{linea}`; CPU de Windows {cpu_windows()}; ajenos: "
            + (", ".join(f"{n} {c}" for n, c in ajenos()) or "ninguno") + "\n")
    except Colgado as e:
        R.w(f"\n{e}\n")
        destino = BASE / f"SALIDA-corrida-{corrida}.PARCIAL.md"
        raise
    finally:
        destino.write_text("\n".join(R.OUT) + "\n", encoding="utf-8")
        R.sh(["docker", "rm", "-f", "srv"])
        R.sh(["docker", "network", "rm", R.RED])
        print(f"\n[fin] {destino}")


if __name__ == "__main__":
    main()
