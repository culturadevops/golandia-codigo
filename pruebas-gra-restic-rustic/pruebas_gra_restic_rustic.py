# -*- coding: utf-8 -*-
"""pruebas_gra_restic_rustic.py — golandia 959, «restic (Go) contra rustic (Rust)».

Los binarios OFICIALES (restic 0.20.0 de restic/restic:0.20.0, compilado con go1.27.4; rustic v0.12.0 de
ghcr.io/rustic-rs/rustic:v0.12.0) en la MISMA imagen gl-gra-backup:1 (debian:trixie-slim + GNU time), con --cpus 2 y
--memory 1g. La carpeta de partida es el volumen gl-gra-base (900 MB, 15.667 archivos: el árbol de Go 1.27 entero,
4 binarios aleatorios de 64 MiB y 4 logs de texto). Cada herramienta, en volúmenes propios recién creados:
init → backup inicial → backup sin cambios → el MISMO cambio (1 de cada 50 archivos +1 línea, 1 de cada 97 borrado, un
archivo nuevo de 32 MiB) → backup incremental → restaurar → check → check --read-data. Lo restaurado se compara con
`diff -r` contra la carpeta: si no da 0 diferencias, la corrida se para (¿llegaron los dos al mismo estado?).
La compatibilidad de formato (compat.sh) se prueba aparte y va al principio de la SALIDA.

    python investigacion/pruebas_gra_restic_rustic.py <corrida>  → recursos/pruebas-gra-restic-rustic/SALIDA-corrida-<N>.md
"""
import re
import sys
from datetime import datetime
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
import pruebas_gra_comun as K  # noqa: E402

BASE = K.REC / "pruebas-gra-restic-rustic"
IMG = "gl-gra-backup:1"
BANCO_W = str(BASE).replace("/", "\\")


def vol(*ns):
    for n in ns:
        K.sh(["docker", "volume", "rm", "-f", n], timeout=120)
        K.sh(["docker", "volume", "create", n], timeout=60, check=True)


def correr(herr):
    vol(f"{K.P}-datos", f"{K.P}-repo", f"{K.P}-rest")
    r = K.sh(["docker", "run", "--name", f"{K.P}-{herr}", "--cpus", "2", "--memory", "1g", "--memory-swap", "1g",
              "-v", f"{K.P}-base:/base:ro", "-v", f"{K.P}-datos:/datos", "-v", f"{K.P}-repo:/repo",
              "-v", f"{K.P}-rest:/rest", "-v", f"{BANCO_W}:/banco:ro", IMG, "bash", "/banco/medir.sh", herr], timeout=2400)
    return r.stdout + r.stderr


def pasos(txt):
    d = {}
    for m in re.finditer(r"PASO (\w+) seg=([\d.]+) rss_kb=(\d+) user=([\d.]+) sys=([\d.]+) rc=(\d+)", txt):
        d[m.group(1)] = dict(seg=float(m.group(2)), rss=int(m.group(3)) / 1024, cpu=float(m.group(4)) + float(m.group(5)),
                             rc=int(m.group(6)))
    for m in re.finditer(r"REPO (\w+) bytes=(\d+) archivos=(\d+)", txt):
        d["repo_" + m.group(1)] = int(m.group(2)) / 2**20
    m = re.search(r"RESTAURADO archivos=(\d+) bytes=(\d+) diferencias=(\d+)", txt)
    d["rest"] = tuple(int(x) for x in m.groups()) if m else None
    m = re.search(r"ORIGEN archivos=(\d+) bytes=(\d+)", txt)
    d["origen"] = tuple(int(x) for x in m.groups()) if m else None
    return d


def main():
    n = int(sys.argv[1])
    out = K.Salida(BASE / f"SALIDA-corrida-{n}.md")
    orden = ["restic", "rustic"] if n % 2 else ["rustic", "restic"]
    with K.candado(f"restic-rustic corrida {n}"):
        out.w(f"# restic (Go) contra rustic (Rust) · corrida {n}\n")
        out.w(f"> Corrida del {datetime.now():%Y-%m-%d %H:%M}, con el candado `_logs/banco.lock` tomado y sin `fab producir` "
              f"vivo. Regenerable con `python investigacion/pruebas_gra_restic_rustic.py {n}`. Orden: {' → '.join(orden)}.\n")
        v = K.sh(["docker", "run", "--name", f"{K.P}-ver", IMG, "sh", "-c", "restic version; rustic --version"], timeout=120)
        out.w("## 0 · Versiones (binarios oficiales de sus imágenes)\n")
        for l in v.stdout.strip().splitlines():
            out.w(f"- `{l}`")
        out.w(f"- contenedores ajenos vivos al empezar: {K.ajenos() or 'ninguno'} · CPU de Windows: {K.cpu_windows()}\n")

        out.w("## 1 · ¿Leen y escriben el mismo formato? (compat.sh, sobre go/src/net: 461 archivos)\n")
        vol(f"{K.P}-repo", f"{K.P}-rest")
        c = K.sh(["docker", "run", "--name", f"{K.P}-compat", "--cpus", "2", "--memory", "1g", "-v", f"{K.P}-base:/base:ro",
                  "-v", f"{K.P}-repo:/repo", "-v", f"{K.P}-rest:/rest", "-v", f"{BANCO_W}:/banco:ro", IMG, "bash",
                  "/banco/compat.sh"], timeout=900)
        out.w("```")
        for l in (c.stdout + c.stderr).strip().splitlines():
            out.w(l)
        out.w("```\n")
        if re.search(r"rc=[1-9]", c.stdout) or re.search(r"diferencias[^:]*: [1-9]", c.stdout):
            out.w("⛔ la compatibilidad falló: se para.")
            raise SystemExit(1)

        res = {}
        for h in orden:
            out.w(f"## 2 · {h}: salida cruda de medir.sh\n")
            t = correr(h)
            out.w("```")
            for l in t.strip().splitlines():
                out.w(l)
            out.w("```\n")
            res[h] = pasos(t)
            d = res[h]
            if any(p.get("rc") for k, p in d.items() if isinstance(p, dict)):
                out.w(f"⛔ {h}: algún paso salió con rc≠0: se para.")
                raise SystemExit(1)
            if not d["rest"] or d["rest"][2] != 0 or d["rest"][0] != d["origen"][0]:
                out.w(f"⛔ {h}: lo restaurado NO coincide con la carpeta ({d['rest']} contra {d['origen']}): se para.")
                raise SystemExit(1)

        out.w("## 3 · ¿Llegaron los dos al mismo estado?\n")
        a, b = res["restic"], res["rustic"]
        out.w(f"- carpeta tras el cambio: {a['origen'][0]} archivos, {a['origen'][1]} bytes (restic) · {b['origen'][0]} archivos, "
              f"{b['origen'][1]} bytes (rustic)")
        out.w(f"- restaurado: restic {a['rest'][0]} archivos / {a['rest'][1]} bytes / {a['rest'][2]} diferencias · rustic "
              f"{b['rest'][0]} / {b['rest'][1]} / {b['rest'][2]}")
        if a["origen"] != b["origen"] or a["rest"][:2] != b["rest"][:2]:
            out.w("⛔ NO: se para.")
            raise SystemExit(1)
        out.w("- ✅ sí: misma carpeta de partida, mismo cambio y restauración idéntica (diff -r = 0) en los dos.\n")

        out.w("## 4 · La tabla (2 CPU, 1 GiB; seg = reloj; CPU = user+sys; RAM = máximo residente del proceso)\n")
        out.w("| paso | restic seg | rustic seg | restic CPU s | rustic CPU s | restic RAM MiB | rustic RAM MiB |")
        out.w("|---|---:|---:|---:|---:|---:|---:|")
        for p in ("init", "inicial", "sin_cambios", "incremental", "restaurar", "check", "check_datos"):
            x, y = a[p], b[p]
            out.w(f"| {p} | {x['seg']:.2f} | {y['seg']:.2f} | {x['cpu']:.2f} | {y['cpu']:.2f} | {x['rss']:.1f} | {y['rss']:.1f} |")
        out.w("")
        out.w("| repositorio | restic MiB | rustic MiB |")
        out.w("|---|---:|---:|")
        out.w(f"| tras el inicial | {a['repo_inicial']:.1f} | {b['repo_inicial']:.1f} |")
        out.w(f"| tras el incremental | {a['repo_incremental']:.1f} | {b['repo_incremental']:.1f} |")
        out.w(f"| carpeta de partida | {891834621 / 2**20:.1f} | (la misma) |")
        out.w(f"\n> Terminó {datetime.now():%Y-%m-%d %H:%M}. CPU de Windows al final: {K.cpu_windows()}.")


if __name__ == "__main__":
    main()
