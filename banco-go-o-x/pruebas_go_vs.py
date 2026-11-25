# -*- coding: utf-8 -*-
"""pruebas_go_vs.py — mide Go contra otro lenguaje con el MISMO método del vídeo 148 (Go contra Java) y guarda la salida
LITERAL en `recursos/pruebas-go-vs/SALIDA-<comparación>.md`.

⛔ Cero cifras de blogs: el mismo servidor HTTP en cada lenguaje (solo lo que trae su plataforma), compilado en su imagen
oficial y desplegado en la imagen que se usaría de verdad. Contrato común: /suma?n= hace CPU (multiplicar y módulo) y
devuelve el reloj del propio proceso en «ns»; /espera?ms= simula E/S; el proceso imprime «LISTO_EN_MS x» al escuchar.

Las mediciones (igual que en el 148):
    1. compilar            → en la imagen oficial, contenedor nuevo (caché vacía), mediana de 3
    2. lo que despliegas   → el artefacto (/out de la etapa build) y la imagen final
    3. arranque            → contenedor nuevo → primera respuesta 200, mediana de 8, + LISTO_EN_MS
    4. memoria             → docker stats: reposo, tras la CPU, bajo carga, después
    5. templado            → 2.000 peticiones de CPU: p50/p99 y cómputo primeras 20 contra últimas 500
    6. concurrencia        → 300 peticiones a la vez esperando 100 ms

    python investigacion/pruebas_go_vs.py csharp|kotlin|zig|elixir
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

BASE = Path(__file__).resolve().parent.parent / "recursos" / "pruebas-go-vs"
N_CPU = 200_000
REPES_COMPILA = 3
REPES_FRIO = 8
INVOCACIONES = 2_000
CONCURRENTES = 300
IO_MS = 100

# clave: (carpeta, imagen para compilar, orden de compilar, versión: imagen + orden, etiqueta)
L = {
    "go": ("go", "golang:1.27.2", "CGO_ENABLED=0 go build -o /tmp/servidor .",
           ("golang:1.27.2", "go version"), "Go 1.27"),
    "csharp": ("csharp", "mcr.microsoft.com/dotnet/sdk:10.0", "dotnet publish -c Release -o /tmp/out >/dev/null",
               ("mcr.microsoft.com/dotnet/sdk:10.0", "dotnet --version"), "C# .NET 10 (JIT)"),
    "csharp-aot": ("csharp-aot", "mcr.microsoft.com/dotnet/sdk:10.0-noble-aot",
                   "dotnet publish -c Release -p:PublishAot=true -o /tmp/out >/dev/null",
                   ("mcr.microsoft.com/dotnet/sdk:10.0-noble-aot", "dotnet --version"), "C# .NET 10 (Native AOT)"),
    "kotlin": ("kotlin", "golandia-kotlinc:2.4.21", "kotlinc Servidor.kt -include-runtime -d /tmp/servidor.jar 2>/dev/null",
               ("golandia-kotlinc:2.4.21", "kotlinc -version 2>&1"), "Kotlin 2.4 (JVM 21)"),
    "zig": ("zig", "golandia-zig:0.17.0",
            "zig build-exe main.zig -O ReleaseFast -target x86_64-linux-musl -femit-bin=/tmp/servidor",
            ("golandia-zig:0.17.0", "echo zig $(zig version)"), "Zig 0.17"),
    "elixir": ("elixir", "golandia-elixir-compilar:1", "mix compile >/dev/null",
               ("elixir:1.20.4-otp-29", "elixir --version | tail -1"), "Elixir 1.20 (OTP 29)"),
}
COMPARA = {"csharp": ["go", "csharp", "csharp-aot"], "kotlin": ["go", "kotlin"], "zig": ["go", "zig"],
           "elixir": ["go", "elixir"]}

SALIDA = {}


def sh(args, **kw):
    return subprocess.run(args, capture_output=True, text=True, encoding="utf-8", errors="replace", **kw)


def mib(n):
    return f"{n / 1048576:.2f} MiB"


def p(v, q):
    v = sorted(v)
    return v[max(0, min(len(v) - 1, round(q * (len(v) - 1))))]


def tag(lang):
    return f"golandia-govs-{lang}:1"


def medir_compilacion(langs):
    print("1 · compilar (imagen oficial, contenedor nuevo)")
    filas = {}
    for lang in langs:
        carpeta, imagen, orden, _, _ = L[lang]
        tiempos = []
        for _ in range(REPES_COMPILA):
            r = sh(["docker", "run", "--rm", "-v", f"{BASE / carpeta}:/src:ro", "-w", "/work", imagen, "sh", "-c",
                    f"cp -r /src/. /work/ && t0=$(date +%s%N); {orden}; t1=$(date +%s%N); "
                    "echo COMPILA_MS $(( (t1 - t0) / 1000000 ))"])
            ms = [int(x.split()[1]) for x in r.stdout.splitlines() if x.startswith("COMPILA_MS")]
            if r.returncode != 0 or not ms:
                raise SystemExit(f"[!] no compila {lang}:\n{r.stdout[-2000:]}\n{r.stderr[-2000:]}")
            tiempos.append(ms[0])
        filas[lang] = tiempos
        print(f"    · {lang:11s} {tiempos} ms → mediana {statistics.median(tiempos):.0f} ms")
    SALIDA["compilacion"] = filas


def construir(langs):
    print("2 · artefactos e imágenes de despliegue")
    art, img = {}, {}
    for lang in langs:
        carpeta = BASE / L[lang][0]
        r = sh(["docker", "build", "--target", "build", "-t", f"golandia-govs-{lang}-build:1", str(carpeta)])
        if r.returncode != 0:
            raise SystemExit(f"[!] build {lang}:\n{r.stderr[-3000:]}")
        r = sh(["docker", "run", "--rm", f"golandia-govs-{lang}-build:1", "du", "-sb", "/out"])
        art[lang] = int(r.stdout.split()[0])
        r = sh(["docker", "build", "-t", tag(lang), str(carpeta)])
        if r.returncode != 0:
            raise SystemExit(f"[!] imagen {lang}:\n{r.stderr[-3000:]}")
        rr = sh(["docker", "image", "inspect", tag(lang), "--format", "{{.Size}}"])
        img[lang] = int(rr.stdout.strip())
        print(f"    · {lang:11s} artefacto {art[lang]:>11,} B ({mib(art[lang])}) · imagen {mib(img[lang])}".replace(",", "."))
    SALIDA["artefactos"], SALIDA["imagenes"] = art, img


def parar(nombre):
    sh(["docker", "rm", "-f", nombre])


def esperar_ok(puerto, timeout=90):
    fin = time.perf_counter() + timeout
    while time.perf_counter() < fin:
        try:
            with urllib.request.urlopen(f"http://127.0.0.1:{puerto}/suma?n=1", timeout=2) as r:
                if r.status == 200:
                    r.read()
                    return time.perf_counter()
        except (urllib.error.URLError, ConnectionError, OSError):
            time.sleep(0.002)
    raise SystemExit(f"[!] el puerto {puerto} nunca contestó")


def listo_interno(nombre):
    for _ in range(60):
        r = sh(["docker", "logs", nombre])
        for linea in (r.stderr or "").splitlines() + (r.stdout or "").splitlines():
            if linea.startswith("LISTO_EN_MS"):
                return float(linea.split()[1].replace(",", "."))
        time.sleep(0.05)
    return None


def memoria_mib(nombre):
    bruto = (sh(["docker", "stats", "--no-stream", "--format", "{{.MemUsage}}", nombre]).stdout or "").strip().split("/")[0].strip()
    for suf, f in (("GiB", 1024), ("MiB", 1), ("KiB", 1 / 1024)):
        if bruto.endswith(suf):
            return float(bruto[:-3]) * f
    return None


def medir_arranque(langs):
    print("3 · arranque: contenedor nuevo → primera respuesta")
    filas = {}
    for k, lang in enumerate(langs):
        puerto = 9301 + k
        ext, inte = [], []
        for i in range(REPES_FRIO):
            nombre = f"govs-{lang}-frio-{i}"
            parar(nombre)
            t0 = time.perf_counter()
            r = sh(["docker", "run", "-d", "--name", nombre, "-p", f"{puerto}:8080", tag(lang)])
            if r.returncode != 0:
                raise SystemExit(f"[!] run {lang}: {r.stderr}")
            ext.append((esperar_ok(puerto) - t0) * 1000)
            li = listo_interno(nombre)
            if li is not None:
                inte.append(li)
            parar(nombre)
        filas[lang] = {"externo_ms": ext, "interno_ms": inte}
        print(f"    · {lang:11s} 1ª respuesta {statistics.median(ext):.0f} ms · listo "
              f"{statistics.median(inte) if inte else float('nan'):.1f} ms")
    SALIDA["arranque"] = filas


def medir_caliente(langs):
    print("4-6 · memoria, templado y concurrencia")
    filas = {}
    for k, lang in enumerate(langs):
        puerto = 9401 + k
        nombre = f"govs-{lang}-caliente"
        parar(nombre)
        sh(["docker", "run", "-d", "--name", nombre, "-p", f"{puerto}:8080", tag(lang)])
        esperar_ok(puerto)
        time.sleep(2)
        reposo = memoria_mib(nombre)
        ext, inte = [], []
        for _ in range(INVOCACIONES):
            t0 = time.perf_counter()
            with urllib.request.urlopen(f"http://127.0.0.1:{puerto}/suma?n={N_CPU}", timeout=10) as r2:
                cuerpo = json.loads(r2.read())
            ext.append((time.perf_counter() - t0) * 1000)
            inte.append(cuerpo["ns"] / 1e6)
        suma = cuerpo["suma"]
        tras_cpu = memoria_mib(nombre)
        t0 = time.perf_counter()
        with concurrent.futures.ThreadPoolExecutor(max_workers=CONCURRENTES) as ex:
            futs = [ex.submit(lambda: urllib.request.urlopen(f"http://127.0.0.1:{puerto}/espera?ms={IO_MS}", timeout=60).read())
                    for _ in range(CONCURRENTES)]
            bajo_carga = memoria_mib(nombre)
            for f in futs:
                f.result()
        conc = (time.perf_counter() - t0) * 1000
        despues = memoria_mib(nombre)
        parar(nombre)
        filas[lang] = {"mem_reposo_mib": reposo, "mem_tras_cpu_mib": tras_cpu, "mem_bajo_carga_mib": bajo_carga,
                       "mem_tras_conc_mib": despues, "ext_ms": ext, "int_ms": inte, "conc_total_ms": conc, "suma": suma}
        print(f"    · {lang:11s} reposo {reposo:.1f} MiB · p50 {p(ext, .5):.2f} ms · cómputo últimas 500 "
              f"{statistics.median(inte[-500:]):.3f} ms · {CONCURRENTES} a la vez {conc:.0f} ms · suma {suma}")
    SALIDA["caliente"] = filas


def informe(comp, langs):
    ver = {"docker": (sh(["docker", "--version"]).stdout or "?").strip()}
    for lang in langs:
        img, orden = L[lang][3]
        ver[lang] = ((sh(["docker", "run", "--rm", img, "sh", "-c", orden]).stdout or "?").strip().splitlines() or ["?"])[0]
    li = []
    w = li.append
    w(f"# SALIDA — Go contra {comp}, medidos en esta máquina")
    w("")
    w(f"> Corrida del {time.strftime('%Y-%m-%d %H:%M')}. Regenerable con `python investigacion/pruebas_go_vs.py {comp}`.")
    w("> ⛔ Ninguna cifra de blog: todo salió de aquí. Mismo método que el vídeo 148 (Go contra Java).")
    w("")
    w(f"- `{ver['docker']}`")
    for lang in langs:
        w(f"- {L[lang][4]}: `{ver[lang]}` (compila en `{L[lang][1]}`)")
    w(f"- Trabajo de CPU por petición: {N_CPU:,} vueltas de multiplicar y módulo.".replace(",", "."))
    w("")
    w("## 1 · Compilar (ms, imagen oficial, contenedor nuevo: caché vacía)")
    w("")
    w("| lenguaje | corridas | mediana |")
    w("|---|---|---|")
    for lang in langs:
        c = SALIDA["compilacion"][lang]
        w(f"| {L[lang][4]} | {c} | **{statistics.median(c):.0f} ms** |")
    w("")
    w("## 2 · Lo que despliegas")
    w("")
    w("| lenguaje | artefacto (/out) | imagen final (comprimida: lo que se descarga) |")
    w("|---|---|---|")
    for lang in langs:
        a = SALIDA["artefactos"][lang]
        w(f"| {L[lang][4]} | {a:,} B ({mib(a)}) | {mib(SALIDA['imagenes'][lang])} |".replace(",", "."))
    w("")
    w(f"## 3 · Arranque: contenedor nuevo → primera respuesta (mediana de {REPES_FRIO})")
    w("")
    w("| lenguaje | 1ª respuesta (docker incluido) | proceso «listo» (su propio reloj) |")
    w("|---|---|---|")
    for lang in langs:
        a = SALIDA["arranque"][lang]
        i = f"{statistics.median(a['interno_ms']):.1f} ms" if a["interno_ms"] else "?"
        w(f"| {L[lang][4]} | {statistics.median(a['externo_ms']):.0f} ms | {i} |")
    w("")
    w("Todas las corridas (externo, ms): " + " · ".join(
        f"{lang} {[round(x) for x in SALIDA['arranque'][lang]['externo_ms']]}" for lang in langs))
    w("")
    w("## 4 · Memoria (docker stats, MiB)")
    w("")
    w(f"| lenguaje | reposo | tras {INVOCACIONES:,} de CPU | bajo {CONCURRENTES} concurrentes | tras la carga |".replace(",", "."))
    w("|---|---|---|---|---|")
    for lang in langs:
        d = SALIDA["caliente"][lang]
        w(f"| {L[lang][4]} | {d['mem_reposo_mib']:.1f} | {d['mem_tras_cpu_mib']:.1f} | {d['mem_bajo_carga_mib']:.1f} | "
          f"{d['mem_tras_conc_mib']:.1f} |")
    w("")
    w(f"## 5 · Templado ({INVOCACIONES:,} peticiones, n={N_CPU:,})".replace(",", "."))
    w("")
    w("| lenguaje | ext p50 | ext p99 | cómputo primeras 20 | cómputo últimas 500 | suma devuelta |")
    w("|---|---|---|---|---|---|")
    for lang in langs:
        d = SALIDA["caliente"][lang]
        w(f"| {L[lang][4]} | {p(d['ext_ms'], .5):.2f} ms | {p(d['ext_ms'], .99):.2f} ms | "
          f"{statistics.median(d['int_ms'][:20]):.3f} ms | {statistics.median(d['int_ms'][-500:]):.3f} ms | {d['suma']} |")
    w("")
    w("El «cómputo» es el campo `ns` que declara el propio servidor: el reloj del proceso, sin red ni Docker por medio.")
    w("")
    w(f"## 6 · Concurrencia: {CONCURRENTES} peticiones a la vez, {IO_MS} ms de espera cada una")
    w("")
    w("| lenguaje | total | secuencial teórico |")
    w("|---|---|---|")
    for lang in langs:
        w(f"| {L[lang][4]} | {SALIDA['caliente'][lang]['conc_total_ms']:.0f} ms | {CONCURRENTES * IO_MS:,} ms |".replace(",", "."))
    w("")
    destino = BASE / f"SALIDA-{comp}.md"
    destino.write_text("\n".join(li), encoding="utf-8")
    (BASE / f"_salida_cruda-{comp}.json").write_text(json.dumps(SALIDA, indent=1), encoding="utf-8")
    print(f"\n[ok] escrito {destino}")


if __name__ == "__main__":
    comp = sys.argv[1]
    langs = COMPARA[comp]
    medir_compilacion(langs)
    construir(langs)
    medir_arranque(langs)
    medir_caliente(langs)
    informe(comp, langs)
