# -*- coding: utf-8 -*-
"""correr.py — «Granian, Robyn y Litestar contra Go» (golandia, pedido por @mdt4177 en el vídeo 770 de Cultura DevOps).

El comentario: en la web las empresas cambian Uvicorn por Granian, o usan Robyn y Litestar, «hechos con Rust y libres de GIL».
¿Cómo quedan contra Go? El MISMO banco que el 770 (cultura-devops/recursos/pruebas-frameworks-python) y que la serie «¿Go o X?»:
  · la MISMA lógica (app/logica.py, copiada del 770) en todos, y el MISMO servidor Go con net/http
    (recursos/pruebas-go-java-revancha/go, imagen de investigacion/pruebas_go_vs_x.py);
  · FastAPI y Litestar en la MISMA imagen con uvicorn y con Granian: cambia solo el servidor (SRV en app/arrancar.sh);
  · Robyn 0.88 (runtime en Rust) con 2 procesos. ⚠️ No --fast: calcula los procesos con os.cpu_count(), que dentro del contenedor
    ve todos los núcleos del PC y no los 2 que tiene;
  · servidores con 2 CPU y 1 GiB; `hey -c 16` con 4 CPU desde otro contenedor; versiones fijadas en req/<fw>.txt;
  · antes de medir, se comprueba que todos devuelven el MISMO cuerpo (quitando "lang" y "ns").
Añadido: sección G, «¿libres de GIL?»: UN proceso con 1 y con 2 hilos de trabajo, y la CPU que llega a usar (muestreo.py).

    python correr.py                  → SALIDA.md (~25 min, la máquina sin otras cargas)
    python correr.py --comprobar      → solo construir + sección 0, sin carga
    python correr.py --solo-gil       → solo la sección G (SALIDA-gil.md)
⛔ Ninguna cifra de blog: lo que diga el vídeo sale de la SALIDA de aquí.
"""
import json
import re
import statistics
import sys
import time
import urllib.request
from datetime import datetime
from pathlib import Path

AQUI = Path(__file__).resolve().parent
G = Path(r"<proyecto>\canales\golandia")
sys.path.insert(0, str(G / "investigacion"))
import pruebas_go_java_revancha as R  # noqa: E402
import pruebas_go_vs_x as X  # noqa: E402
from muestreo import Muestreo  # noqa: E402

R.RED = "fwpy"  # la red del banco del 770 (Docker no tenía rangos libres para una nueva)
HEY_IMG = "golandia-revancha-hey:1"
GO_IMG = X.GO_IMG
FWS = ("fastapi", "litestar", "robyn")
VARIANTES = [  # (nombre, imagen, entorno)
    ("go (net/http)", GO_IMG, {}),
    ("fastapi + uvicorn", "golandia-rust-py-fastapi:1", {"SRV": "uvicorn"}),
    ("fastapi + granian", "golandia-rust-py-fastapi:1", {"SRV": "granian"}),
    ("litestar + uvicorn", "golandia-rust-py-litestar:1", {"SRV": "uvicorn"}),
    ("litestar + granian", "golandia-rust-py-litestar:1", {"SRV": "granian"}),
    ("robyn", "golandia-rust-py-robyn:1", {"SRV": "robyn"}),
]
REF = "go (net/http)"
SERIES = AQUI / "series"

# ⚠️ 2026-11-09, primera corrida: Robyn responde en cuanto su PRIMER proceso está listo, con el segundo aún arrancando. Si la
# carga empieza ahí, las 16 conexiones keep-alive se quedan en un proceso: 122 pet/s en vez de 240, en 3 de 10 arranques
# (SALIDA.md, sección R). Así que antes de medir se espera a que TODOS los procesos hayan escrito su LISTO_EN_MS.
_arrancar, _esperar = R.arrancar, R.esperar_200
_procesos = [1]


def arrancar(nombre, imagen, env, puerto=None):
    _procesos[0] = 1 if imagen == GO_IMG else int(env.get("WORKERS", "2"))
    return _arrancar(nombre, imagen, env, puerto)


def esperar_200(puerto, t0):
    ms = _esperar(puerto, t0)  # lo que se apunta como «1ª respuesta» sigue siendo la primera 200
    while True:
        log = R.sh(["docker", "logs", "srv"])
        if len(re.findall(r"LISTO_EN_MS", log.stdout + log.stderr)) >= _procesos[0]:
            return ms
        if time.perf_counter() - t0 > 30:
            raise SystemExit(f"no arrancaron los {_procesos[0]} procesos")
        time.sleep(0.02)


R.arrancar, R.esperar_200 = arrancar, esperar_200


def img(fw):
    return f"golandia-rust-py-{fw}:1"


def construir_py(fw):
    base = ["docker", "build", "-q", "--build-arg", f"FW={fw}", "-f", str(AQUI / "Dockerfile"), "-t", img(fw)]
    r = R.sh(base + [str(AQUI)])
    assert r.returncode == 0, r.stderr[-2000:]
    t0 = time.perf_counter()
    r = R.sh(base[:3] + ["--no-cache"] + base[3:] + [str(AQUI)])
    seg = time.perf_counter() - t0
    assert r.returncode == 0, r.stderr[-2000:]
    return seg


def construir_go():
    df = G / "recursos" / "pruebas-go-vs-comun" / "Dockerfile.go"
    df.parent.mkdir(exist_ok=True)
    df.write_text(X.GO_DF, encoding="utf-8")
    base = ["docker", "build", "-q", "-f", str(df), "-t", GO_IMG]
    ctx = str(X.GO_SRC)
    r = R.sh(base + [ctx])
    assert r.returncode == 0, r.stderr[-2000:]
    t0 = time.perf_counter()
    r = R.sh(base[:3] + ["--no-cache"] + base[3:] + [ctx])
    seg = time.perf_counter() - t0
    assert r.returncode == 0, r.stderr[-2000:]
    return seg


def pedir(ruta, n, puerto=9300):
    with urllib.request.urlopen(f"http://127.0.0.1:{puerto}/{ruta}?n={n}") as r:
        return r.read(), r.headers.get("Content-Type")


def normal(b):
    return re.sub(rb',"ns":\d+}$', b"}", re.sub(rb'"lang":"[a-z]+",', b"", b))


def mismo_resultado():
    R.w("## 0 · ¿Hacen lo mismo? (antes de medir)\n")
    R.w(f"El cuerpo de /trabajo?n={R.N_TRABAJO}, /suma?n={R.N_SUMA} y /suma?n=1, comparado byte a byte con {REF} quitando solo "
        "\"lang\" (go o python) y \"ns\" (el reloj del cálculo).\n")
    salidas = {}
    for nombre, imagen, env in R.VARIANTES:
        t0 = R.arrancar(nombre, imagen, env, puerto=9300)
        R.esperar_200(9300, t0)
        res = []
        for ruta, n in (("trabajo", R.N_TRABAJO), ("suma", R.N_SUMA), ("suma", 1)):
            cuerpo, ct = pedir(ruta, n)
            res.append((normal(cuerpo), ct.split(";")[0].strip().lower()))
        salidas[nombre] = res
        R.sh(["docker", "rm", "-f", "srv"])
    ref = salidas[REF]
    for nombre, v in salidas.items():
        ok = v == ref
        j = json.loads(v[0][0])
        s = json.loads(v[1][0])
        R.w(f"- {nombre}: total {j['total']}, primeros ids {[x['id'] for x in j['primeros']][:5]}…, suma {s['suma']}, "
            f"{len(v[0][0])} + {len(v[1][0])} + {len(v[2][0])} bytes, {v[0][1]} → {'idéntico' if ok else '⛔ DISTINTO'}")
        if not ok:
            raise SystemExit("las variantes no devuelven lo mismo: no se compara")
    R.w()


def computo_una():
    R.w("## A · Cómputo de UNA petición (mediana del campo ns, 30 de calentamiento + 200 medidas, de una en una)\n")
    R.w("| variante | /trabajo (20.000 objetos) | /suma (200.000 vueltas) |")
    R.w("|---|---|---|")
    for nombre, imagen, env in R.VARIANTES:
        t0 = R.arrancar(nombre, imagen, env, puerto=9300)
        R.esperar_200(9300, t0)
        res = []
        for ruta, n in (("trabajo", R.N_TRABAJO), ("suma", R.N_SUMA)):
            ns = []
            for i in range(230):
                j = json.loads(pedir(ruta, n)[0])
                if i >= 30:
                    ns.append(j["ns"])
            res.append(statistics.median(ns) / 1e6)
        R.sh(["docker", "rm", "-f", "srv"])
        R.w(f"| {nombre} | {res[0]:.2f} ms | {res[1]:.3f} ms |")
    R.w()


GIL = [  # (nombre, imagen, entorno): UN proceso; ¿un segundo hilo de trabajo duplica las peticiones de CPU?
    ("go, 1 núcleo (GOMAXPROCS=1)", GO_IMG, {"GOMAXPROCS": "1"}),
    ("go, 2 núcleos (GOMAXPROCS=2)", GO_IMG, {"GOMAXPROCS": "2"}),
    ("robyn, 1 proceso × 1 worker", img("robyn"), {"SRV": "robyn", "WORKERS": "1", "HILOS": "1"}),
    ("robyn, 1 proceso × 2 workers", img("robyn"), {"SRV": "robyn", "WORKERS": "1", "HILOS": "2"}),
    ("granian, 1 proceso (fastapi def)", img("fastapi"), {"SRV": "granian", "WORKERS": "1"}),
    ("robyn, 2 procesos × 1 worker", img("robyn"), {"SRV": "robyn", "WORKERS": "2", "HILOS": "1"}),
]


def sin_gil():
    R.w("## G · ¿Libres de GIL? UN proceso con 1 y con 2 hilos de trabajo, /suma (CPU pura) a tope\n")
    R.w(f"`hey -c {R.C}` {R.SEG_CALIENTE} s tras {R.SEG_FRIO} s de calentamiento; servidor con {R.CPUS} CPU y {R.MEM}. CPU medida "
        "en el cgroup cada 0,5 s (muestreo.py): 100 % = un núcleo entero, el techo con 2 CPU es 200 %. Series en series/gil-*.json.\n")
    R.w("| variante | peticiones/s | p50 | CPU media | CPU máx. | hilos (máx.) |")
    R.w("|---|---|---|---|---|---|")
    SERIES.mkdir(exist_ok=True)
    url = f"http://srv:8080/suma?n={R.N_SUMA}"
    for nombre, imagen, env in GIL:
        t0 = R.arrancar(nombre, imagen, env, puerto=9300)
        R.esperar_200(9300, t0)
        R.hey(url, R.SEG_FRIO)
        with Muestreo("srv") as m:
            rps, lat = R.hey(url, R.SEG_CALIENTE)
        clave = re.sub(r"[^a-z0-9]+", "-", nombre.lower()).strip("-")
        m.guardar(SERIES / f"gil-{clave}.json", csv_tambien=False, extra={"variante": nombre, "rps": rps})
        r = m.resumen(desde=1.0)
        R.sh(["docker", "rm", "-f", "srv"])
        R.w(f"| {nombre} | {rps:.0f} | {lat[50]:.1f} ms | {r['cpu_media']:.0f} % | {r['cpu_max']:.0f} % | {r['hilos_max']} |")
    R.w()


def serie_carga(ruta, n, etiqueta):
    """CPU y memoria en el tiempo de cada variante principal bajo la carga dada (para las gráficas del vídeo)."""
    SERIES.mkdir(exist_ok=True)
    url = f"http://srv:8080/{ruta}?n={n}"
    for nombre, imagen, env in R.VARIANTES:
        t0 = R.arrancar(nombre, imagen, env, puerto=9300)
        R.esperar_200(9300, t0)
        with Muestreo("srv") as m:
            rps, _ = R.hey(url, R.SEG_CALIENTE)
        clave = re.sub(r"[^a-z0-9]+", "-", nombre.lower()).strip("-")
        m.guardar(SERIES / f"{etiqueta}-{clave}.json", csv_tambien=False, extra={"variante": nombre, "rps": rps})
        R.sh(["docker", "rm", "-f", "srv"])


def robyn_arranque(repes=10):
    """La trampa del arranque de Robyn: la carga empieza con la PRIMERA respuesta, o cuando los dos procesos están listos."""
    R.w(f"## R · Robyn: ¿empezar la carga con la primera respuesta? ({repes} arranques de cada forma, /suma, `hey -c {R.C}` 4 s)\n")
    R.w("Contenedor nuevo cada vez. «Primera 200»: la carga arranca en cuanto responde una petición. «Dos LISTO»: espera a que los "
        "dos procesos escriban su marca. «Un proceso» = la corrida quedó por debajo de 180 pet/s.\n")
    R.w("| cuándo empieza la carga | pet/s de cada arranque | con un proceso |")
    R.w("|---|---|---|")
    url = f"http://srv:8080/suma?n={R.N_SUMA}"
    for etiqueta, esperar in (("primera 200", _esperar), ("dos LISTO", esperar_200)):
        vals = []
        for _ in range(repes):
            t0 = arrancar("robyn", img("robyn"), {"SRV": "robyn"}, puerto=9300)
            esperar(9300, t0)
            vals.append(R.hey(url, 4)[0])
            R.sh(["docker", "rm", "-f", "srv"])
        R.w(f"| {etiqueta} | {', '.join(f'{v:.0f}' for v in vals)} | {sum(v < 180 for v in vals)} de {repes} |")
    R.w()


def cpu_por_proceso():
    """%CPU de cada proceso del contenedor (docker top), sin el proceso padre que no atiende."""
    t = R.sh(["docker", "top", "srv", "o", "pid,pcpu,args"]).stdout.splitlines()[1:]
    return sorted((float(x.split()[1]) for x in t if x.split()), reverse=True)


def balanceo(repes=30):
    """Sección B: ¿cuántos arranques acaban con TODAS las conexiones en un solo proceso? (2 procesos, /suma, keep-alive)."""
    R.w(f"## B · Reparto entre los 2 procesos: {repes} arranques de cada uno (/suma, `hey -c {R.C}` 6 s)\n")
    R.w("Contenedor nuevo cada vez; la carga empieza cuando los DOS procesos escribieron LISTO_EN_MS. A los 4 s se lee el %CPU de cada "
        "proceso (`docker top`). «Medio servidor»: el arranque quedó por debajo del 75 % de la mediana de sus propias corridas.\n")
    R.w("| variante | mediana pet/s | arranques a medio servidor | pet/s de esos arranques | %CPU de los 2 procesos que más usan, en uno de ellos |")
    R.w("|---|---|---|---|---|")
    url = f"http://srv:8080/suma?n={R.N_SUMA}"
    solo = sys.argv[sys.argv.index("--b-solo") + 1].split(",") if "--b-solo" in sys.argv else None  # reanudar (apagón 2026-11-09)
    for nombre, imagen, env in [v for v in VARIANTES if v[1] != GO_IMG and (solo is None or v[0] in solo)]:
        vals, cpus = [], []
        for _ in range(repes):
            t0 = arrancar(nombre, imagen, env, puerto=9300)
            esperar_200(9300, t0)
            hc = "heyb"
            R.sh(["docker", "rm", "-f", hc])
            R.sh(["docker", "run", "-d", "--name", hc, "--network", R.RED, "--cpus", "4", HEY_IMG, "-z", "6s", "-c", str(R.C), url])
            time.sleep(4)
            cpus.append(cpu_por_proceso()[:3])
            while R.sh(["docker", "inspect", "-f", "{{.State.Running}}", hc]).stdout.strip() == "true":
                time.sleep(0.2)
            t = R.sh(["docker", "logs", hc]).stdout
            R.sh(["docker", "rm", "-f", hc])
            vals.append(float(re.search(r"Requests/sec:\s+([\d.]+)", t).group(1)))
            R.sh(["docker", "rm", "-f", "srv"])
        med = statistics.median(vals)
        malos = [i for i, v in enumerate(vals) if v < 0.75 * med]
        ej = cpus[malos[0]] if malos else cpus[0]
        R.w(f"| {nombre} | {med:.0f} | {len(malos)} de {repes} | {', '.join(f'{vals[i]:.0f}' for i in malos) or '—'} | "
            f"{' · '.join(f'{c:.0f} %' for c in ej)} |")
    R.w()


def main():
    comprobar = "--comprobar" in sys.argv
    if "--solo-balanceo" in sys.argv:
        R.OUT.clear()
        R.VARIANTES[:] = VARIANTES
        if "--b-solo" not in sys.argv:
            balanceo()
        else:  # las filas que ya están en _log-balanceo.txt van primero, tal cual se midieron
            previas = [x for x in (AQUI / "_log-balanceo.txt").read_text(encoding="utf-8").splitlines() if x.strip()]
            balanceo()
            nuevas = [x for x in R.OUT if x.startswith("| ") and not x.startswith("| variante")]
            R.OUT[:] = previas + nuevas + ["", "(Las tres primeras filas, del 2026-11-09 antes de un apagón; las demás, de la reanudación "
                                            "con --b-solo, en la misma máquina y con el mismo método.)", ""]
        salida = AQUI / "SALIDA.md"
        salida.write_text(salida.read_text(encoding="utf-8").rstrip("\n") + "\n\n" + "\n".join(R.OUT) + "\n", encoding="utf-8")
        print(f"\n[ok] sección B añadida a {salida}")
        return
    R.OUT.clear()
    R.sh(["docker", "network", "create", R.RED])
    if R.sh(["docker", "image", "inspect", HEY_IMG]).returncode != 0:
        r = R.sh(["docker", "build", "-q", "-f", str(G / "recursos" / "pruebas-go-java-revancha" / "Dockerfile.hey"), "-t", HEY_IMG,
                  str(G / "recursos" / "pruebas-go-java-revancha")])
        assert r.returncode == 0, r.stderr[-1500:]
    R.VARIANTES[:] = VARIANTES

    if "--solo-gil" in sys.argv:
        R.w(f"# SALIDA (solo sección G) — corrida del {datetime.now():%Y-%m-%d %H:%M}\n")
        sin_gil()
        (AQUI / "SALIDA-gil.md").write_text("\n".join(R.OUT) + "\n", encoding="utf-8")
        return

    R.w("# SALIDA — Granian, Robyn y Litestar contra Go (golandia; pedido por @mdt4177 en el 770 de Cultura DevOps)\n")
    R.w(f"> Corrida del {datetime.now():%Y-%m-%d %H:%M}. Regenerable con `python correr.py`. ⛔ Ninguna cifra de blog: todo salió "
        "de aquí. La misma lógica (app/logica.py) en Python y el mismo servidor Go de la serie «¿Go o X?».\n")
    segs = {fw: construir_py(fw) for fw in FWS}
    segs["go"] = construir_go()
    v = R.sh(["docker", "run", "--rm", "--entrypoint", "python", img("fastapi"), "--version"])
    gv = R.sh(["docker", "run", "--rm", "golang:1.27.2", "go", "version"])
    R.w(f"- `{(v.stdout or v.stderr).strip()}` (python:3.13.17-slim, la base del 770) · `{gv.stdout.strip()}` (FROM scratch, "
        f"binario estático, solo net/http). Docker `{R.sh(['docker', 'version', '--format', '{{.Server.Version}}']).stdout.strip()}`.")
    R.w(f"- Servidores con {R.CPUS} CPU y {R.MEM}. Python: 2 procesos (uno por CPU). Go: 1 proceso (GOMAXPROCS lo fija el runtime "
        f"según el límite del contenedor). `hey -c {R.C}` con 4 CPU desde otro contenedor.")
    R.w(f"- /trabajo?n={R.N_TRABAJO}: crea {R.N_TRABAJO} objetos, los mete en un dict/mapa, los ordena y arma un JSON a mano. "
        f"/suma?n={R.N_SUMA}: multiplicar y módulo en un bucle. /suma?n=1: el cálculo es casi cero, queda el coste del servidor.")
    R.w("- Rutas Python síncronas (`def`): FastAPI → threadpool; Litestar `sync_to_thread=True` → threadpool; Robyn `def` (su runtime "
        "en Rust la llama). Granian con `--interface asgi`: la MISMA app que con uvicorn.\n")
    R.w("| imagen | dependencias directas (versiones fijadas en req/<fw>.txt) |")
    R.w("|---|---|")
    for fw in FWS:
        d = ", ".join(x.strip() for x in (AQUI / "req" / f"{fw}.in").read_text(encoding="utf-8").splitlines() if x.strip())
        R.w(f"| {img(fw)} | {d} |")
    R.w(f"| {GO_IMG} | ninguna: librería estándar (net/http) |")
    R.w()
    R.w("## 1 · Construir la imagen desde cero (`docker build --no-cache`; la imagen base ya descargada no cuenta)\n")
    R.w("| imagen | paquetes Python | segundos |")
    R.w("|---|---|---|")
    for fw in FWS:
        n = len([x for x in (AQUI / "req" / f"{fw}.txt").read_text(encoding="utf-8").splitlines() if x.strip()])
        R.w(f"| {fw} | {n} | {segs[fw]:.1f} |")
    R.w(f"| go | 0 | {segs['go']:.1f} |")
    R.w()
    R.w("## 2 · Lo que despliegas\n")
    R.w("| imagen | tamaño en disco |")
    R.w("|---|---|")
    for fw, im in [(fw, img(fw)) for fw in FWS] + [("go", GO_IMG)]:
        s = int(R.sh(["docker", "image", "inspect", "-f", "{{.Size}}", im]).stdout.strip())
        R.w(f"| {fw} | {s / 1048576:.1f} MiB |")
    R.w()

    mismo_resultado()
    if comprobar:
        (AQUI / "SALIDA-comprobacion.md").write_text("\n".join(R.OUT) + "\n", encoding="utf-8")
        print(f"\n[ok] {AQUI / 'SALIDA-comprobacion.md'} (sin carga)")
        return
    X.arranque_y_memoria()
    computo_una()
    X.carga_max("trabajo", R.N_TRABAJO, "4 · Carga con memoria: /trabajo (objetos, dict, orden, JSON), a tope")
    X.carga_max("suma", R.N_SUMA, "5 · Carga de CPU pura: /suma, a tope")
    X.carga_max("suma", 1, "6 · El coste del servidor: /suma?n=1 (casi sin cálculo), a tope")
    X.ritmo_fijo("trabajo", R.N_TRABAJO, 0, "7 · Latencia de una petición sola: /trabajo")
    X.ritmo_fijo("suma", 1, 0, "8 · Latencia de una petición sola: /suma?n=1 (el servidor solo)")
    sin_gil()
    robyn_arranque()
    serie_carga("suma", 1, "n1")
    (AQUI / "SALIDA.md").write_text("\n".join(R.OUT) + "\n", encoding="utf-8")
    print(f"\n[ok] {AQUI / 'SALIDA.md'}")


if __name__ == "__main__":
    main()
