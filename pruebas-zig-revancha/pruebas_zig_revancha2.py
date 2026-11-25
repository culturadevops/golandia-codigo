# -*- coding: utf-8 -*-
"""pruebas_zig_revancha2.py — segunda parte de la revancha de Zig: lo que salió raro en la primera se comprueba antes de contarlo.

  igual     zig-eventos devuelve lo mismo que Go
  causa     keep-alive: el pool de N hilos atiende N clientes y el resto no. Prueba: (HILOS, conexiones) = (2,2) (2,3) (2,16) (16,16)
  pooln     /espera?ms=100 sin keep-alive da 20 pet/s con 2 hilos. Prueba: si la causa es dormir el hilo, pet/s = N / 0,1 s
  diag      ¿por qué zig-hilos pasa de 300 hilos con 300 conexiones? binario con contadores (diag/main.zig) + hilos y conexiones
  cpuset    los p99 de ~100 ms de zig-hilos, ¿son el límite de CPU (--cpus, periodos de 100 ms)? Lo mismo con --cpuset-cpus: dos
            núcleos enteros, sin cuota. El cliente, en otros núcleos
  pelicula  las series de las gráficas, las cuatro variantes (tras el corte de luz del 2026-11-05 se repiten todas)
  carga     zig-eventos con las 10 configuraciones, intercalado con Go, que hace de ancla contra la primera tanda

⛔ Toma el candado _logs/banco.lock en cada sección y lo suelta al terminar.

    python investigacion/pruebas_zig_revancha2.py igual causa pooln diag cpuset pelicula carga informe
"""
import json
import sys
import time
import urllib.request
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
import pruebas_zig_revancha as R  # noqa: E402
from muestreo import Muestreo  # noqa: E402

GO = R.VARIANTES[0]
HILOS, POOL = R.VARIANTES[1], R.VARIANTES[2]
EV = ("zig-eventos", "vzr-zig-ev:2", {"HILOS": "2"})
DIAG = ("zig-hilos-diag", "vzr-zig-diag:1", {"MODO": "hilos"})
TODAS = [GO, HILOS, POOL, EV]


def guardar(clave, valor):
    d = R.cargar()
    d[clave] = valor
    R.guardar(d)


def igual():
    salidas = {}
    for nombre, imagen, env in (GO, EV):
        R.arrancar(imagen, env)
        j = json.loads(urllib.request.urlopen(f"http://127.0.0.1:{R.PUERTO}{R.RUTAS['trabajo']}").read())
        s = json.loads(urllib.request.urlopen(f"http://127.0.0.1:{R.PUERTO}{R.RUTAS['suma']}").read())
        salidas[nombre] = [j["total"], [x["id"] for x in j["primeros"]], s["suma"]]
        R.parar()
    print(salidas["go"][0], salidas["go"][2], salidas["go"] == salidas["zig-eventos"], flush=True)
    if salidas["go"] != salidas["zig-eventos"]:
        raise SystemExit("zig-eventos no devuelve lo mismo")
    guardar("igual2", salidas)


def causa():
    filas = {}
    for n, c in ((2, 2), (2, 3), (2, 16), (16, 16)):
        k = f"hilos{n}-c{c}"
        filas[k] = []
        for r in range(R.REPES):
            res = R.corrida(k, POOL[1], {"MODO": "pool", "HILOS": str(n)}, "suma", c, True)
            filas[k].append(res)
            print(f"   {k} r{r + 1}: ok {res['ok']} ({res['ok'] / R.SEG_MED:.0f} pet/s) errores {res['errores']} "
                  f"hilos {res['muestreo'].get('hilos_max')}", flush=True)
    guardar("causa_keepalive", filas)


def pooln():
    filas = {}
    for n in (2, 8, 32, 300):
        filas[str(n)] = []
        for r in range(R.REPES):
            res = R.corrida(f"pool{n}", POOL[1], {"MODO": "pool", "HILOS": str(n)}, "espera", 300, False)
            filas[str(n)].append(res)
            print(f"   HILOS={n} r{r + 1}: {res['ok'] / R.SEG_MED:.0f} pet/s (teoría {n / 0.1:.0f}) p50 {res['p50']} "
                  f"mem {res['muestreo'].get('mem_max_mib')}", flush=True)
    guardar("pool_n", filas)


def diag():
    filas = {}
    for reg, ruta in (("keepalive", "suma"), ("nueva", "trabajo"), ("keepalive", "espera")):
        for r in range(2):
            nombre, imagen, env = DIAG
            R.arrancar(imagen, env)
            with Muestreo(R.SRV, puerto=8080) as m:
                time.sleep(2)
                R.hey(ruta, 300, R.SEG_CAL, reg == "keepalive")
                time.sleep(2)
                res = R.hey(ruta, 300, R.SEG_MED, reg == "keepalive")
                time.sleep(3)
            log = R.sh(["docker", "logs", R.SRV])
            lineas = [x for x in (log.stderr + log.stdout).splitlines() if x.startswith("DIAG")]
            R.parar()
            et = f"diag-{reg}-{ruta}-c300-r{r + 1}"
            m.guardar(R.SERIES / f"{et}.json", extra={"hey": res, "diag": lineas})
            filas[et] = {"hey": res, "resumen": m.resumen(), "ultima": lineas[-1] if lineas else None}
            print(f"   {et}: {m.resumen()} · {filas[et]['ultima']}", flush=True)
    guardar("diagnostico2", filas)


def cpuset():
    R.LIM_SRV, R.LIM_HEY, R.NANO_ESPERADO = ["--cpuset-cpus", "2,4"], ["--cpuset-cpus", "8-11"], "0"
    filas = {}
    for reg, ruta, c in (("keepalive", "suma", 16), ("nueva", "trabajo", 16), ("nueva", "suma", 300)):
        k = f"{reg}-{ruta}-c{c}"
        filas[k] = {n: [] for n, _, _ in TODAS if n != "zig-pool" or reg == "nueva"}
        for r in range(R.REPES):
            orden = [v for v in TODAS if v[0] in filas[k]]
            orden = orden[r % len(orden):] + orden[:r % len(orden)]
            for nombre, imagen, env in orden:
                res = R.corrida(nombre, imagen, env, ruta, c, reg == "keepalive")
                filas[k][nombre].append(res)
                print(f"   {k} r{r + 1} {nombre:12s} {res['ok'] / R.SEG_MED:7.0f} pet/s p50 {res['p50']} p99 {res['p99']} hilos "
                      f"{res['muestreo'].get('hilos_max')} frenadas {res['muestreo'].get('frenadas')}", flush=True)
    guardar("cpuset", filas)
    R.LIM_SRV, R.LIM_HEY, R.NANO_ESPERADO = ["--cpus", R.CPUS], ["--cpus", "4"], "2000000000"


def pelicula():
    R.VARIANTES[:] = TODAS
    R.pelicula()


def carga():
    filas_todas = R.cargar().get("carga2", {})
    for reg, ruta, c in R.CONFIGS:
        k = f"{reg}-{ruta}-c{c}"
        if k in filas_todas:
            continue
        filas = {"go": [], "zig-eventos": []}
        with R.candado(f"carga2 {k}"):
            for r in range(R.REPES):
                for nombre, imagen, env in ((GO, EV) if r % 2 == 0 else (EV, GO)):
                    res = R.corrida(nombre, imagen, env, ruta, c, reg == "keepalive", etiqueta=f"c2-{k}-{nombre}-r{r + 1}")
                    filas[nombre].append(res)
                    mu = res["muestreo"]
                    print(f"   {k} r{r + 1} {nombre:12s} {res['ok'] / R.SEG_MED:7.0f} pet/s p50 {res['p50']} p99 {res['p99']} "
                          f"err {res['errores']} cpu {mu.get('cpu_media')} mem {mu.get('mem_max_mib')} hilos {mu.get('hilos_max')}",
                          flush=True)
        filas_todas[k] = filas
        guardar("carga2", filas_todas)


def final():
    """La tabla del vídeo: las cuatro variantes en la MISMA sesión, orden rotado, 3 corridas, con el arena por hilo de zig-eventos.
    zig-pool no corre con keep-alive: ahí no compara nada (sección B)."""
    tabla = R.cargar().get("final", {})
    for reg, ruta, c in R.CONFIGS:
        k = f"{reg}-{ruta}-c{c}"
        if k in tabla:
            continue
        vs = [v for v in TODAS if v[0] != "zig-pool" or reg == "nueva"]
        filas = {v[0]: [] for v in vs}
        with R.candado(f"final {k}"):
            for r in range(R.REPES):
                for nombre, imagen, env in vs[r % len(vs):] + vs[:r % len(vs)]:
                    res = R.corrida(nombre, imagen, env, ruta, c, reg == "keepalive", etiqueta=f"final-{k}-{nombre}-r{r + 1}")
                    filas[nombre].append(res)
                    mu = res["muestreo"]
                    print(f"   {k} r{r + 1} {nombre:12s} {res['rps']:7.0f} pet/s p50 {res['p50']} p99 {res['p99']} err {res['errores']} "
                          f"cpu {mu.get('cpu_media')} mem {mu.get('mem_max_mib')} hilos {mu.get('hilos_max')}", flush=True)
        tabla[k] = filas
        guardar("final", tabla)


def _fila(filas):
    mu = [f["muestreo"] for f in filas]
    rps = [f["rps"] for f in filas]
    return (f"**{R.med(rps):.0f}** [{R.rango(rps)}] | {R._ms(R.med([f['p50'] for f in filas]))} | "
            f"{R._ms(R.med([f['p99'] for f in filas]))} | {R.med([f['errores'] for f in filas]):.0f} | "
            f"{R.med([m.get('cpu_media') for m in mu]):.0f} % | {R.med([m.get('mem_max_mib') for m in mu]):.1f} MiB | "
            f"{R.med([m.get('hilos_max') for m in mu]):.0f} | {R.med([m.get('frenadas') for m in mu]):.0f}")


def informe():
    R.informe()
    d = R.cargar()
    o = []
    w = o.append
    w("# Segunda parte: lo raro, comprobado (2026-11-05, tras reiniciar la máquina)\n")
    w("Binarios de la primera parte sin tocar. `zig-eventos` (eventos/eventos.zig): 2 hilos, cada uno con su epoll y su socket "
      "SO_REUSEPORT, escrito a mano sobre std.os.linux porque std.Io.Evented no compila en 0.17.0. Mismo /suma y /trabajo.\n")
    h = "| variante | peticiones/s [rango] | p50 | p99 | errores | CPU media | memoria máx. | hilos máx. | frenadas |"
    s = "|---|---|---|---|---|---|---|---|---|"
    if d.get("final"):
        w("## F · LA TABLA DEL VÍDEO: las cuatro variantes en la misma sesión (zig-eventos con arena por hilo)\n")
        for k, filas in d["final"].items():
            w(f"### {k}\n")
            w(h)
            w(s)
            for n, f in filas.items():
                w(f"| {n} | {_fila(f)} |")
            w("")
    if d.get("carga2"):
        w("## A · zig-eventos contra Go (intercalados; Go es el ancla contra la primera parte)\n")
        for k, filas in d["carga2"].items():
            antes = R.med([f["rps"] for f in d["carga"][k]["go"]])
            ahora = R.med([f["rps"] for f in filas["go"]])
            w(f"### {k} (Go en la primera parte: {antes:.0f} pet/s; ahora {ahora:.0f}: {(ahora - antes) / antes * 100:+.1f} %)\n")
            w(h)
            w(s)
            for n, f in filas.items():
                w(f"| {n} | {_fila(f)} |")
            w("")
    if d.get("causa_keepalive"):
        w("## B · Causa: keep-alive y pool bloqueante (/suma, 15 s medidos, variando hilos y conexiones)\n")
        w("| hilos del pool · conexiones | respuestas por corrida | peticiones/s | errores (clientes sin respuesta) |")
        w("|---|---|---|---|")
        for k, f in d["causa_keepalive"].items():
            w(f"| {k} | {[x['ok'] for x in f]} | {R.med([x['ok'] / R.SEG_MED for x in f]):.0f} (respuestas ÷ 15 s) | {[x['errores'] for x in f]} |")
        w("")
    if d.get("pool_n"):
        w("## C · Causa: /espera?ms=100 sin keep-alive, 300 conexiones, pool de N hilos que duermen\n")
        w("| hilos | peticiones/s [rango] | teoría N ÷ 0,1 s | p50 | memoria máx. |")
        w("|---|---|---|---|---|")
        for n, f in d["pool_n"].items():
            rps = [x["rps"] for x in f]
            w(f"| {n} | **{R.med(rps):.0f}** [{R.rango(rps)}] | {int(n) / 0.1:.0f} | {R._ms(R.med([x['p50'] for x in f]))} | "
              f"{R.med([x['muestreo'].get('mem_max_mib') for x in f]):.1f} MiB |")
        w("")
    if d.get("diagnostico2"):
        w("## D · Diagnóstico: hilos contra conexiones contra tareas (binario diag/main.zig, solo para esto)\n")
        w("«vivas_max» = aceptadas menos terminadas (cuenta las que esperan hilo); «corriendo_max» = dentro de atender().\n")
        w("| corrida | hilos máx. | conexiones máx. | último DIAG |")
        w("|---|---|---|---|")
        for et, x in d["diagnostico2"].items():
            r = x["resumen"]
            w(f"| {et} | {r['hilos_max']} | {r.get('conexiones_max')} | `{x['ultima']}` |")
        w("")
    if d.get("cpuset"):
        w("## E · ¿Es el límite de CPU? Lo mismo con `--cpuset-cpus 2,4` (dos núcleos enteros, sin cuota; hey en 8-11)\n")
        for k, filas in d["cpuset"].items():
            w(f"### {k}\n")
            w(h)
            w(s)
            for n, f in filas.items():
                w(f"| {n} | {_fila(f)} |")
            w("")
    w("## G · Lo que hay detrás del límite de CPU (ejecutado 2026-11-05)\n")
    w("```")
    w("$ docker run --rm --cpus 2 golang:1.27.2 sh -c 'go run main.go; cat /sys/fs/cgroup/cpu.max'")
    w("GOMAXPROCS 2 NumCPU 24")
    w("cpu.max: 200000 100000")
    w("$ docker run --rm golang:1.27.2 go run main.go")
    w("GOMAXPROCS 24 NumCPU 24")
    w("```\n")
    w("## H · Notas de lectura\n")
    w("- Para comparar variantes, usa la sección F: las cuatro en la misma sesión. Entre sesiones Go se movió un 5 % (la primera parte, "
      "antes de un corte de luz, salió más alta) y zig-hilos con `--cpus 2` mucho más: /suma keep-alive 16 conexiones, "
      "7.149 en la primera parte y 11.695 en la F. Dentro de una sesión su rango es estrecho; entre sesiones no: es la variante más "
      "sensible a la cuota.")
    w("- zig-eventos de la carga A y de la sección E retenía un arena por conexión (571 MiB en keep-alive /trabajo con 300). Fallo mío, no "
      "del método: la F y las películas usan un arena por hilo (13,4 MiB).")
    w("- El control de la sección 3: el 4435 de zig-pool no es el muestreo. Con keep-alive el reparto depende de qué conexiones toman los "
      "dos hilos; si uno se queda con una conexión ociosa, atiende la mitad.")
    w("- «Frenadas» no explica sola la latencia: zig-pool sin keep-alive también marca ~149 y su p99 es 1,9 ms. La prueba de causa es la "
      "sección E (sin cuota), no el contador.")
    w("- Los servidores Zig entienden lo justo de HTTP; net/http de Go es completo. Parte de la ventaja viene de ahí (sin aislar).")
    w("")
    with open(R.BASE / "SALIDA.md", "a", encoding="utf-8") as f:
        f.write("\n" + "\n".join(o) + "\n")
    print("[ok] SALIDA.md con segunda parte")


if __name__ == "__main__":
    sys.stdout.reconfigure(encoding="utf-8")
    for sec in sys.argv[1:]:
        print(f"== {sec}", flush=True)
        if sec == "informe":
            informe()
        elif sec == "carga":
            carga()
        elif sec == "final":
            final()
        else:
            with R.candado(sec):
                {"igual": igual, "causa": causa, "pooln": pooln, "diag": diag, "cpuset": cpuset, "pelicula": pelicula}[sec]()
