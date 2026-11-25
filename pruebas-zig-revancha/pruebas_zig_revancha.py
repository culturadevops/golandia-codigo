# -*- coding: utf-8 -*-
"""pruebas_zig_revancha.py — la revancha de Zig que pidió un comentario del vídeo 487 (golandia, _I6uG3utnm4):
«sobre los 300 hilos de zig creo que lo estás implementando mal. Con pool de workers obtendrías más rps con menos hilos».

Variantes (mismo contrato HTTP que el servidor de Go de la serie: /suma, /trabajo, /espera):
  · go         — recursos/pruebas-go-java-revancha/go (net/http), el MISMO de 651-654, 699 y 803-808
  · zig-hilos  — recursos/pruebas-zig-revancha/zig con MODO=hilos: una tarea por conexión (Io.Group.concurrent), como el 487
  · zig-pool   — el mismo binario con MODO=pool HILOS=2: dos hilos fijos (uno por CPU) que sacan conexiones de una cola
  · (Evented, io_uring: NO se puede en Zig 0.17.0 → pruebas-zig-revancha/evented/SALIDA-evented.md)

Método (el de la serie): servidores con 2 CPU y 1 GiB, `hey` con 4 CPU desde otro contenedor, la comprobación de que todos
devuelven lo mismo antes de medir. Añadidos de esta revancha:
  · dos regímenes: keep-alive (lo que hace `hey` por defecto) y conexión nueva por petición (`-disable-keepalive`)
  · 16 y 300 conexiones; /suma, /trabajo y /espera?ms=100 (solo con 300)
  · cada medición 3 veces, en orden ROTADO (go→hilos→pool, hilos→pool→go, pool→go→hilos), contenedor nuevo en cada una,
    5 s de calentamiento y 15 s medidos; se publica la mediana y el rango
  · CPU %, memoria e hilos muestreados cada 0,5 s durante la medición (investigacion/muestreo.py) → series/*.json + *.csv
  · control: las mismas corridas con y sin muestreo, para ver que el muestreo no ensucia

⛔ Toma el candado _logs/banco.lock en cada sección y lo suelta al terminar (aunque falle).

    python investigacion/pruebas_zig_revancha.py imagenes|igual|carga|control|pelicula|informe [--rapido]
"""
import json
import os
import re
import shutil
import statistics
import subprocess
import sys
import time
import urllib.request
from contextlib import contextmanager
from datetime import datetime
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
from muestreo import Muestreo  # noqa: E402

REC = Path(__file__).resolve().parent.parent / "recursos"
BASE = REC / "pruebas-zig-revancha"
SERIES = BASE / "series"
CRUDO = BASE / "_salida_cruda.json"
CANDADO = Path(r"<proyecto>\_logs\banco.lock")
RED, SRV, PUERTO = "vzr-red", "vzr-srv", 19300
CPUS, MEM = "2", "1g"
N_SUMA, N_TRABAJO, ESPERA_MS = 200_000, 20_000, 100
SEG_CAL, SEG_MED, REPES = 5, 15, 3
LIM_SRV, LIM_HEY, NANO_ESPERADO = ["--cpus", CPUS], ["--cpus", "4"], "2000000000"  # pruebas_zig_revancha2.py cpuset los cambia
IMG = {"go": "vzr-go:1", "zig": "vzr-zig:1", "hey": "vzr-hey:1"}
VARIANTES = [("go", IMG["go"], {}), ("zig-hilos", IMG["zig"], {"MODO": "hilos"}),
             ("zig-pool", IMG["zig"], {"MODO": "pool", "HILOS": "2"})]
RUTAS = {"suma": f"/suma?n={N_SUMA}", "trabajo": f"/trabajo?n={N_TRABAJO}", "espera": f"/espera?ms={ESPERA_MS}"}
CONFIGS = [(reg, ruta, c) for reg in ("keepalive", "nueva") for ruta, c in
           (("suma", 16), ("suma", 300), ("trabajo", 16), ("trabajo", 300), ("espera", 300))]


def sh(args, **kw):
    return subprocess.run(args, capture_output=True, text=True, encoding="utf-8", errors="replace", **kw)


@contextmanager
def candado(que):
    while True:
        try:
            CANDADO.mkdir()
            break
        except FileExistsError:
            print(f"   (candado ocupado: {(CANDADO / 'dueño.txt').read_text(encoding='utf-8', errors='replace')[:80] if (CANDADO / 'dueño.txt').exists() else '?'}; espero)", flush=True)
            time.sleep(20)
    (CANDADO / "dueño.txt").write_text(f"agente revancha-zig (golandia), {que}, desde {datetime.now():%H:%M:%S}\n", encoding="utf-8")
    try:
        yield
    finally:
        shutil.rmtree(CANDADO, ignore_errors=True)


def cargar():
    return json.loads(CRUDO.read_text(encoding="utf-8")) if CRUDO.exists() else {}


def guardar(d):
    CRUDO.write_text(json.dumps(d, indent=1, ensure_ascii=False), encoding="utf-8")


# ── contenedores ────────────────────────────────────────────────────────────────────────────────────
def arrancar(imagen, env):
    sh(["docker", "rm", "-f", SRV])
    args = ["docker", "run", "-d", "--rm", "--name", SRV, "--network", RED, *LIM_SRV, "--memory", MEM,
            "-p", f"{PUERTO}:8080"]
    for k, v in env.items():
        args += ["-e", f"{k}={v}"]
    r = sh(args + [imagen])
    assert r.returncode == 0, r.stderr
    t0 = time.perf_counter()
    while True:
        try:
            urllib.request.urlopen(f"http://127.0.0.1:{PUERTO}/suma?n=1", timeout=1).read()
            break
        except Exception:
            if time.perf_counter() - t0 > 30:
                raise SystemExit("no arrancó")
            time.sleep(0.01)
    # mismo estado: los límites que de verdad tiene el contenedor
    ins = sh(["docker", "inspect", "-f", "{{.HostConfig.NanoCpus}} {{.HostConfig.Memory}}", SRV]).stdout.split()
    assert ins == [NANO_ESPERADO, str(1024 ** 3)], f"límites distintos: {ins}"


def parar():
    sh(["docker", "rm", "-f", SRV])


def hey(ruta, c, seg, keepalive):
    args = ["docker", "run", "--rm", "--network", RED, *LIM_HEY,
            # conexiones nuevas a miles: más puertos de origen y reutilizar TIME_WAIT en el cliente (igual para todos)
            "--sysctl", "net.ipv4.ip_local_port_range=1024 65535", "--sysctl", "net.ipv4.tcp_tw_reuse=1",
            IMG["hey"], "-z", f"{seg}s", "-c", str(c)]
    if not keepalive:
        args.append("-disable-keepalive")
    t = sh(args + [f"http://{SRV}:8080{RUTAS[ruta]}"]).stdout
    m = re.search(r"Requests/sec:\s+([\d.]+)", t)
    lat = {}
    for q in (50, 99):
        mm = re.search(rf"\s{q}% in ([\d.]+) secs", t)
        lat[q] = float(mm.group(1)) * 1000 if mm else None
    ok = sum(int(n) for cod, n in re.findall(r"\[(\d+)\]\s+(\d+) responses", t) if cod == "200")
    otros = sum(int(n) for cod, n in re.findall(r"\[(\d+)\]\s+(\d+) responses", t) if cod != "200")
    errores = sum(int(n) for n in re.findall(r"^\s+\[(\d+)\]\s+(?:Get|Post)", t, re.M))
    tot = re.search(r"Total:\s+([\d.]+) secs", t)
    return {"rps": float(m.group(1)) if m else 0.0, "p50": lat[50], "p99": lat[99], "ok": ok, "no200": otros,
            "errores": errores, "total_s": float(tot.group(1)) if tot else None, "rps_ok": ok / SEG_MED}


def corrida(nombre, imagen, env, ruta, c, keepalive, muestrear=True, etiqueta=None):
    arrancar(imagen, env)
    hey(ruta, c, SEG_CAL, keepalive)
    res = {}
    if muestrear:
        with Muestreo(SRV) as m:
            res = hey(ruta, c, SEG_MED, keepalive)
        res["muestreo"] = m.resumen()
        if etiqueta:
            m.guardar(SERIES / f"{etiqueta}.json", extra={"variante": nombre, "ruta": ruta, "conexiones": c,
                                                          "keepalive": keepalive, "hey": {k: v for k, v in res.items()
                                                                                          if k != "muestreo"}})
    else:
        res = hey(ruta, c, SEG_MED, keepalive)
    parar()
    return res


def rotar(r):
    return VARIANTES[r % 3:] + VARIANTES[:r % 3]


# ── secciones ───────────────────────────────────────────────────────────────────────────────────────
def imagenes():
    for tag, df, ctx in [(IMG["go"], REC / "pruebas-go-vs-comun" / "Dockerfile.go", REC / "pruebas-go-java-revancha" / "go"),
                         (IMG["zig"], BASE / "zig" / "Dockerfile", BASE / "zig"),
                         (IMG["hey"], REC / "pruebas-go-java-revancha" / "Dockerfile.hey", REC / "pruebas-go-java-revancha")]:
        r = sh(["docker", "build", "-q", "-f", str(df), "-t", tag, str(ctx)])
        assert r.returncode == 0, r.stderr[-2000:]
    sh(["docker", "network", "create", RED])
    d = cargar()
    d["versiones"] = {
        "go": sh(["docker", "run", "--rm", "golang:1.27.2", "go", "version"]).stdout.strip(),
        "zig": sh(["docker", "run", "--rm", "golandia-zig:0.17.0", "zig", "version"]).stdout.strip(),
        "docker": sh(["docker", "version", "--format", "{{.Server.Version}}"]).stdout.strip(),
        "imagen_mib": {k: int(sh(["docker", "image", "inspect", "-f", "{{.Size}}", v]).stdout.strip()) / 1048576
                       for k, v in IMG.items() if k != "hey"},
    }
    guardar(d)
    print(d["versiones"])


def igual():
    d = cargar()
    salidas = {}
    for nombre, imagen, env in VARIANTES:
        arrancar(imagen, env)
        j = json.loads(urllib.request.urlopen(f"http://127.0.0.1:{PUERTO}{RUTAS['trabajo']}").read())
        s = json.loads(urllib.request.urlopen(f"http://127.0.0.1:{PUERTO}{RUTAS['suma']}").read())
        salidas[nombre] = [j["total"], [x["id"] for x in j["primeros"]], s["suma"]]
        log = sh(["docker", "logs", SRV])
        salidas[nombre].append((log.stderr + log.stdout).strip().splitlines()[-1:])
        parar()
    ref = salidas["go"][:3]
    for k, v in salidas.items():
        print(k, v[0], v[1][:5], v[2], "OK" if v[:3] == ref else "⛔ DISTINTO", v[3])
        if v[:3] != ref:
            raise SystemExit("no devuelven lo mismo: no se compara")
    d["igual"] = salidas
    guardar(d)


def carga(solo=None):
    SERIES.mkdir(exist_ok=True)
    d = cargar()
    d.setdefault("carga", {})
    for reg, ruta, c in CONFIGS:
        clave = f"{reg}-{ruta}-c{c}"
        if solo and clave not in solo:
            continue
        print(f"== {clave}", flush=True)
        filas = {n: [] for n, _, _ in VARIANTES}
        with candado(f"carga {clave}"):
          for r in range(REPES):
            for nombre, imagen, env in rotar(r):
                res = corrida(nombre, imagen, env, ruta, c, reg == "keepalive", etiqueta=f"{clave}-{nombre}-r{r + 1}")
                res["orden"] = r
                filas[nombre].append(res)
                mu = res["muestreo"]
                print(f"   r{r + 1} {nombre:10s} {res['rps']:9.0f} pet/s  p50 {res['p50']}  p99 {res['p99']}  ok {res['ok']} "
                      f"err {res['errores']}  cpu {mu.get('cpu_media')}%  mem {mu.get('mem_max_mib')}  hilos {mu.get('hilos_max')}",
                      flush=True)
        d["carga"][clave] = filas
        guardar(d)


def control():
    """Las mismas corridas con y sin muestreo, intercaladas: ¿el auxiliar que lee el cgroup cambia las peticiones/s?"""
    d = cargar()
    d.setdefault("control", {})
    for reg, ruta, c in (("keepalive", "suma", 16), ("nueva", "trabajo", 16)):
        clave = f"{reg}-{ruta}-c{c}"
        filas = {n: {"con": [], "sin": []} for n, _, _ in VARIANTES}
        for r in range(REPES):
            for nombre, imagen, env in rotar(r):
                orden = (True, False) if r % 2 == 0 else (False, True)
                for mu in orden:
                    res = corrida(nombre, imagen, env, ruta, c, reg == "keepalive", muestrear=mu)
                    filas[nombre]["con" if mu else "sin"].append(res["rps"])
                    print(f"   {clave} r{r + 1} {nombre:10s} {'con' if mu else 'sin'} muestreo {res['rps']:9.0f} pet/s", flush=True)
        d["control"][clave] = filas
        guardar(d)


def pelicula():
    """Una corrida por variante para las gráficas: 3 s en reposo, 12 s de carga con 300 conexiones, 5 s después."""
    SERIES.mkdir(exist_ok=True)
    d = cargar()
    d.setdefault("pelicula", {})
    for reg, ruta in (("nueva", "trabajo"), ("keepalive", "suma"), ("keepalive", "espera")):
        for nombre, imagen, env in VARIANTES:
            arrancar(imagen, env)
            with Muestreo(SRV, puerto=8080) as m:
                time.sleep(3)
                res = hey(ruta, 300, 12, reg == "keepalive")
                time.sleep(5)
            parar()
            et = f"pelicula-{reg}-{ruta}-c300-{nombre}"
            m.guardar(SERIES / f"{et}.json", extra={"variante": nombre, "ruta": ruta, "conexiones": 300,
                                                     "keepalive": reg == "keepalive", "hey": res,
                                                     "nota": "3 s en reposo, 12 s de carga, 5 s después"})
            d["pelicula"][et] = {"hey": res, "resumen": m.resumen()}
            print(f"   {et}: {res['rps']:.0f} pet/s · {m.resumen()}", flush=True)
    guardar(d)


def diagnostico():
    """¿Por qué zig-hilos llega a ~1.100 hilos con 300 conexiones? La corrida de la sección de carga (5 s de calentamiento + 15 s
    medidos, keep-alive, /suma, 300 conexiones) con el muestreo encendido DESDE el principio y contando conexiones establecidas."""
    SERIES.mkdir(exist_ok=True)
    d = cargar()
    d.setdefault("diagnostico", {})
    for nombre, imagen, env in VARIANTES[:2]:
        arrancar(imagen, env)
        with Muestreo(SRV, puerto=8080) as m:
            time.sleep(2)
            cal = hey("suma", 300, SEG_CAL, True)
            time.sleep(2)
            res = hey("suma", 300, SEG_MED, True)
            time.sleep(3)
        parar()
        et = f"diagnostico-keepalive-suma-c300-{nombre}"
        m.guardar(SERIES / f"{et}.json", extra={"variante": nombre, "nota": "2 s reposo, 5 s calentamiento, 2 s pausa, 15 s medidos, 3 s después",
                                                 "hey_calentamiento": cal, "hey": res})
        d["diagnostico"][et] = {"hey": res, "resumen": m.resumen()}
        print(f"   {et}: {res['rps']:.0f} pet/s · {m.resumen()}", flush=True)
    guardar(d)


# ── informe ─────────────────────────────────────────────────────────────────────────────────────────
def med(v):
    v = [x for x in v if x is not None]
    return statistics.median(v) if v else None


def rango(v, f="{:.0f}"):
    v = [x for x in v if x is not None]
    return f"{f.format(min(v))}–{f.format(max(v))}" if v else "—"


def informe():
    d = cargar()
    o = []
    w = o.append
    v = d["versiones"]
    w("# SALIDA — la revancha de Zig: hilo por conexión contra pool de hilos, y Go (golandia)\n")
    w(f"> Corrida del {datetime.now():%Y-%m-%d}, con el candado `_logs/banco.lock` tomado en cada sección. Regenerable con "
      "`python investigacion/pruebas_zig_revancha.py imagenes igual carga control pelicula informe`. ⛔ Ninguna cifra de blog.\n")
    w(f"- go: `{v['go']}` (net/http) · imagen {v['imagen_mib']['go']:.1f} MiB")
    w(f"- zig: `{v['zig']}` (std.Io Threaded, solo librería estándar) · imagen {v['imagen_mib']['zig']:.1f} MiB")
    w(f"- Docker `{v['docker']}`. Servidores con {CPUS} CPU y {MEM}; `hey` con 4 CPU desde otro contenedor. Contenedor nuevo en "
      f"cada corrida, {SEG_CAL} s de calentamiento y {SEG_MED} s medidos, {REPES} corridas en orden rotado. Se da la MEDIANA y, "
      "entre corchetes, el rango (mínimo–máximo) de las 3.")
    w("- zig-hilos: una tarea por conexión (`Io.Group.concurrent`, como el 487). zig-pool: 2 hilos fijos + cola (`Io.Queue`).")
    w("- «peticiones/s» = el `Requests/sec` de hey. ⚠️ En zig-pool con keep-alive ese número no compara nada: solo N clientes reciben respuesta (ver «errores») y hey espera 20 s a los demás, así que divide entre 20 s; lo que importa ahí son los clientes sin respuesta. p50/p99: solo de las respuestas que llegaron.")
    w("- Los dos Zig con backlog 4096 (el de Go; el de Zig por defecto es 128) y keep-alive.")
    w("- CPU/memoria/hilos: `investigacion/muestreo.py`, cada 0,5 s, del cgroup del contenedor (100 % = un núcleo; techo 200 %; "
      "el reloj del auxiliar tiene 10 ms de resolución, así que un punto suelto puede pasar de 200 %). «Frenadas»: `nr_throttled` "
      "de cpu.stat en los 15 s medidos: cuántos periodos de 100 ms el límite de 2 CPU detuvo al contenedor.\n")
    w("## 0 · ¿Hacen lo mismo?\n")
    for k, s in d["igual"].items():
        w(f"- {k}: total {s[0]}, primeros ids {s[1][:5]}…, suma {s[2]} · `{(s[3] or [''])[0]}`")
    w("")
    n = 1
    for reg, titulo in (("keepalive", "keep-alive (lo que hace hey por defecto: cada cliente reutiliza su conexión)"),
                        ("nueva", "conexión nueva por petición (`hey -disable-keepalive`)")):
        w(f"## {n} · Régimen {titulo}\n")
        n += 1
        for ruta in ("suma", "trabajo", "espera"):
            for c in (16, 300):
                clave = f"{reg}-{ruta}-c{c}"
                if clave not in d.get("carga", {}):
                    continue
                w(f"### {RUTAS[ruta]} · {c} conexiones\n")
                w("| variante | peticiones/s [rango] | p50 | p99 | respuestas 200 | errores | CPU media | memoria máx. | hilos máx. | frenadas por el límite de CPU |")
                w("|---|---|---|---|---|---|---|---|---|---|")
                for nombre, filas in d["carga"][clave].items():
                    mu = [f["muestreo"] for f in filas]
                    w(f"| {nombre} | **{med([f['rps'] for f in filas]):.0f}** [{rango([f['rps'] for f in filas])}] | "
                      f"{_ms(med([f['p50'] for f in filas]))} | {_ms(med([f['p99'] for f in filas]))} | "
                      f"{med([f['ok'] for f in filas]):.0f} | {med([f['errores'] for f in filas]):.0f} | "
                      f"{med([m.get('cpu_media') for m in mu]):.0f} % | {med([m.get('mem_max_mib') for m in mu]):.1f} MiB | "
                      f"{med([m.get('hilos_max') for m in mu]):.0f} | {med([m.get('frenadas') for m in mu]):.0f} |")
                w("")
    if d.get("control"):
        w(f"## {n} · ¿El muestreo ensucia? (mismas corridas con y sin el auxiliar, intercaladas)\n")
        w("| configuración | variante | con muestreo (mediana) | sin muestreo (mediana) | diferencia |")
        w("|---|---|---|---|---|")
        for clave, filas in d["control"].items():
            for nombre, cs in filas.items():
                a, b = med(cs["con"]), med(cs["sin"])
                w(f"| {clave} | {nombre} | {a:.0f} {cs['con']and [round(x) for x in cs['con']]} | "
                  f"{b:.0f} {[round(x) for x in cs['sin']]} | {(a - b) / b * 100:+.1f} % |")
        w("")
        n += 1
    if d.get("pelicula"):
        w(f"## {n} · Las series en el tiempo (para las gráficas)\n")
        w("3 s en reposo, 12 s con 300 conexiones, 5 s después; un punto cada 0,5 s. Ficheros: `series/pelicula-*.json` y `.csv`.\n")
        w("| serie | peticiones/s | CPU media | CPU máx. | memoria máx. | hilos máx. | conexiones máx. |")
        w("|---|---|---|---|---|---|---|")
        for et, x in d["pelicula"].items():
            r = x["resumen"]
            w(f"| {et} | {x['hey']['rps']:.0f} | {r['cpu_media']} % | {r['cpu_max']} % | {r['mem_max_mib']} MiB | {r['hilos_max']} | {r.get('conexiones_max', '—')} |")
        w("")
    if d.get("diagnostico"):
        w(f"## {n + 1} · Diagnóstico: hilos contra conexiones abiertas (keep-alive, /suma, 300 conexiones)\n")
        w("La corrida de la sección de carga con el muestreo encendido desde el reposo: 5 s de calentamiento, 2 s de pausa, 15 s "
          "medidos. «Conexiones» = TCP establecidas en el puerto 8080 del contenedor. Series: `series/diagnostico-*.json`.\n")
        w("| serie | peticiones/s (15 s) | hilos máx. | conexiones máx. | memoria máx. |")
        w("|---|---|---|---|---|")
        for et, x in d["diagnostico"].items():
            r = x["resumen"]
            w(f"| {et} | {x['hey']['rps']:.0f} | {r['hilos_max']} | {r.get('conexiones_max', '—')} | {r['mem_max_mib']} MiB |")
        w("")
    (BASE / "SALIDA.md").write_text("\n".join(o) + "\n", encoding="utf-8")
    print(f"[ok] {BASE / 'SALIDA.md'}")


def _ms(x):
    return "—" if x is None else f"{x:.1f} ms"


if __name__ == "__main__":
    os.environ["MSYS_NO_PATHCONV"] = "1"
    sys.stdout.reconfigure(encoding="utf-8")
    sec = sys.argv[1]
    if sec == "informe":
        informe()
    elif sec == "imagenes":
        imagenes()
    elif sec == "carga":
        carga(sys.argv[2:] or None)  # toma el candado en cada configuración
    else:
        with candado(sec):
            {"igual": igual, "control": control, "pelicula": pelicula, "diagnostico": diagnostico}[sec]()
