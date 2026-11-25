# -*- coding: utf-8 -*-
"""pruebas_go_java.py — mide Go contra Java y guarda la salida LITERAL en
`recursos/pruebas-go-java/SALIDA.md`.

Vídeo 148 · `go-vs-java` · la decisión: aprendiste Java, ¿te cambias a Go?

## Por qué se mide así y no de otra forma

⛔ **Cero cifras de blogs.** El canon del canal dice que lo que sale en pantalla se ejecuta
antes. Aquí se ejecuta el MISMO servidor HTTP en los dos lenguajes, cada uno con SU librería
estándar (net/http contra com.sun.net.httpserver), cero frameworks, cada uno en su imagen
oficial (`golang:1.26` para compilar, `eclipse-temurin:21` para compilar y ejecutar Java).

⚠️ **Java del 2026, no del 2015.** El servidor de Java usa el executor de hilos virtuales de
Java 21 (Project Loom). Comparar Go contra el Java de hace diez años sería ganar con trampa.

## Las seis mediciones

    1. compilar            → dentro de la imagen oficial de cada uno, mediana de 3
    2. lo que despliegas   → binario+scratch contra jar+JRE: artefacto e imagen, medidos
    3. arranque            → contenedor nuevo → primera respuesta 200, mediana de 8
                             + el «listo» que declara el propio proceso (LISTO_EN_MS)
    4. memoria en reposo   → docker stats con el servidor recién levantado
    5. templado y el JIT   → 2.000 peticiones de CPU: p50/p99, y frío contra templado
                             usando el reloj DEL PROPIO PROCESO (campo ns del JSON)
    6. concurrencia        → 300 peticiones a la vez esperando 100 ms + memoria bajo carga

    python investigacion/pruebas_go_java.py
"""
import concurrent.futures
import io
import json
import statistics
import subprocess
import sys
import time
import urllib.error
import urllib.request
from pathlib import Path

sys.stdout = io.TextIOWrapper(sys.stdout.buffer, encoding="utf-8")

BASE = Path(__file__).resolve().parent.parent / "recursos" / "pruebas-go-java"

N_CPU = 200_000          # el mismo trabajo que en el vídeo de Lambda (121)
REPES_COMPILA = 3
REPES_FRIO = 8
INVOCACIONES = 2_000
CONCURRENTES = 300
IO_MS = 100

LENGUAJES = [
    # clave    imagen de despliegue      puerto
    ("go",   "golandia-govsjava-go:1",   9201),
    ("java", "golandia-govsjava-java:1", 9202),
]

SALIDA = {}


def sh(args, **kw):
    return subprocess.run(args, capture_output=True, text=True,
                          encoding="utf-8", errors="replace", **kw)


def mib(n):
    return f"{n/1048576:.2f} MiB"


def p(v, q):
    v = sorted(v)
    i = max(0, min(len(v) - 1, round(q * (len(v) - 1))))
    return v[i]


# ── 1 · compilar, dentro de la imagen oficial ────────────────────────────
def medir_compilacion():
    print("1 · compilar (dentro de la imagen oficial, el reloj del contenedor)")
    filas = {}
    ordenes = {
        "go": ["docker", "run", "--rm", "-v", f"{BASE / 'go'}:/src", "-w", "/src",
               "-e", "CGO_ENABLED=0", "golang:1.26", "sh", "-c",
               "t0=$(date +%s%N); go build -o /tmp/servidor .; t1=$(date +%s%N); "
               "echo COMPILA_MS $(( (t1 - t0) / 1000000 ))"],
        "java": ["docker", "run", "--rm", "-v", f"{BASE / 'java'}:/src", "-w", "/src",
                 "eclipse-temurin:21-jdk", "sh", "-c",
                 "t0=$(date +%s%N); javac -d /tmp/classes Servidor.java; "
                 "t1=$(date +%s%N); echo COMPILA_MS $(( (t1 - t0) / 1000000 ))"],
    }
    for lang, orden in ordenes.items():
        tiempos = []
        for _ in range(REPES_COMPILA):
            r = sh(orden)
            if r.returncode != 0:
                raise SystemExit(f"[!] no compila {lang}:\n{r.stderr}")
            for linea in r.stdout.splitlines():
                if linea.startswith("COMPILA_MS"):
                    tiempos.append(int(linea.split()[1]))
        filas[lang] = tiempos
        print(f"    · {lang:5s} {tiempos} ms → mediana {statistics.median(tiempos):.0f} ms")
    SALIDA["compilacion"] = filas


# ── 2 · lo que despliegas ────────────────────────────────────────────────
def construir_y_medir_artefactos():
    print("2 · artefactos e imágenes de despliegue")
    (BASE / "go" / "_bin").mkdir(exist_ok=True)
    (BASE / "java" / "_bin").mkdir(exist_ok=True)

    # El binario de Go, estático y optimizado como se despliega de verdad.
    r = sh(["docker", "run", "--rm", "-v", f"{BASE / 'go'}:/src", "-w", "/src",
            "-e", "CGO_ENABLED=0", "golang:1.26", "sh", "-c",
            "go build -trimpath -ldflags='-s -w' -o _bin/servidor . && chmod +x _bin/servidor"])
    if r.returncode != 0:
        raise SystemExit(f"[!] build go: {r.stderr}")
    # También el plano, solo para el dato del tamaño.
    r = sh(["docker", "run", "--rm", "-v", f"{BASE / 'go'}:/src", "-w", "/src",
            "-e", "CGO_ENABLED=0", "golang:1.26", "sh", "-c",
            "go build -o _bin/servidor-plano ."])
    if r.returncode != 0:
        raise SystemExit(f"[!] build go plano: {r.stderr}")

    # El jar de Java.
    r = sh(["docker", "run", "--rm", "-v", f"{BASE / 'java'}:/src", "-w", "/src",
            "eclipse-temurin:21-jdk", "sh", "-c",
            "javac -d _bin/classes Servidor.java && "
            "jar --create --file _bin/servidor.jar --main-class Servidor -C _bin/classes ."])
    if r.returncode != 0:
        raise SystemExit(f"[!] build java: {r.stderr}")

    artefactos = {
        "go binario (-s -w)": (BASE / "go" / "_bin" / "servidor").stat().st_size,
        "go binario plano": (BASE / "go" / "_bin" / "servidor-plano").stat().st_size,
        "java servidor.jar": (BASE / "java" / "_bin" / "servidor.jar").stat().st_size,
    }
    for k, v in artefactos.items():
        print(f"    · {k:22s} {v:>10,} B  ({mib(v)})".replace(",", "."))

    # Las imágenes de despliegue.
    for lang, imagen, _ in LENGUAJES:
        r = sh(["docker", "build", "-t", imagen, str(BASE / lang)])
        if r.returncode != 0:
            raise SystemExit(f"[!] docker build {lang}:\n{r.stderr}")

    imagenes = {}
    en_disco = {}
    r = sh(["docker", "images", "--format", "{{.Repository}}:{{.Tag}}\t{{.Size}}"])
    for linea in (r.stdout or "").splitlines():
        if "\t" in linea:
            ref, tam = linea.split("\t", 1)
            en_disco[ref] = tam.strip()
    for ref in ["golandia-govsjava-go:1", "golandia-govsjava-java:1",
                "eclipse-temurin:21-jre", "eclipse-temurin:21-jdk", "golang:1.26"]:
        rr = sh(["docker", "image", "inspect", ref, "--format", "{{.Size}}"])
        if rr.returncode == 0:
            imagenes[ref] = {"descarga": int(rr.stdout.strip()),
                             "disco": en_disco.get(ref, "?")}
            print(f"    · {ref:28s} descarga {mib(imagenes[ref]['descarga']):>11s}  "
                  f"disco {imagenes[ref]['disco']}")
    SALIDA["artefactos"] = {k: v for k, v in artefactos.items()}
    SALIDA["imagenes"] = imagenes


# ── utilidades de contenedor ─────────────────────────────────────────────
def parar(nombre):
    sh(["docker", "rm", "-f", nombre])


def esperar_ok(puerto, timeout=60):
    fin = time.perf_counter() + timeout
    while time.perf_counter() < fin:
        try:
            with urllib.request.urlopen(
                    f"http://127.0.0.1:{puerto}/suma?n=1", timeout=2) as r:
                if r.status == 200:
                    r.read()
                    return time.perf_counter()
        except (urllib.error.URLError, ConnectionError, OSError):
            time.sleep(0.002)
    raise SystemExit(f"[!] el puerto {puerto} nunca contestó")


def listo_interno(nombre):
    for _ in range(50):
        r = sh(["docker", "logs", nombre])
        for linea in (r.stderr or "").splitlines() + (r.stdout or "").splitlines():
            if linea.startswith("LISTO_EN_MS"):
                return float(linea.split()[1])
        time.sleep(0.05)
    return None


def memoria_mib(nombre):
    r = sh(["docker", "stats", "--no-stream", "--format", "{{.MemUsage}}", nombre])
    bruto = (r.stdout or "").strip().split("/")[0].strip()   # «12.34MiB»
    if bruto.endswith("GiB"):
        return float(bruto[:-3]) * 1024
    if bruto.endswith("MiB"):
        return float(bruto[:-3])
    if bruto.endswith("KiB"):
        return float(bruto[:-3]) / 1024
    return None


# ── 3 · arranque ─────────────────────────────────────────────────────────
def medir_arranque():
    print("3 · arranque: contenedor nuevo → primera respuesta")
    filas = {}
    for lang, imagen, puerto in LENGUAJES:
        externos, internos = [], []
        for i in range(REPES_FRIO):
            nombre = f"govsjava-{lang}-frio-{i}"
            parar(nombre)
            t0 = time.perf_counter()
            r = sh(["docker", "run", "-d", "--name", nombre,
                    "-p", f"{puerto}:8080", imagen])
            if r.returncode != 0:
                raise SystemExit(f"[!] docker run {lang}: {r.stderr}")
            t1 = esperar_ok(puerto)
            externos.append((t1 - t0) * 1000)
            li = listo_interno(nombre)
            if li is not None:
                internos.append(li)
            parar(nombre)
        filas[lang] = {"externo_ms": externos, "interno_ms": internos}
        print(f"    · {lang:5s} 1ª respuesta mediana {statistics.median(externos):.0f} ms "
              f"(docker incluido) · proceso listo en "
              f"{statistics.median(internos):.1f} ms" if internos else f"    · {lang}: sin interno")
    SALIDA["arranque"] = filas


# ── 4-6 · reposo, templado (JIT) y concurrencia ──────────────────────────
def medir_caliente():
    print("4-6 · memoria, templado (el JIT) y concurrencia")
    filas = {}
    for lang, imagen, puerto in LENGUAJES:
        nombre = f"govsjava-{lang}-caliente"
        parar(nombre)
        r = sh(["docker", "run", "-d", "--name", nombre, "-p", f"{puerto}:8080", imagen])
        if r.returncode != 0:
            raise SystemExit(f"[!] docker run {lang}: {r.stderr}")
        esperar_ok(puerto)
        time.sleep(2)
        mem_reposo = memoria_mib(nombre)
        print(f"    · {lang:5s} memoria en reposo {mem_reposo:.1f} MiB")

        # templado: 2.000 peticiones de CPU; el campo «ns» es el reloj del propio proceso
        externos, internos = [], []
        for _ in range(INVOCACIONES):
            t0 = time.perf_counter()
            with urllib.request.urlopen(
                    f"http://127.0.0.1:{puerto}/suma?n={N_CPU}", timeout=10) as r2:
                cuerpo = json.loads(r2.read())
            externos.append((time.perf_counter() - t0) * 1000)
            internos.append(cuerpo["ns"] / 1e6)
        primeras = internos[:20]
        ultimas = internos[-500:]
        print(f"      templado ext p50 {p(externos, .5):.2f} ms · p99 {p(externos, .99):.2f} ms")
        print(f"      cómputo interno: primeras 20 → {statistics.median(primeras):.3f} ms · "
              f"últimas 500 → {statistics.median(ultimas):.3f} ms")
        mem_tras_cpu = memoria_mib(nombre)

        # concurrencia: 300 a la vez, esperando 100 ms cada una
        t0 = time.perf_counter()
        with concurrent.futures.ThreadPoolExecutor(max_workers=CONCURRENTES) as ex:
            futs = [ex.submit(
                lambda: urllib.request.urlopen(
                    f"http://127.0.0.1:{puerto}/espera?ms={IO_MS}", timeout=30).read())
                for _ in range(CONCURRENTES)]
            mem_bajo_carga = memoria_mib(nombre)
            for f in futs:
                f.result()
        total_conc = (time.perf_counter() - t0) * 1000
        mem_tras_conc = memoria_mib(nombre)
        print(f"      {CONCURRENTES} a la vez × {IO_MS} ms → total {total_conc:.0f} ms · "
              f"memoria bajo carga {mem_bajo_carga:.1f} MiB")
        parar(nombre)

        filas[lang] = {
            "mem_reposo_mib": mem_reposo, "mem_tras_cpu_mib": mem_tras_cpu,
            "ext_ms": externos, "int_ms": internos,
            "conc_total_ms": total_conc, "mem_bajo_carga_mib": mem_bajo_carga,
            "mem_tras_conc_mib": mem_tras_conc,
        }
    SALIDA["caliente"] = filas


# ── el informe ───────────────────────────────────────────────────────────
def escribir_salida():
    v = {}
    for cmd, clave in [(["docker", "--version"], "docker"),
                       (["docker", "run", "--rm", "golang:1.26", "go", "version"], "go"),
                       (["docker", "run", "--rm", "eclipse-temurin:21-jdk", "java", "--version"],
                        "java")]:
        r = sh(cmd)
        v[clave] = ((r.stdout or "").strip().splitlines() or ["?"])[0]

    li = []
    w = li.append
    w("# SALIDA — Go contra Java, medidos en esta máquina")
    w("")
    w(f"> Corrida del {time.strftime('%Y-%m-%d %H:%M')}. Regenerable con")
    w("> `python investigacion/pruebas_go_java.py`. ⛔ Ninguna cifra de blog: todo salió de aquí.")
    w("")
    w(f"- `{v['docker']}`")
    w(f"- `{v['go']}` (imagen oficial `golang:1.26`)")
    w(f"- `{v['java']}` (imagen oficial `eclipse-temurin:21`)")
    w(f"- Mismo servidor HTTP, librería estándar en ambos; Java con hilos virtuales (Loom).")
    w(f"- Trabajo de CPU por petición: {N_CPU:,} vueltas de multiplicar y módulo.".replace(",", "."))
    w("")

    c = SALIDA["compilacion"]
    w("## 1 · Compilar (ms, dentro de la imagen oficial; reloj del contenedor)")
    w("")
    w("| lenguaje | corridas | mediana |")
    w("|---|---|---|")
    for lang in ("go", "java"):
        w(f"| {lang} | {c[lang]} | **{statistics.median(c[lang]):.0f} ms** |")
    w("")

    a = SALIDA["artefactos"]
    w("## 2 · Lo que despliegas")
    w("")
    w("| artefacto | bytes |")
    w("|---|---|")
    for k, n in a.items():
        w(f"| {k} | {n:,} B ({mib(n)}) |".replace(",", "."))
    w("")
    w("| imagen | descarga | en disco |")
    w("|---|---|---|")
    for ref, d in SALIDA["imagenes"].items():
        w(f"| `{ref}` | {mib(d['descarga'])} | {d['disco']} |")
    w("")

    ar = SALIDA["arranque"]
    w("## 3 · Arranque: contenedor nuevo → primera respuesta (mediana de "
      f"{REPES_FRIO})")
    w("")
    w("| lenguaje | 1ª respuesta (docker incluido) | proceso «listo» (su propio reloj) |")
    w("|---|---|---|")
    for lang in ("go", "java"):
        e = statistics.median(ar[lang]["externo_ms"])
        i = statistics.median(ar[lang]["interno_ms"]) if ar[lang]["interno_ms"] else None
        w(f"| {lang} | {e:.0f} ms | {i:.1f} ms |")
    w("")
    w("Todas las corridas (externo, ms): "
      + " · ".join(f"{lang} {[round(x) for x in ar[lang]['externo_ms']]}"
                   for lang in ("go", "java")))
    w("")

    ca = SALIDA["caliente"]
    w("## 4 · Memoria (docker stats, MiB)")
    w("")
    w("| lenguaje | reposo | tras 2.000 de CPU | bajo 300 concurrentes | tras la carga |")
    w("|---|---|---|---|---|")
    for lang in ("go", "java"):
        d = ca[lang]
        w(f"| {lang} | {d['mem_reposo_mib']:.1f} | {d['mem_tras_cpu_mib']:.1f} | "
          f"{d['mem_bajo_carga_mib']:.1f} | {d['mem_tras_conc_mib']:.1f} |")
    w("")

    w(f"## 5 · Templado y el JIT ({INVOCACIONES:,} peticiones, n={N_CPU:,})"
      .replace(",", "."))
    w("")
    w("| lenguaje | ext p50 | ext p99 | cómputo primeras 20 | cómputo últimas 500 |")
    w("|---|---|---|---|---|")
    for lang in ("go", "java"):
        d = ca[lang]
        w(f"| {lang} | {p(d['ext_ms'], .5):.2f} ms | {p(d['ext_ms'], .99):.2f} ms | "
          f"{statistics.median(d['int_ms'][:20]):.3f} ms | "
          f"{statistics.median(d['int_ms'][-500:]):.3f} ms |")
    w("")
    w("El «cómputo» es el campo `ns` que declara el propio servidor: el reloj del proceso,")
    w("sin red ni Docker por medio. Ahí se ve el JIT de Java trabajando.")
    w("")

    w(f"## 6 · Concurrencia: {CONCURRENTES} peticiones a la vez, {IO_MS} ms de espera cada una")
    w("")
    w("| lenguaje | total | secuencial teórico |")
    w("|---|---|---|")
    for lang in ("go", "java"):
        w(f"| {lang} | {ca[lang]['conc_total_ms']:.0f} ms | "
          f"{CONCURRENTES * IO_MS:,} ms |".replace(",", "."))
    w("")
    w("Java va con el executor de hilos virtuales de Java 21. El Java de 2015 no está invitado.")
    w("")

    destino = BASE / "SALIDA.md"
    destino.write_text("\n".join(li), encoding="utf-8")
    print(f"\n[ok] escrito {destino}")
    (BASE / "_salida_cruda.json").write_text(
        json.dumps(SALIDA, indent=1), encoding="utf-8")


if __name__ == "__main__":
    medir_compilacion()
    construir_y_medir_artefactos()
    medir_arranque()
    medir_caliente()
    escribir_salida()
