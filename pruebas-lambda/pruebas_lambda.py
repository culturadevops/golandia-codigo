# -*- coding: utf-8 -*-
"""pruebas_lambda.py — mide Go contra Python y contra Node en AWS Lambda, y guarda la salida
LITERAL en `recursos/pruebas-lambda/SALIDA.md`.

Vídeo 121 · `go-lambda-vs-python-node` · serie «Go serverless que se destruye sola».

## Por qué se mide así y no de otra forma

⛔ **Cero cifras de blogs.** El canon del canal dice que lo que sale en pantalla se ejecuta antes.
Aquí se ejecuta el MISMO handler en los tres lenguajes, cada uno dentro de **su imagen oficial de
AWS** (`public.ecr.aws/lambda/*`), que trae dentro el **Runtime Interface Emulator** — el mismo
que AWS publica para probar Lambda en local.

⚠️ **Qué mide esto y qué NO.** El emulador arranca el *runtime* de Lambda, no la microVM de
Firecracker. O sea: mide **el arranque del lenguaje y de tu código**, que es justo la parte que
depende del lenguaje. No mide lo que AWS tarda en encontrarte una microVM, que es igual para los
tres. Eso se dice EN PANTALLA en el vídeo; sin ese aviso el número engaña.

El número bueno es la línea `INIT REPORT(durationMs: …)` que imprime el propio emulador: es el
tiempo de inicialización que Lambda factura aparte de la invocación.

## Las cinco mediciones

    1. tamaño del binario/paquete que se sube      (y qué le hacen lambda.norpc y -s -w)
    2. tamaño de la imagen de contenedor           (la otra forma de desplegar)
    3. arranque en frío  → INIT REPORT, N contenedores recién levantados
    4. latencia templada → p50 y p99 con el contenedor ya caliente, trabajo de CPU
    5. el matiz honesto  → el mismo handler pero esperando 100 ms de E/S

    python investigacion/pruebas_lambda.py
"""
import io
import json
import re
import shutil
import statistics
import subprocess
import sys
import time
import urllib.error
import urllib.request
import zipfile
from pathlib import Path

sys.stdout = io.TextIOWrapper(sys.stdout.buffer, encoding="utf-8")

BASE = Path(__file__).resolve().parent.parent / "recursos" / "pruebas-lambda"
BIN = BASE / "go" / "_bin"

# El handler es el mismo trabajo en los tres: N vueltas de multiplicar y sacar módulo.
N_CPU = 200_000
IO_MS = 100
REPETICIONES_FRIO = 8
INVOCACIONES_TEMPLADO = 200
INVOCACIONES_IO = 40

LENGUAJES = [
    # clave      imagen                    puerto  base oficial
    ("go", "golandia-lambda-go:1", 9101, "public.ecr.aws/lambda/provided:al2023"),
    ("python", "golandia-lambda-py:1", 9102, "public.ecr.aws/lambda/python:3.13"),
    ("node", "golandia-lambda-node:1", 9103, "public.ecr.aws/lambda/nodejs:22"),
]

COMBOS = [
    ("plano", [], []),
    ("lambda.norpc", ["-tags", "lambda.norpc"], []),
    ("lambda.norpc + -s -w", ["-tags", "lambda.norpc"], ["-ldflags", "-s -w"]),
]


def sh(args, **kw):
    return subprocess.run(args, capture_output=True, text=True,
                          encoding="utf-8", errors="replace", **kw)


def mib(n):
    return f"{n/1048576:.2f} MiB"


def miles(n):
    return f"{n:,}".replace(",", ".")


# ── 1 · el paquete que se sube ───────────────────────────────────────────
def medir_binarios():
    print("1 · binarios de Go")
    BIN.mkdir(parents=True, exist_ok=True)
    filas = []
    for nombre, tags, ld in COMBOS:
        for arch in ("amd64", "arm64"):
            slug = re.sub(r"[^a-z0-9]+", "_", nombre.lower()).strip("_")
            salida = BIN / f"bootstrap-{slug}-{arch}"
            p = sh(["go", "build", *tags, *ld, "-o", str(salida), "."],
                   cwd=str(BASE / "go"),
                   env={**__import__("os").environ, "GOOS": "linux", "GOARCH": arch})
            if p.returncode != 0:
                raise SystemExit(f"[!] no compila {nombre}/{arch}:\n{p.stderr}")
            # El .zip es lo que de verdad se sube a Lambda.
            z = salida.with_suffix(".zip")
            with zipfile.ZipFile(z, "w", zipfile.ZIP_DEFLATED, compresslevel=9) as f:
                f.write(salida, arcname="bootstrap")
            filas.append({"combo": nombre, "arch": arch,
                          "bytes": salida.stat().st_size, "zip": z.stat().st_size})
            print(f"    · {nombre:22s} {arch}  {mib(salida.stat().st_size)}  "
                  f"(zip {mib(z.stat().st_size)})")
    # Y lo que se sube en Python y en Node: solo el fuente, el runtime lo pone AWS.
    otros = []
    for lang, ruta in (("python", BASE / "python" / "app.py"), ("node", BASE / "node" / "app.mjs")):
        z = BASE / f"_paquete-{lang}.zip"
        with zipfile.ZipFile(z, "w", zipfile.ZIP_DEFLATED, compresslevel=9) as f:
            f.write(ruta, arcname=ruta.name)
        otros.append({"lang": lang, "bytes": ruta.stat().st_size, "zip": z.stat().st_size})
        print(f"    · {lang:22s} fuente {ruta.stat().st_size} B  (zip {z.stat().st_size} B)")
    return filas, otros


# ── 2 · las imágenes ─────────────────────────────────────────────────────
def medir_imagenes():
    """⚠️ Dos números distintos, y confundirlos es publicar una cifra falsa.

    `docker image inspect --format {{.Size}}` devuelve, con el almacén de containerd, la suma de
    las **capas comprimidas** del registro: lo que te DESCARGAS. Se comprobó resolviendo a mano
    el manifiesto de amd64 con `docker manifest inspect` — sale el mismo byte.
    `docker images` enseña la imagen **descomprimida**: lo que OCUPA en tu disco.
    """
    print("2 · imágenes de contenedor")
    en_disco = {}
    p = sh(["docker", "images", "--format", "{{.Repository}}:{{.Tag}}\t{{.Size}}"])
    for linea in (p.stdout or "").splitlines():
        if "\t" in linea:
            ref, tam = linea.split("\t", 1)
            en_disco[ref] = tam.strip()
    filas = []
    for clave, imagen, _, base in LENGUAJES:
        for etiqueta, ref in (("base oficial", base), ("con el handler", imagen)):
            p = sh(["docker", "image", "inspect", ref, "--format", "{{.Size}}"])
            if p.returncode != 0:
                raise SystemExit(f"[!] falta la imagen {ref}")
            n = int(p.stdout.strip())
            filas.append({"lang": clave, "que": etiqueta, "ref": ref,
                          "descarga_bytes": n, "en_disco": en_disco.get(ref, "?")})
            print(f"    · {clave:8s} {etiqueta:16s} descarga {mib(n):>10s}  "
                  f"en disco {en_disco.get(ref, '?')}")
    return filas


# ── el emulador ──────────────────────────────────────────────────────────
def invocar(puerto, cuerpo, timeout=30):
    d = json.dumps(cuerpo).encode()
    r = urllib.request.Request(
        f"http://127.0.0.1:{puerto}/2015-03-31/functions/function/invocations",
        data=d, method="POST")
    t0 = time.perf_counter()
    with urllib.request.urlopen(r, timeout=timeout) as resp:
        texto = resp.read().decode("utf-8", "replace")
    return (time.perf_counter() - t0) * 1000.0, texto


def levantar(nombre, imagen, puerto):
    sh(["docker", "rm", "-f", nombre])
    p = sh(["docker", "run", "-d", "--name", nombre, "-p", f"{puerto}:8080", imagen])
    if p.returncode != 0:
        raise SystemExit(f"[!] no levanta {imagen}: {p.stderr}")
    return p.stdout.strip()


def init_ms(nombre):
    """El número bueno: lo que el propio emulador declara como inicialización."""
    p = sh(["docker", "logs", nombre])
    todo = (p.stdout or "") + (p.stderr or "")
    m = re.findall(r"INIT REPORT\(durationMs: ([\d.]+)\)", todo)
    return float(m[0]) if m else None


def duraciones_facturadas(nombre):
    p = sh(["docker", "logs", nombre])
    todo = (p.stdout or "") + (p.stderr or "")
    return [float(x) for x in re.findall(r"\tDuration: ([\d.]+) ms", todo)]


# ── 3 · arranque en frío ─────────────────────────────────────────────────
def medir_frio():
    print(f"3 · arranque en frío ({REPETICIONES_FRIO} contenedores nuevos por lenguaje)")
    res = {}
    for clave, imagen, puerto, _ in LENGUAJES:
        inits, primeras = [], []
        for i in range(REPETICIONES_FRIO):
            nombre = f"frio-{clave}-{i}"
            levantar(nombre, imagen, puerto)
            ms, texto = None, ""
            for _ in range(200):
                try:
                    ms, texto = invocar(puerto, {"n": N_CPU, "io_ms": 0}, timeout=5)
                    break
                except (urllib.error.URLError, OSError):
                    time.sleep(0.05)
            if ms is None:
                raise SystemExit(f"[!] {clave}: el emulador no respondió")
            if '"lang"' not in texto:
                raise SystemExit(f"[!] {clave}: respuesta rara -> {texto[:200]}")
            time.sleep(0.2)                       # que el emulador termine de escribir el log
            ini = init_ms(nombre)
            if ini is None:
                raise SystemExit(f"[!] {clave}: sin línea INIT REPORT en el log")
            inits.append(ini)
            primeras.append(ms)
            sh(["docker", "rm", "-f", nombre])
        res[clave] = {"init_ms": inits, "primera_ms": primeras}
        print(f"    · {clave:8s} init mediana {statistics.median(inits):8.1f} ms   "
              f"1ª respuesta mediana {statistics.median(primeras):8.1f} ms")
    return res


# ── 4 y 5 · templado y E/S ───────────────────────────────────────────────
def pct(v, q):
    v = sorted(v)
    k = min(len(v) - 1, int(round(q * (len(v) - 1))))
    return v[k]


def medir_templado():
    print(f"4 · latencia templada ({INVOCACIONES_TEMPLADO} invocaciones, {miles(N_CPU)} vueltas)")
    print(f"5 · el matiz: el mismo handler esperando {IO_MS} ms de E/S")
    res = {}
    for clave, imagen, puerto, _ in LENGUAJES:
        nombre = f"templado-{clave}"
        levantar(nombre, imagen, puerto)
        for _ in range(200):                       # calentar
            try:
                invocar(puerto, {"n": 1, "io_ms": 0}, timeout=5)
                break
            except (urllib.error.URLError, OSError):
                time.sleep(0.05)
        for _ in range(20):
            invocar(puerto, {"n": N_CPU, "io_ms": 0})
        marca = len(duraciones_facturadas(nombre))
        cpu = [invocar(puerto, {"n": N_CPU, "io_ms": 0})[0]
               for _ in range(INVOCACIONES_TEMPLADO)]
        facturada = duraciones_facturadas(nombre)[marca:]
        es = [invocar(puerto, {"n": 0, "io_ms": IO_MS})[0] for _ in range(INVOCACIONES_IO)]
        sh(["docker", "rm", "-f", nombre])
        res[clave] = {"cpu_ms": cpu, "facturada_ms": facturada, "io_ms": es}
        print(f"    · {clave:8s} cpu p50 {pct(cpu,.5):7.2f} ms  p99 {pct(cpu,.99):7.2f} ms   "
              f"| E/S p50 {pct(es,.5):7.2f} ms")
    return res


# ── el informe ───────────────────────────────────────────────────────────
def entorno():
    def uno(args):
        p = sh(args)
        return (p.stdout or p.stderr or "").strip().splitlines()[0]
    return {"go": uno(["go", "version"]), "docker": uno(["docker", "--version"])}


# ── 6 · la factura, con los precios publicados por AWS ───────────────────
# ⛔ Estos tres números NO se miden aquí: se COPIAN de la página de precios de AWS. La fuente va
# escrita en el informe para que se pueda comprobar. Lo que sí es nuestro es la aritmética.
PRECIO_PETICIONES_MILLON = 0.20          # US$ por millón de peticiones
PRECIO_GB_S_X86 = 0.0000166667           # US$ por GB-segundo, x86
GRATIS_PETICIONES = 1_000_000            # siempre gratis, al mes
GRATIS_GB_S = 400_000                    # siempre gratis, al mes
FUENTE_PRECIOS = "https://aws.amazon.com/lambda/pricing/ (us-east-1, consultado 2026-08-25)"


def factura(ms_facturados, invocaciones, mb_memoria):
    """Lambda factura la duración redondeada hacia arriba al milisegundo."""
    import math
    ms = max(1, math.ceil(ms_facturados))
    gb_s = (mb_memoria / 1024.0) * (ms / 1000.0) * invocaciones
    gb_s_cobrados = max(0.0, gb_s - GRATIS_GB_S)
    pet_cobradas = max(0, invocaciones - GRATIS_PETICIONES)
    coste_pet = pet_cobradas / 1_000_000 * PRECIO_PETICIONES_MILLON
    coste_cpu = gb_s_cobrados * PRECIO_GB_S_X86
    return {"ms_facturados": ms, "gb_s": gb_s, "coste_peticiones": coste_pet,
            "coste_computo": coste_cpu, "total": coste_pet + coste_cpu}


def medir_factura(templado):
    print("6 · la factura (aritmética sobre los precios publicados por AWS)")
    escenarios = [(10_000_000, 128), (1_000_000_000, 128), (1_000_000_000, 1024)]
    filas = []
    for invocaciones, mb in escenarios:
        for clave, _, _, _ in LENGUAJES:
            d = templado[clave]["facturada_ms"]
            p50 = pct(d, .5) if d else 0.0
            f = factura(p50, invocaciones, mb)
            filas.append({"lang": clave, "invocaciones": invocaciones, "mb": mb,
                          "p50_medido_ms": p50, **f})
        g = [x for x in filas if x["invocaciones"] == invocaciones and x["mb"] == mb]
        peor, mejor = max(g, key=lambda x: x["total"]), min(g, key=lambda x: x["total"])
        print(f"    · {miles(invocaciones)} invocaciones a {mb} MB → "
              f"mejor {mejor['lang']} {mejor['total']:.2f} $ · "
              f"peor {peor['lang']} {peor['total']:.2f} $ · "
              f"diferencia {peor['total']-mejor['total']:.2f} $")
    return filas


def informe(bins, otros, imgs, frio, templado, cuentas, env):
    L = []
    A = L.append
    A("# Pruebas ejecutadas — ¿Go es el mejor lenguaje para AWS Lambda?")
    A("")
    A("> Salida **literal** de esta máquina. La genera `investigacion/pruebas_lambda.py` y se")
    A("> vuelve a generar sola. ⛔ Lo que no esté aquí no puede salir en pantalla.")
    A("")
    A(f"**Entorno citable en el vídeo:** `{env['go']}` · `{env['docker']}` · "
      "`github.com/aws/aws-lambda-go v1.54.0`")
    A("")
    A("**Imágenes oficiales de AWS usadas** (traen dentro el Runtime Interface Emulator):")
    for _, _, _, base in LENGUAJES:
        A(f"- `{base}`")
    A("")
    A("⚠️ **Qué mide esto y qué NO.** El emulador arranca el *runtime* de Lambda, no la microVM")
    A("de Firecracker. Mide **el arranque del lenguaje y de tu código** —la parte que depende del")
    A("lenguaje—, no lo que AWS tarda en encontrarte una microVM, que es igual para los tres.")
    A("")
    A("El handler es el mismo trabajo en los tres lenguajes: `n` vueltas de `i*i % 1000003`, y")
    A("opcionalmente esperar `io_ms` milisegundos.")
    A("")

    A("---")
    A("")
    A("## 1 · Lo que se sube: el paquete")
    A("")
    A("| compilado con | arquitectura | binario | .zip que se sube |")
    A("|---|---|---:|---:|")
    for f in bins:
        A(f"| `{f['combo']}` | {f['arch']} | {miles(f['bytes'])} B ({mib(f['bytes'])}) | "
          f"{miles(f['zip'])} B ({mib(f['zip'])}) |")
    A("")
    A("Y lo que se sube en los otros dos, donde el runtime lo pone AWS:")
    A("")
    A("| lenguaje | fuente | .zip que se sube |")
    A("|---|---:|---:|")
    for f in otros:
        A(f"| {f['lang']} | {miles(f['bytes'])} B | {miles(f['zip'])} B |")
    A("")
    plano = next(f for f in bins if f["combo"] == "plano" and f["arch"] == "amd64")
    mejor = next(f for f in bins if f["combo"].startswith("lambda.norpc +")
                 and f["arch"] == "arm64")
    A(f"> 📉 Del binario plano de amd64 ({mib(plano['bytes'])}) al de arm64 con "
      f"`lambda.norpc` y `-s -w` ({mib(mejor['bytes'])}): "
      f"**−{100*(1-mejor['bytes']/plano['bytes']):.1f} %**.")
    A("")

    A("---")
    A("")
    A("## 2 · La otra forma de desplegar: la imagen")
    A("")
    A("⚠️ Son **dos números distintos** y confundirlos es publicar una cifra falsa: lo que te")
    A("descargas (capas comprimidas del registro) no es lo que ocupa en tu disco (descomprimida).")
    A("La columna de descarga se comprobó resolviendo a mano el manifiesto de amd64 con")
    A("`docker manifest inspect`: sale el mismo byte.")
    A("")
    A("| lenguaje | qué | referencia | descarga | en disco |")
    A("|---|---|---|---:|---:|")
    for f in imgs:
        A(f"| {f['lang']} | {f['que']} | `{f['ref']}` | {mib(f['descarga_bytes'])} | "
          f"{f['en_disco']} |")
    A("")

    A("---")
    A("")
    A(f"## 3 · Arranque en frío — {REPETICIONES_FRIO} contenedores nuevos por lenguaje")
    A("")
    A("`INIT REPORT(durationMs: …)` del propio emulador: la inicialización que Lambda factura")
    A("aparte. «1ª respuesta» es lo que tarda el cliente desde que el contenedor existe.")
    A("")
    A("| lenguaje | init mediana | init mínimo | init máximo | 1ª respuesta mediana |")
    A("|---|---:|---:|---:|---:|")
    for clave, _, _, _ in LENGUAJES:
        d = frio[clave]
        A(f"| {clave} | {statistics.median(d['init_ms']):.1f} ms | {min(d['init_ms']):.1f} ms | "
          f"{max(d['init_ms']):.1f} ms | {statistics.median(d['primera_ms']):.1f} ms |")
    A("")
    A("<details><summary>las medidas, una a una</summary>")
    A("")
    A("```")
    for clave, _, _, _ in LENGUAJES:
        A(f"{clave:8s} init(ms)  " + "  ".join(f"{x:.1f}" for x in frio[clave]["init_ms"]))
    A("```")
    A("")
    A("</details>")
    A("")

    A("---")
    A("")
    A(f"## 4 · Latencia templada — {INVOCACIONES_TEMPLADO} invocaciones con "
      f"{miles(N_CPU)} vueltas de CPU")
    A("")
    A("| lenguaje | p50 | p90 | p99 | duración facturada p50 |")
    A("|---|---:|---:|---:|---:|")
    for clave, _, _, _ in LENGUAJES:
        d = templado[clave]
        fact = f"{pct(d['facturada_ms'], .5):.2f} ms" if d["facturada_ms"] else "—"
        A(f"| {clave} | {pct(d['cpu_ms'],.5):.2f} ms | {pct(d['cpu_ms'],.9):.2f} ms | "
          f"{pct(d['cpu_ms'],.99):.2f} ms | {fact} |")
    A("")

    A("---")
    A("")
    A(f"## 5 · El matiz honesto — el mismo handler esperando {IO_MS} ms de E/S")
    A("")
    A("Aquí no hay trabajo de CPU: solo espera, que es lo que hace tu Lambda de verdad mientras")
    A("DynamoDB o una API ajena le contestan.")
    A("")
    A("| lenguaje | p50 | p99 |")
    A("|---|---:|---:|")
    for clave, _, _, _ in LENGUAJES:
        d = templado[clave]
        A(f"| {clave} | {pct(d['io_ms'],.5):.2f} ms | {pct(d['io_ms'],.99):.2f} ms |")
    A("")
    p50 = {c: pct(templado[c]["io_ms"], .5) for c, _, _, _ in LENGUAJES}
    A(f"> 🥇 Entre el mejor y el peor hay **{max(p50.values())-min(p50.values()):.2f} ms** sobre "
      f"{IO_MS} ms de espera. Cuando tu Lambda espera, el lenguaje no es tu problema.")
    A("")

    A("---")
    A("")
    A("## 6 · La factura")
    A("")
    A("⛔ Los precios **no se miden aquí: se copian de AWS**, y la fuente va escrita para que se")
    A("pueda comprobar. Lo que es nuestro es la aritmética y la duración facturada medida arriba.")
    A("")
    A(f"**Fuente de los precios:** {FUENTE_PRECIOS}")
    A("")
    A(f"- {PRECIO_PETICIONES_MILLON:.2f} $ por millón de peticiones")
    A(f"- {PRECIO_GB_S_X86:.10f} $ por GB-segundo (x86)")
    A(f"- siempre gratis al mes: {miles(GRATIS_PETICIONES)} peticiones y "
      f"{miles(GRATIS_GB_S)} GB-s")
    A("- Lambda redondea la duración **hacia arriba al milisegundo**")
    A("")
    A("| invocaciones/mes | memoria | lenguaje | p50 facturada medida | ms cobrados | GB-s | "
      "peticiones | cómputo | **total/mes** |")
    A("|---:|---:|---|---:|---:|---:|---:|---:|---:|")
    for f in cuentas:
        A(f"| {miles(f['invocaciones'])} | {f['mb']} MB | {f['lang']} | "
          f"{f['p50_medido_ms']:.2f} ms | {f['ms_facturados']} ms | {miles(round(f['gb_s']))} | "
          f"{f['coste_peticiones']:.2f} $ | {f['coste_computo']:.2f} $ | "
          f"**{f['total']:.2f} $** |")
    A("")
    for inv, mb in ((10_000_000, 128), (1_000_000_000, 128), (1_000_000_000, 1024)):
        g = [x for x in cuentas if x["invocaciones"] == inv and x["mb"] == mb]
        peor, mejor = max(g, key=lambda x: x["total"]), min(g, key=lambda x: x["total"])
        A(f"> Con {miles(inv)} invocaciones al mes a {mb} MB, entre el mejor ({mejor['lang']}) y "
          f"el peor ({peor['lang']}) hay **{peor['total']-mejor['total']:.2f} $ al mes**.")
        A("")
    return "\n".join(L)


def main():
    # ⚠️ `--solo-informe` reescribe SALIDA.md a partir del SALIDA.json que ya está en disco, sin
    # volver a medir. Existe porque cada corrida da números ligeramente distintos: si el guion ya
    # está escrito contra unos números, retocar la redacción del informe NO puede moverlos.
    if "--solo-informe" in sys.argv:
        crudo = json.loads((BASE / "SALIDA.json").read_text(encoding="utf-8"))
        doc = informe(crudo["binarios"], crudo["otros"], crudo["imagenes"], crudo["frio"],
                      crudo["templado"], crudo["factura"], crudo["entorno"])
        (BASE / "SALIDA.md").write_text(doc, encoding="utf-8")
        print(f"✅ (solo informe) {BASE / 'SALIDA.md'}")
        return
    if not shutil.which("docker"):
        raise SystemExit("[!] no hay docker")
    env = entorno()
    bins, otros = medir_binarios()
    imgs = medir_imagenes()
    frio = medir_frio()
    templado = medir_templado()
    cuentas = medir_factura(templado)
    doc = informe(bins, otros, imgs, frio, templado, cuentas, env)
    (BASE / "SALIDA.md").write_text(doc, encoding="utf-8")
    crudo = {"frio": frio, "templado": templado, "binarios": bins, "otros": otros,
             "imagenes": imgs, "factura": cuentas, "entorno": env}
    (BASE / "SALIDA.json").write_text(json.dumps(crudo, indent=1), encoding="utf-8")
    print(f"\n✅ {BASE / 'SALIDA.md'}")


if __name__ == "__main__":
    main()
