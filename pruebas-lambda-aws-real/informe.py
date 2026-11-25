# -*- coding: utf-8 -*-
"""Escribe SALIDA.md a partir de _datos/*.json (lo que dejaron construir, medir, memoria y limpiar).
Nada a mano: cada cifra sale de un JSON que escribió banco.py. Se llama con `python banco.py informe`."""
import json
import math
import re
import statistics as st
from pathlib import Path

import banco as B

AQUI = Path(__file__).resolve().parent
V121 = AQUI.parent / "pruebas-lambda" / "SALIDA.json"

# Precios copiados de la página oficial (no se miden aquí). us-east-1, x86.
FUENTE_PRECIOS = "https://aws.amazon.com/lambda/pricing/ (us-east-1, x86, consultado 2026-11-18)"
PRECIO_PETICION = 0.20 / 1_000_000
PRECIO_GBS = 0.0000166667

NOMBRE = {v: c["nombre"] for v, c in B.VARIANTES.items()}
SIN_GO = ["python", "node", "java", "java-snap", "dotnet-aot"]  # el bloque para cultura-devops


def pct(vals, p):
    v = sorted(vals)
    return v[max(0, math.ceil(p / 100 * len(v)) - 1)]


def med(vals):
    return st.median(vals)


def n(x, d=1):
    return f"{x:.{d}f}"


def miles(x):
    return f"{int(x):,}".replace(",", ".")


def arranque(r):
    """init (o restore en SnapStart) de una línea REPORT."""
    return r.get("init", r.get("restore"))


def cargar_todo():
    datos = {}
    for v in B.ORDEN:
        d = B.cargar(f"medida-{v}.json")
        if d and d.get("caliente_cpu"):
            datos[v] = d
    return datos


def version_runtime(d):
    for linea in d.get("init_start", []):
        m = re.search(r"Runtime Version: (\S+)", linea)
        if m:
            return m.group(1)
    return "—"


def coste_millon(ms_facturados, mb):
    gbs = ms_facturados / 1000 * mb / 1024 * 1_000_000
    return 1_000_000 * PRECIO_PETICION + gbs * PRECIO_GBS


def escribir():
    datos = cargar_todo()
    cons = B.cargar("construir.json", {})
    limp = B.cargar("limpieza.json")
    vs = [v for v in B.ORDEN if v in datos]
    L = []
    w = L.append

    w("# Pruebas ejecutadas — Go en AWS Lambda de verdad (la revancha del 121)")
    w("")
    w("> Salida **literal**. La generan `banco.py` (construir · medir · memoria · limpiar) e `informe.py`, a partir")
    w("> de `_datos/*.json`. ⛔ Lo que no esté aquí no puede salir en pantalla. El número de cuenta va tachado.")
    w("")
    fin = max((datos[v].get("fin", "") for v in vs), default="")
    w(f"**Dónde:** AWS Lambda, `{B.REGION}`, arquitectura `{B.ARQ}`, **{B.MEMORIA} MB**, timeout {B.TIMEOUT} s. "
      f"Funciones `gl-lambda-*` (todas borradas al final: §11). Última medida: {fin}.")
    w("")
    w("**El mismo handler en todos** (el del 121): `n` vueltas de `i*i % 1000003` y, si se pide, esperar `io_ms`.")
    w("Python y Node son **literalmente** los archivos del 121 (Node suma con `BigInt`, como allí).")
    w("Java gestionado y Java nativo usan el **mismo** `Manejador.java`; el nativo lleva además un bucle del")
    w("Runtime API escrito a mano (`Bootstrap.java`, sin librería). Go usa `aws-lambda-go`, Rust `lambda_runtime`,")
    w(".NET `Amazon.Lambda.RuntimeSupport` con Native AOT.")
    w("")
    w("**Cómo se fuerza un arranque en frío:** antes de cada invocación fría se cambia una variable de entorno de la")
    w("función (`FRIO=1, 2, 3…`); eso retira los entornos viejos. En Java + SnapStart, además, se publica una")
    w("versión nueva por cada arranque (SnapStart solo actúa sobre versiones publicadas). Las cifras de `Init`,")
    w("`Restore`, `Duration` y `Billed Duration` son las de la línea `REPORT` que escribe Lambda.")
    w("")
    w("⚠️ **«Punta a punta»** es lo que tarda `invoke` medido desde esta máquina: **incluye la red de mi casa**")
    w("hasta us-east-1. Sirve para comparar entre lenguajes, no como cifra absoluta.")
    w("")

    # ── 0 · versiones
    w("---")
    w("")
    w("## 0 · Versiones")
    w("")
    w("| variante | runtime de Lambda | versión del runtime (INIT_START) | compilado con |")
    w("|---|---|---|---|")
    for v in vs:
        c = cons.get(B.VARIANTES[v].get("paquete", v), {})
        w(f"| {NOMBRE[v]} | `{datos[v]['funcion']['Runtime']}` | `{version_runtime(datos[v])}` | {('`' + c['imagen'] + '`') if c.get('imagen') else '— (se sube el fuente)'} |")
    w("")
    w("Lo que imprimieron las herramientas al compilar:")
    w("")
    w("```")
    for v in ["go", "rust", "dotnet-aot", "java-native", "java"]:
        if cons.get(v, {}).get("salida"):
            w(f"# {v}")
            for linea in cons[v]["salida"].splitlines():
                if linea.strip():
                    w(B.tachar(linea.rstrip()))
    w("```")
    w("")

    # ── 1 · comprobación
    w("---")
    w("")
    w("## 1 · ¿Devuelven lo mismo? (antes de medir nada)")
    w("")
    w(f"Referencia calculada aparte: suma de `i*i % 1000003` para i = 1…{miles(B.N_CPU)} = **{B.REFERENCIA}**.")
    w("")
    w("| variante | n=200000 → suma | n=1000 → suma | io_ms=100 → io_ms | ¿igual? |")
    w("|---|---:|---:|---:|---|")
    sumas1000 = set()
    for v in vs:
        c = datos[v]["comprobacion"]
        ok = c[0]["respuesta"]["suma"] == B.REFERENCIA and c[2]["respuesta"]["io_ms"] == 100
        sumas1000.add(c[1]["respuesta"]["suma"])
        w(f"| {NOMBRE[v]} | {c[0]['respuesta']['suma']} | {c[1]['respuesta']['suma']} | {c[2]['respuesta']['io_ms']} | {'✅' if ok else '❌'} |")
    w("")
    w(f"> Sumas distintas para n=1000 entre todas las variantes: **{len(sumas1000)}** (1 = todas iguales).")
    w("")
    w("Respuesta literal de cada una (n=200000):")
    w("")
    w("```")
    for v in vs:
        w(json.dumps(datos[v]["comprobacion"][0]["respuesta"], separators=(",", ":")))
    w("```")
    w("")

    # ── 2 · paquete
    w("---")
    w("")
    w("## 2 · Lo que se sube: el .zip")
    w("")
    w("`CodeSize` es lo que devuelve Lambda al crear la función. «Compilar» es el `docker run` entero, en un")
    w("contenedor nuevo, con la imagen ya bajada pero **bajando las dependencias** (no hay caché), como en un CI")
    w("limpio. Dos rondas (`python banco.py cronometrar`, con el candado de la fábrica tomado: nadie más compilaba).")
    w("")
    w("| variante | sin comprimir | .zip (CodeSize) | compilar, ronda 1 | ronda 2 |")
    w("|---|---:|---:|---:|---:|")
    cr = [B.cargar("cronometrar-1.json", {}), B.cargar("cronometrar-2.json", {})]
    for v in vs:
        if v == "java-snap":
            continue
        c = cons.get(v, {})
        segs = [f"{n(r[v]['segundos'])} s" if v in r else "—" for r in cr]
        w(f"| {NOMBRE[v]} | {miles(c.get('sin_comprimir', 0))} B ({tam(c.get('sin_comprimir', 0))}) | "
          f"{miles(datos[v]['funcion']['CodeSize'])} B ({tam(datos[v]['funcion']['CodeSize'])}) | {segs[0]} | {segs[1]} |")
    w("")
    w("> Java + SnapStart sube el mismo .zip que Java (JVM).")
    w("")

    # ── 3 · frío
    w("---")
    w("")
    w(f"## 3 · Arranque en frío — {B.N_FRIOS} entornos nuevos por variante")
    w("")
    w("«Init» es `Init Duration` del REPORT; en SnapStart no hay Init: es `Restore Duration` (restaurar la")
    w("instantánea). «Duración» es la de esa primera invocación (200.000 vueltas). Con 25 muestras, el p99 es el máximo.")
    w("")
    w("| variante | fríos confirmados | init/restore mediana | p90 | p99 | mín | duración 1ª | facturada mediana | init + duración mediana | punta a punta mediana |")
    w("|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|")
    resumen_frio = {}
    for v in vs:
        fr = [r for r in datos[v]["frios"] if arranque(r) is not None]
        a = [arranque(r) for r in fr]
        tot = [arranque(r) + r["duracion"] for r in fr]
        resumen_frio[v] = dict(med=med(a), p90=pct(a, 90), p99=pct(a, 99), mn=min(a), dur=med([r["duracion"] for r in fr]),
                               fact=med([r["facturada"] for r in fr]), tot=med(tot), e2e=med([r["e2e"] for r in fr]),
                               n=len(fr), mem=max(r["memoria"] for r in fr),
                               rfact=med([r["restore_facturada"] for r in fr]) if all("restore_facturada" in r for r in fr) else None)
        R = resumen_frio[v]
        w(f"| {NOMBRE[v]} | {R['n']}/{len(datos[v]['frios'])} | {ms(R['med'])} | {ms(R['p90'])} | {ms(R['p99'])} | {ms(R['mn'])} | "
          f"{ms(R['dur'])} | {ms(R['fact'], 0)} | {ms(R['tot'])} | {ms(R['e2e'], 0)} |")
    w("")
    orden_frio = sorted(resumen_frio, key=lambda v: resumen_frio[v]["tot"])
    w("Ordenadas por «init + duración» (lo que espera la primera petición dentro de AWS):")
    w("")
    w("```")
    for v in orden_frio:
        R = resumen_frio[v]
        w(f"{NOMBRE[v]:<22} {R['tot']:>8.1f} ms   (init/restore {R['med']:.1f} + duración {R['dur']:.1f})")
    w("```")
    w("")
    w("¿Se cobra el arranque? La «facturada» de la invocación fría contra init + duración:")
    w("")
    w("| variante | init/restore mediana | duración mediana | facturada mediana | Billed Restore Duration mediana |")
    w("|---|---:|---:|---:|---:|")
    for v in vs:
        R = resumen_frio[v]
        w(f"| {NOMBRE[v]} | {ms(R['med'])} | {ms(R['dur'])} | {ms(R['fact'], 0)} | {ms(R['rfact'], 0) if R['rfact'] is not None else '—'} |")
    w("")
    rel = B.cargar("medida-rust-relleno.json")
    if rel and "rust" in datos and "go" in datos:
        a_rel = [arranque(r) for r in rel["frios"] if arranque(r) is not None]
        w("**Comprobación aparte: ¿es el tamaño del .zip?** El MISMO binario de Rust, con un archivo de relleno aleatorio")
        w(f"({miles(cons['rust-relleno']['relleno'])} B, incompresible) para que el .zip pese lo que el de Go. {len(a_rel)} arranques en frío:")
        w("")
        w("| variante | .zip (CodeSize) | init mediana | p90 | mín |")
        w("|---|---:|---:|---:|---:|")
        for etiqueta, d_, a_ in (("Rust", datos["rust"], [arranque(r) for r in datos["rust"]["frios"]]),
                                 ("Rust + relleno", rel, a_rel),
                                 ("Go", datos["go"], [arranque(r) for r in datos["go"]["frios"]])):
            w(f"| {etiqueta} | {miles(d_['funcion']['CodeSize'])} B | {ms(med(a_))} | {ms(pct(a_, 90))} | {ms(min(a_))} |")
        w("")
        w("> Con el mismo tamaño de .zip, Rust sigue arrancando como Rust: el peso del paquete NO explica el init de Go.")
        w("> Por qué Go tarda más en inicializarse no se midió aquí.")
        w("")
    raf = {v: B.cargar(f"rafaga-{v}.json") for v in vs}
    raf = {v: r for v, r in raf.items() if r}
    if raf:
        w(f"**Ráfaga: {B.N_RAFAGA} invocaciones a la vez.** Cada una retiene su entorno 1,5 s (`io_ms=1500`), así que casi")
        w("todas necesitan un entorno nuevo, y nacen todos al mismo tiempo. En Java + SnapStart va contra la última versión")
        w("publicada, cuya instantánea ya se había restaurado antes (no recién publicada, como en la tabla de arriba).")
        w("")
        w("| variante | entornos nuevos | init/restore mediana | p90 | máx. | mín. |")
        w("|---|---:|---:|---:|---:|---:|")
        for v, r in raf.items():
            a_ = [arranque(x) for x in r["invocaciones"] if arranque(x) is not None]
            if a_:
                w(f"| {NOMBRE[v]} | {len(a_)}/{len(r['invocaciones'])} | {ms(med(a_))} | {ms(pct(a_, 90))} | {ms(max(a_))} | {ms(min(a_))} |")
        w("")
    w("Una línea REPORT de cada variante, tal cual (la primera invocación fría):")
    w("")
    w("```")
    for v in vs:
        w(f"# {NOMBRE[v]}")
        w(B.tachar(datos[v]["frios"][0]["linea"]).replace("\t", "  "))
    w("```")
    w("")
    w("<details><summary>todas las medidas de arranque, una a una (ms)</summary>")
    w("")
    w("```")
    for v in vs:
        w(f"{v:<12} init/restore  " + "  ".join(n(arranque(r)) if arranque(r) is not None else "¿?" for r in datos[v]["frios"]))
        w(f"{'':<12} punta a punta " + "  ".join(n(r["e2e"], 0) for r in datos[v]["frios"]))
    w("```")
    w("")
    w("</details>")
    w("")

    # ── 4 · caliente CPU
    w("---")
    w("")
    w(f"## 4 · En caliente — {B.N_CALIENTE} invocaciones con 200.000 vueltas de CPU")
    w("")
    w("| variante | duración p50 | duración p99 | facturada p50 | facturada p99 | memoria máx. usada | punta a punta p50 | p99 | invocaciones con arranque |")
    w("|---|---:|---:|---:|---:|---:|---:|---:|---:|")
    resumen_cpu = {}
    for v in vs:
        c = datos[v]["caliente_cpu"]
        d_ = [r["duracion"] for r in c]
        f_ = [r["facturada"] for r in c]
        e_ = [r["e2e"] for r in c]
        resumen_cpu[v] = dict(p50=pct(d_, 50), p99=pct(d_, 99), f50=pct(f_, 50), f99=pct(f_, 99),
                              mem=max(r["memoria"] for r in c), e50=pct(e_, 50), e99=pct(e_, 99))
        R = resumen_cpu[v]
        w(f"| {NOMBRE[v]} | {ms(R['p50'], 2)} | {ms(R['p99'], 2)} | {ms(R['f50'], 0)} | {ms(R['f99'], 0)} | {n(R['mem'], 0)} MB | "
          f"{ms(R['e50'], 0)} | {ms(R['e99'], 0)} | {datos[v]['caliente_cpu_con_arranque']} |")
    w("")
    w("```")
    for v in sorted(resumen_cpu, key=lambda v: resumen_cpu[v]["p50"]):
        w(f"{NOMBRE[v]:<22} p50 {resumen_cpu[v]['p50']:>8.2f} ms   facturada {resumen_cpu[v]['f50']:>5.0f} ms   memoria {resumen_cpu[v]['mem']:.0f} MB")
    w("```")
    w("")

    # ── 5 · E/S
    w("---")
    w("")
    w(f"## 5 · El matiz del 121, en AWS — {B.N_IO} invocaciones esperando 100 ms de E/S (sin CPU)")
    w("")
    w("| variante | duración p50 | duración p99 | facturada p50 | punta a punta p50 |")
    w("|---|---:|---:|---:|---:|")
    p50io = {}
    for v in vs:
        c = datos[v]["caliente_io"]
        d_ = [r["duracion"] for r in c]
        p50io[v] = pct(d_, 50)
        w(f"| {NOMBRE[v]} | {ms(pct(d_, 50), 2)} | {ms(pct(d_, 99), 2)} | {ms(pct([r['facturada'] for r in c], 50), 0)} | "
          f"{ms(pct([r['e2e'] for r in c], 50), 0)} |")
    w("")
    w(f"> Entre la mejor y la peor p50 esperando 100 ms: **{n(max(p50io.values()) - min(p50io.values()), 2)} ms**.")
    w("")

    # ── 6 · memoria
    mems = {v: B.cargar(f"memoria-{v}.json") for v in vs}
    mems = {v: m for v, m in mems.items() if m}
    if mems:
        w("---")
        w("")
        w(f"## 6 · La CPU la reparte la memoria — el mismo trabajo con 128, {B.MEMORIA} y 1024 MB")
        w("")
        w(f"Duración p50 de {B.N_MEMORIA} invocaciones calientes (200.000 vueltas) con cada memoria; la de {B.MEMORIA} MB es la de §4.")
        w("")
        w(f"| variante | 128 MB | {B.MEMORIA} MB | 1024 MB | init a 128 MB | init a 1024 MB |")
        w("|---|---:|---:|---:|---:|---:|")
        for v, m in mems.items():
            p = {mb: pct([r["duracion"] for r in m[mb]["caliente"]], 50) for mb in m}
            i128, i1024 = arranque(m["128"]["frio"]), arranque(m["1024"]["frio"])
            w(f"| {NOMBRE[v]} | {ms(p['128'], 2)} | {ms(resumen_cpu[v]['p50'], 2)} | {ms(p['1024'], 2)} | "
              f"{ms(i128) if i128 else '—'} | {ms(i1024) if i1024 else '—'} |")
        w("")
        w("> «init a 128/1024 MB» es UNA sola medida (el primer arranque tras cambiar la memoria), no una mediana.")
        w("")

    # ── 7 · el 121
    if V121.exists():
        a = json.loads(V121.read_text(encoding="utf-8"))
        w("---")
        w("")
        w("## 7 · El 121 (emulador RIE en mi PC) contra AWS de verdad")
        w("")
        w("Los números del 121 salen de `recursos/pruebas-lambda/SALIDA.json` (8 arranques, 200 invocaciones; Python 3.13 y")
        w("Node 22 allí, Python 3.14 y Node 24 aquí). La duración es la del REPORT en los dos.")
        w("")
        w("| lenguaje | init 121 (emulador) | init AWS | × | caliente p50 121 | caliente p50 AWS | × | facturada p50 121 | facturada p50 AWS |")
        w("|---|---:|---:|---:|---:|---:|---:|---:|---:|")
        for v in ("go", "python", "node"):
            if v not in datos:
                continue
            i121 = med(a["frio"][v]["init_ms"])
            c121 = pct(a["templado"][v]["cpu_ms"], 50)
            f121 = pct(a["templado"][v]["facturada_ms"], 50)
            w(f"| {NOMBRE[v]} | {ms(i121)} | {ms(resumen_frio[v]['med'])} | {n(resumen_frio[v]['med'] / i121)} | "
              f"{ms(c121, 2)} | {ms(resumen_cpu[v]['p50'], 2)} | {n(resumen_cpu[v]['p50'] / c121)} | {ms(f121, 2)} | {ms(resumen_cpu[v]['f50'], 0)} |")
        w("")
        w("> ⚠️ «caliente p50 121» es lo que el 121 llamó «p50» (cronometrado desde el cliente, en local); la facturada del 121")
        w("> era la duración del REPORT del emulador. Aquí la duración es la del REPORT de Lambda con 512 MB.")
        w("")
        tres = [v for v in ("go", "python", "node") if v in datos]
        orden_121 = sorted(tres, key=lambda v: med(a["frio"][v]["init_ms"]))
        orden_aws = sorted(tres, key=lambda v: resumen_frio[v]["med"])
        w(f"> Orden de arranque en el 121: {' < '.join(NOMBRE[v] for v in orden_121)}. En AWS: {' < '.join(NOMBRE[v] for v in orden_aws)}.")
        w("")

    # ── 8 · factura
    w("---")
    w("")
    w("## 8 · La factura")
    w("")
    w("⛔ Los precios **no se miden: se copian de AWS**. Lo nuestro es la aritmética y la duración facturada medida arriba.")
    w("")
    w(f"**Fuente:** {FUENTE_PRECIOS}")
    w("")
    w("- 0.20 $ por millón de peticiones")
    w("- 0.0000166667 $ por GB-segundo (x86)")
    w("- capa gratuita al mes: 1.000.000 de peticiones y 400.000 GB-s (aquí NO se descuenta)")
    w("- SnapStart: «SnapStart pricing does not apply to supported Java managed runtimes»")
    w("")
    w(f"Un millón de invocaciones **en caliente** con {B.MEMORIA} MB, con la facturada p50 de §4; y lo que cuesta un millón")
    w("de **arranques en frío**, con la facturada mediana de §3 (que ya incluye el init cuando Lambda lo cobra):")
    w("")
    w("| variante | facturada p50 caliente | $ por millón en caliente | facturada fría mediana | $ por millón de arranques en frío |")
    w("|---|---:|---:|---:|---:|")
    for v in sorted(vs, key=lambda v: resumen_cpu[v]["f50"]):
        fc, ff = resumen_cpu[v]["f50"], resumen_frio[v]["fact"]
        w(f"| {NOMBRE[v]} | {ms(fc, 0)} | {n(coste_millon(fc, B.MEMORIA), 2)} $ | {ms(ff, 0)} | {n(coste_millon(ff, B.MEMORIA), 2)} $ |")
    w("")
    w(f"> Solo peticiones, sin cómputo: {n(1_000_000 * PRECIO_PETICION, 2)} $ por millón.")
    w("")
    if mems:
        w("La misma cuenta en caliente con 128 y 1024 MB (facturada p50 de §6):")
        w("")
        w(f"| variante | 128 MB: facturada p50 | $ / millón | {B.MEMORIA} MB: $ / millón | 1024 MB: facturada p50 | $ / millón |")
        w("|---|---:|---:|---:|---:|---:|")
        for v, m in mems.items():
            f128 = pct([r["facturada"] for r in m["128"]["caliente"]], 50)
            f1024 = pct([r["facturada"] for r in m["1024"]["caliente"]], 50)
            w(f"| {NOMBRE[v]} | {ms(f128, 0)} | {n(coste_millon(f128, 128), 2)} $ | {n(coste_millon(resumen_cpu[v]['f50'], B.MEMORIA), 2)} $ | "
              f"{ms(f1024, 0)} | {n(coste_millon(f1024, 1024), 2)} $ |")
        w("")
    if V121.exists() and mems:
        a = json.loads(V121.read_text(encoding="utf-8"))
        w("**La factura del 121, rehecha con lo que factura AWS de verdad.** El 121 calculó con la duración del emulador en mi PC.")
        w("Mismos precios, mismas invocaciones; aquí con la facturada p50 medida en Lambda con esa memoria:")
        w("")
        w("| lenguaje | invocaciones/mes | memoria | facturada 121 | total/mes 121 | facturada AWS | total/mes AWS |")
        w("|---|---:|---:|---:|---:|---:|---:|")
        totales = {}
        for f in a["factura"]:
            v, mb, inv = f["lang"], f["mb"], f["invocaciones"]
            if v not in mems or str(mb) not in mems[v]:
                continue
            fa = pct([r["facturada"] for r in mems[v][str(mb)]["caliente"]], 50)
            gbs = math.ceil(fa) / 1000 * mb / 1024 * inv
            tot = max(0, inv - 1_000_000) * PRECIO_PETICION + max(0, gbs - 400_000) * PRECIO_GBS  # como el 121: con capa gratuita
            totales[(v, mb, inv)] = (f["total"], tot)
            w(f"| {NOMBRE[v]} | {miles(inv)} | {mb} MB | {f['ms_facturados']} ms | {n(f['total'], 2)} $ | {ms(fa, 0)} | {n(tot, 2)} $ |")
        w("")
        w("> Esta tabla, como la del 121, SÍ descuenta la capa gratuita (1 M de peticiones y 400.000 GB-s al mes).")
        for (v, mb, inv) in [k for k in totales if k[0] == "python"]:
            if ("go", mb, inv) in totales:
                g, p_ = totales[("go", mb, inv)], totales[(v, mb, inv)]
                w(f"> {miles(inv)} invocaciones a {mb} MB, Python menos Go: el 121 decía **{n(p_[0] - g[0], 2)} $**; con AWS de verdad, **{n(p_[1] - g[1], 2)} $**.")
        w("")
    w("")

    # ── 9 · bloque sin Go
    w("---")
    w("")
    w("## 9 · Solo Python, Node, Java y .NET (para la versión de Cultura DevOps, sin Go)")
    w("")
    w("Mismas medidas que arriba, separadas. Nada nuevo: es un subconjunto de §2-§8.")
    w("")
    w("| variante | .zip | init/restore mediana | p90 | p99 | init + duración | caliente p50 | caliente p99 | facturada p50 | memoria | $ / millón caliente |")
    w("|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|")
    for v in [x for x in SIN_GO if x in vs]:
        F, C_ = resumen_frio[v], resumen_cpu[v]
        w(f"| {NOMBRE[v]} | {tam(datos[v]['funcion']['CodeSize'])} | {ms(F['med'])} | {ms(F['p90'])} | {ms(F['p99'])} | "
          f"{ms(F['tot'])} | {ms(C_['p50'], 2)} | {ms(C_['p99'], 2)} | {ms(C_['f50'], 0)} | {n(C_['mem'], 0)} MB | "
          f"{n(coste_millon(C_['f50'], B.MEMORIA), 2)} $ |")
    w("")

    # ── 10 · lo que no se midió
    w("---")
    w("")
    w("## 10 · Lo que NO se midió")
    w("")
    w("- arm64 (Graviton): todo es x86_64.")
    w("- SnapStart con «priming» (ejecutar código antes de la instantánea): el Java + SnapStart de aquí es el handler tal cual.")
    w("- Imágenes de contenedor: todo se subió como .zip.")
    w("- Concurrencia dentro de un mismo entorno: salvo la ráfaga de §3, las invocaciones de cada función van de una en una.")
    w("- Llamadas reales a otros servicios (DynamoDB, S3): la E/S es una espera de 100 ms.")
    w("- Los primeros arranques tras publicar una versión de SnapStart pueden ir más lentos mientras AWS calienta su caché de")
    w("  instantáneas; aquí cada arranque es de una versión recién publicada, o sea, el peor caso.")
    w("")

    # ── 11 · limpieza
    w("---")
    w("")
    w("## 11 · Limpieza: no queda nada `gl-lambda-*` en la cuenta")
    w("")
    if limp:
        w(f"Hecho el {limp['cuando']} por `python banco.py limpiar`:")
        w("")
        w("```")
        for h in limp["hecho"]:
            w(h)
        w("```")
        w("")
        w("Comprobación posterior, mirando donde vive cada cosa:")
        w("")
        w("```")
        w(f"lambda list-functions (nombres que empiezan por gl-lambda-) → {len(limp['quedan_funciones'])} {limp['quedan_funciones']}")
        w(f"iam get-role gl-lambda-rol → {limp['rol_get_role']}")
        w(f"iam list-roles (nombres que empiezan por gl-lambda-) → {len(limp['quedan_roles'])} {limp['quedan_roles']}")
        w(f"logs describe-log-groups --log-group-name-prefix /aws/lambda/gl-lambda- → {len(limp['quedan_log_groups'])} {limp['quedan_log_groups']}")
        w(f"fallos al borrar → {len(limp['fallos'])} {limp['fallos']}")
        w("```")
        for v2 in limp.get("verificaciones", []):
            w("")
            w(f"Comprobación repetida el {v2['cuando']} (`python banco.py verificar`, sin borrar nada):")
            w("")
            w("```")
            w(f"lambda list-functions (gl-lambda-*) → {len(v2['quedan_funciones'])} {v2['quedan_funciones']}")
            w(f"iam get-role gl-lambda-rol → {v2['rol_get_role']}")
            w(f"iam list-roles (gl-lambda-*) → {len(v2['quedan_roles'])} {v2['quedan_roles']}")
            w(f"logs describe-log-groups /aws/lambda/gl-lambda- → {len(v2['quedan_log_groups'])} {v2['quedan_log_groups']}")
            w("```")
    else:
        w("⚠️ todavía no se ha ejecutado `python banco.py limpiar`.")
    w("")

    texto = B.tachar("\n".join(L))
    (AQUI / "SALIDA.md").write_text(texto, encoding="utf-8")
    print(f"SALIDA.md: {len(L)} líneas")


def tam(b):
    return f"{b / 1048576:.2f} MiB" if b >= 1048576 else f"{b / 1024:.1f} KiB"


def ms(x, d=1):
    return f"{x:.{d}f} ms"
