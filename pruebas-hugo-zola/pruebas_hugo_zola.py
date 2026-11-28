# -*- coding: utf-8 -*-
"""pruebas_hugo_zola.py — «Go contra Rust: herramientas populares, lo medí» · Hugo (Go) contra Zola (Rust).

El MISMO sitio (recursos/pruebas-hugo-zola/generar_sitio.py, semilla fija): N páginas Markdown con 4 secciones, lista,
código, tabla, 3 etiquetas de 20, y plantillas equivalentes. Se mide:
  · ¿generan lo mismo? (mismos archivos, mismo texto en una muestra de páginas; si no, se para);
  · build completo: la primera (sin public/) y repetida (con public/), mediana de 5, tiempo, RAM máx. y CPU;
  · servidor de desarrollo: tiempo hasta servir, RAM, y editar una página / una plantilla hasta verla servida.
Binarios oficiales de cada release (Dockerfile). Contenedor con --cpus 2 y --memory 1g, prefijo gl-grb-.

    python investigacion/pruebas_hugo_zola.py <corrida>   → recursos/pruebas-hugo-zola/SALIDA-corrida-<corrida>.md
"""
import json
import statistics
import sys
from datetime import datetime
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
import grb_comun as G  # noqa: E402

BASE = G.REC / "pruebas-hugo-zola"
IMG = "gl-grb-hugo:1"
VOL = "gl-grb-hugo-datos"
TAMANOS = [10000, 30000]
TAMANOS_SERVE = [10000]
REPS = 5
EDICIONES = 10
LIM = ["--cpus", "2", "--memory", "1g"]


def correr(args, timeout, check=True, mem="1g"):
    lim = ["--cpus", "2", "--memory", mem]
    return G.docker_run(lim + ["-v", f"{VOL}:/datos:ro", "-v", f"{BASE}:/s:ro", IMG] + args, timeout=timeout, check=check)


def jsonl(r):
    return [json.loads(l) for l in r.stdout.splitlines() if l.startswith("{")]


def miles(n):
    return f"{n:,}".replace(",", ".")


def med(xs):
    xs = [x for x in xs if x is not None]
    return statistics.median(xs) if xs else None


def f2(x):
    return "-" if x is None else f"{x:.2f}"


def main():
    corrida = sys.argv[1]
    S = G.Salida(BASE / f"SALIDA-corrida-{corrida}.md")
    S.w(f"# Hugo (Go) contra Zola (Rust) · corrida {corrida}\n")
    S.w(f"> Corrida del {datetime.now():%Y-%m-%d %H:%M}. Cada bloque, con el candado `_logs/banco.lock` tomado. "
        f"Regenerable con `python investigacion/pruebas_hugo_zola.py {corrida}`.\n")
    S.w(f"- Máquina (docker info): {G.host()}")
    S.w(f"- Contenedor: `--cpus 2 --memory 1g` (y 4g donde se dice), imagen {IMG} (debian:trixie-slim + binarios oficiales)\n")
    with G.candado("hugo-zola: versiones"):
        r = correr(["sh", "-c", "hugo version; zola --version; ls -l /usr/local/bin/hugo /usr/local/bin/zola; cat /zola-tar.sha256; nproc; "
                    "for n in " + " ".join(map(str, TAMANOS)) + "; do echo \"$n: $(ls /datos/sitio-$n/hugo/content/posts | wc -l) md hugo, "
                    "$(ls /datos/sitio-$n/zola/content/posts | wc -l) md zola, $(du -sh /datos/sitio-$n/hugo/content | cut -f1)\"; done"], 300)
    S.w("## 0 · Versiones, binarios y sitios\n")
    S.w("```\n" + r.stdout.strip() + "\n```\n")

    S.w("## A · ¿Generan lo mismo? (10.000 con 1 GiB; 30.000 con 4 GiB, porque con 1 GiB Hugo no termina: ver B)\n")
    for n, mem in ((10000, "1g"), (30000, "4g")):
        with G.candado(f"hugo-zola: iguales {n}"):
            r = correr(["python3", "/s/medir_sitio.py", "iguales", f"/datos/sitio-{n}"], 1800, mem=mem)
        d = jsonl(r)[0]
        rh, rz = d["res"]["hugo"], d["res"]["zola"]
        S.w(f"- {miles(n)} páginas (--memory {mem}): hugo rc {rh['rc']} · {miles(rh['html'])} HTML · {miles(d['archivos_hugo'])} archivos | "
            f"zola rc {rz['rc']} · {miles(rz['html'])} HTML · {miles(d['archivos_zola'])} archivos")
        S.w(f"  - solo en hugo: {d['solo_hugo']} · solo en zola: {d['solo_zola']}")
        S.w(f"  - texto de {d['muestra']} páginas comparado (sin etiquetas HTML): distinto en {len(d['texto_distinto'])} {d['texto_distinto']}")
        S.w(f"  - ejemplo: «{d['ejemplo']}…»")
        # solo cuentan las páginas HTML: Zola parte el sitemap en sitemap1/2.xml pasadas 30.000 URL, y eso no es otra página
        html_distinto = [x for x in d["solo_hugo"] + d["solo_zola"] if x.endswith(".html")]
        if d["solo_hugo"] or d["solo_zola"]:
            S.w(f"  - archivos que no son HTML y no coinciden (no paran el banco): "
                f"{[x for x in d['solo_hugo'] + d['solo_zola'] if not x.endswith('.html')]}")
        if rh["rc"] or rz["rc"] or html_distinto or d["texto_distinto"] or rh["html"] != rz["html"]:
            S.w(f"\n⛔ No generan lo mismo ({rh.get('err')} {rz.get('err')}). Paro.")
            sys.exit(1)
    S.w()

    S.w("## B · Build completo («primera» = sin public/, «repetida» = otra vez con public/ ya escrito)\n")
    S.w("Variantes de Hugo: HUGO_MEMORYLIMIT (GB; la documentación dice que por defecto es el 25 % de la memoria del sistema, "
        "y dentro del contenedor el sistema son 16 GB) y GOMEMLIMIT (el límite blando del recolector de Go).\n")
    S.w("| páginas | memoria | herramienta | build | reps | s (mediana) | s (mín–máx) | RAM máx. MiB | CPU s (u+s) | HTML | rc |")
    S.w("|---|---|---|---|---|---|---|---|---|---|---|")
    PLAN = [
        (10000, "1g", REPS, "hugo,zola", True, 900),
        (10000, "1g", 3, "hugo:HUGO_MEMORYLIMIT=1,hugo:GOMEMLIMIT=700MiB", False, 900),
        (30000, "1g", 1, "hugo", False, 900),
        (30000, "1g", 3, "zola,hugo:HUGO_MEMORYLIMIT=1,hugo:GOMEMLIMIT=700MiB", False, 900),
        (10000, "4g", 3, "hugo,zola", True, 900),
        (30000, "4g", 3, "hugo,zola", True, 900),
    ]
    for n, mem, reps, herrs, rep, tope in PLAN:
        with G.candado(f"hugo-zola: build {n} {mem} {herrs}"):
            r = correr(["python3", "/s/medir_sitio.py", "build", f"/datos/sitio-{n}", str(reps), herrs,
                        "repetida" if rep else "-", str(tope)], reps * len(herrs.split(",")) * tope * 2 + 600, check=False, mem=mem)
        filas = jsonl(r)
        if r.returncode:
            S.w(f"| {miles(n)} | {mem} | ⚠️ rc del contenedor {r.returncode} · {r.stderr[-200:]!r} |")
        for h in sorted({f["herr"] for f in filas}, key=lambda x: (not x.startswith("hugo"), x)):
            for que in ("primera", "repetida"):
                xs = [f for f in filas if f["herr"] == h and f["que"] == que]
                if not xs:
                    continue
                seg = [f["seg"] for f in xs if f["seg"] is not None]
                S.w(f"| {miles(n)} | {mem} | {h} | {que} | {len(xs)} | {f2(med(seg))} | {f2(min(seg) if seg else None)}–{f2(max(seg) if seg else None)} | "
                    f"{med([f['rss_mib'] for f in xs])} | {med([f['cpu'] for f in xs])} | {sorted({f['html'] for f in xs})} | "
                    f"{sorted({str(f['rc']) for f in xs})} |")
    S.w()

    S.w(f"## C · Servidor de desarrollo con 1 GiB (hugo server / zola serve): listo, RAM y {EDICIONES} ediciones de página + 3 de plantilla\n")
    S.w("Tiempo medido desde fuera: desde que se guarda el archivo hasta que la página servida por HTTP ya trae el cambio "
        "(sondeo cada 5 ms). Hugo con sus opciones por defecto (Fast Render Mode activo).\n")
    S.w("| páginas | herramienta | listo para servir (s) | RAM al estar listo MiB | editar página s (mediana · mín–máx) | "
        "editar plantilla s (mediana · mín–máx) | RAM al final MiB (máx.) |")
    S.w("|---|---|---|---|---|---|---|")
    for n in TAMANOS_SERVE:
        with G.candado(f"hugo-zola: serve {n}"):
            r = correr(["python3", "/s/medir_sitio.py", "serve", f"/datos/sitio-{n}", str(EDICIONES)], 3600, check=False)
        filas = jsonl(r)
        if r.returncode:
            S.w(f"\n⚠️ {n}: rc {r.returncode} · {r.stderr[-600:]}\n")
        for h in ("hugo", "zola"):
            li = next((f for f in filas if f["herr"] == h and f["que"] == "listo"), {})
            fin = next((f for f in filas if f["herr"] == h and f["que"] == "memoria-al-final"), {})
            ep = [f["seg"] for f in filas if f["herr"] == h and f["que"] == "editar-pagina"]
            et = [f["seg"] for f in filas if f["herr"] == h and f["que"] == "editar-plantilla"]
            okp, okt = [x for x in ep if x is not None], [x for x in et if x is not None]
            S.w(f"| {miles(n)} | {h} | {li.get('seg')} | {li.get('VmRSS')} | "
                f"{f2(med(okp))} · {f2(min(okp) if okp else None)}–{f2(max(okp) if okp else None)} ({len(okp)}/{len(ep)}) | "
                f"{f2(med(okt))} · {f2(min(okt) if okt else None)}–{f2(max(okt) if okt else None)} ({len(okt)}/{len(et)}) | "
                f"{fin.get('VmRSS')} ({fin.get('VmHWM')}) |")
        S.w("\n```")
        for f in filas:
            S.w(json.dumps(f, ensure_ascii=False))
        S.w("```\n")
    S.w("FIN")


if __name__ == "__main__":
    main()
