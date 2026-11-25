# -*- coding: utf-8 -*-
"""Regenera SALIDA.md del vídeo 714 (golandia, «5 cambios en Go que están pasando ahora (2026)»).

    python canales/golandia/recursos/cinco-cambios-go-2026/ejecutar_todo.py

Lo nuevo del 714 (go fix sobre un servicio viejo, go.mod 1.27 contra 1.21, //go:fix inline) se ejecuta aquí, en
golang:1.27.2 (--rm, --cpus 2, nombre v714-demos, montaje de solo lectura; go fix reescribe solo una copia en mktemp).
El resto del vídeo reutiliza SALIDA.md ya ejecutados: 700 (novedades-go-1-27), 703 (pruebas-ia-escribe-go-viejo),
705 (pruebas-ia-goroutines-que-no-terminan) y 545 (pruebas-json-v2). Este script copia sus cifras con la ruta de origen.
"""
import collections
import datetime
import json
import re
import subprocess
from pathlib import Path

AQUI = Path(__file__).resolve().parent
REC = AQUI.parent
IMAGEN = "golang:1.27.2"


def docker():
    cmd = ["docker", "run", "--rm", "--cpus", "2", "--name", "v714-demos",
           "-v", f"{AQUI}:/trabajo:ro", IMAGEN, "bash", "/trabajo/correr.sh"]
    r = subprocess.run(cmd, capture_output=True, text=True, encoding="utf-8", errors="replace")
    return r.stdout + r.stderr


def por_analizador(texto, marca):
    """Cuenta los avisos de go vet -vettool=fix -json por analizador."""
    m = re.search(rf"### {marca}\n(.*?)### FIN-JSON", texto, re.S)
    cuenta = collections.Counter()
    if not m:
        return cuenta
    bloque = re.sub(r"^#.*$", "", m.group(1), flags=re.M)
    dec, i = json.JSONDecoder(), 0
    while True:
        j = bloque.find("{", i)
        if j < 0:
            break
        obj, i = dec.raw_decode(bloque, j)
        for paquete in obj.values():
            for analizador, avisos in paquete.items():
                if isinstance(avisos, list):
                    cuenta[analizador] += len(avisos)
    return cuenta


def sin_json(texto):
    return re.sub(r"### JSON-\d+\n.*?### FIN-JSON\n", "", texto, flags=re.S)


def reutilizado():
    filas = []
    for nombre, patron in [
        ("novedades-go-1-27/SALIDA.md", r"## 01-metodos-genericos · go1\.27\.2.*?\[código de salida: 0\]"),
        ("novedades-go-1-27/SALIDA.md", r"## 01-metodos-genericos · go1\.26\.10.*?\[código de salida: 1\]"),
        ("novedades-go-1-27/SALIDA.md", r"## Go 1\.27 con «go 1\.26» en go\.mod.*?\[código de salida: 1\]"),
        ("novedades-go-1-27/SALIDA.md", r"## 06-goroutineleak · go1\.27\.2.*?\[código de salida: 2\]"),
        ("novedades-go-1-27/SALIDA.md", r"## 06-goroutineleak · go1\.26\.10.*?\[código de salida: 0\]"),
        ("novedades-go-1-27/SALIDA.md", r"\| BenchmarkGzipDefault.*?\n"),
        ("pruebas-ia-escribe-go-viejo/SALIDA.md", r"## Tabla principal.*?\*sobre las que compilan\."),
        ("pruebas-ia-escribe-go-viejo/SALIDA.md", r"## ¿Cambió algo la frase.*?byte a byte a su A\.\n- claude-code.*?\n"),
        ("pruebas-ia-goroutines-que-no-terminan/SALIDA.md", r"## Tabla principal.*?\*sobre las que compilan\."),
        ("pruebas-ia-goroutines-que-no-terminan/SALIDA.md", r"goroutineleak profile: total 2\n1 @ 0x4866ea 0x41531c.*?ElMasRapido\.func2\+0x24\t\t/w/m/m0138/tarea\.go:20"),
        ("pruebas-json-v2/SALIDA.md", r"1 · Slice y map sin inicializar.*?member name \"nombre\"\n"),
        ("pruebas-json-v2/SALIDA.md", r"\$ GOEXPERIMENT=nojsonv2 go test -bench.*?BenchmarkUnmarshal_v2API.*?\n"),
    ]:
        texto = (REC / nombre).read_text(encoding="utf-8")
        m = re.search(patron, texto, re.S)
        filas += [f"\n### De `recursos/{nombre}`\n", "```", (m.group(0).strip() if m else "⛔ NO ENCONTRADO"), "```"]
    return filas


def main():
    ahora = datetime.datetime.now().strftime("%Y-%m-%d %H:%M")
    t = docker()
    c127, c121 = por_analizador(t, "JSON-127"), por_analizador(t, "JSON-121")
    partes = [
        "# SALIDA — 714 «5 cambios en Go que están pasando ahora (2026)»\n",
        f"> Corrida del {ahora} con `python ejecutar_todo.py`. Lo nuevo, en `{IMAGEN}` (Docker, --cpus 2). "
        "Lo reutilizado se copia abajo de los SALIDA.md de 700, 703, 705 y 545, con su ruta.\n",
        "\n# Lo nuevo del 714 · golang:1.27.2\n", "```", sin_json(t).rstrip(), "```",
        "\n# Avisos de go fix por analizador (go vet -vettool=fix -json)\n",
        f"- go.mod go 1.27: {sum(c127.values())} avisos en {len(c127)} analizadores: {dict(sorted(c127.items()))}",
        f"- go.mod go 1.21: {sum(c121.values())} avisos en {len(c121)} analizadores: {dict(sorted(c121.items()))}",
        f"- Solo con go 1.27: {sorted(set(c127) - set(c121))}",
        "\n# Fuentes con fecha\n",
        "- go.dev/blog/gofix (Alan Donovan, 17-feb-2026): copia en recursos/pruebas-ia-escribe-go-viejo/fuente-go-blog-gofix.txt.",
        "- go.dev/doc/go1.27 (bajada el 2026-10-29): recursos/novedades-go-1-27/fuente-go1.27.txt.",
        "\n# Reutilizado (copias literales)\n",
    ] + reutilizado()
    (AQUI / "SALIDA.md").write_text("\n".join(partes) + "\n", encoding="utf-8")
    print(f"SALIDA.md escrito · 1.27: {dict(c127)} · 1.21: {dict(c121)}")


if __name__ == "__main__":
    main()
