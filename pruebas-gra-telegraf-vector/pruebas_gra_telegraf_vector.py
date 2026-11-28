# -*- coding: utf-8 -*-
"""pruebas_gra_telegraf_vector.py — golandia 960, «Telegraf (Go) contra Vector (Rust)».

Los binarios OFICIALES (Telegraf 1.40.4 de telegraf:1.40.4; Vector 0.62.0 de timberio/vector:0.62.0-debian) en la MISMA
imagen gl-gra-logs:1, con --cpus 2 y --memory 1g. Cada uno corre como servicio, leyendo un archivo desde el principio
(inputs.tail / source file) y escribiendo JSON a un archivo; medir.sh mira la salida cada 0,25 s y lee CPU y memoria en
/proc. Dos pipelines con 1.000.000 de líneas (volumen gl-gra-logsbase, generar.sh, deterministas):
  A · logs JSON → quitar nivel=debug (1 de cada 10) → añadir entorno=prod → renombrar ms → 900.000 líneas;
  B · log de acceso de nginx en texto → separar los campos (grok COMBINED_LOG_FORMAT / parse_nginx_log combined).
Antes de comparar: la huella (md5 de las columnas ordenadas) de la salida de los dos tiene que ser IGUAL; si no, se para.
Extra: Telegraf con max_undelivered_lines por defecto (100.000 líneas, gl-gra-logshumo) y la memoria en reposo.

    python investigacion/pruebas_gra_telegraf_vector.py <corrida>  → recursos/pruebas-gra-telegraf-vector/SALIDA-corrida-<N>.md
"""
import re
import sys
from datetime import datetime
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
import pruebas_gra_comun as K  # noqa: E402

BASE = K.REC / "pruebas-gra-telegraf-vector"
IMG = "gl-gra-logs:1"
BANCO_W = str(BASE).replace("/", "\\")
REPS = 2


def correr(h, esc, esp, vol=f"{K.P}-logsbase", timeout=1200):
    r = K.sh(["docker", "run", "--name", f"{K.P}-{h}", "--cpus", "2", "--memory", "1g", "--memory-swap", "1g",
              "-v", f"{vol}:/base:ro", "-v", f"{BANCO_W}:/banco:ro", IMG, "bash", "/banco/medir.sh", h, esc, str(esp)],
             timeout=timeout)
    return (r.stdout + r.stderr).strip()


def res(t):
    m = re.search(r"RESULTADO lineas=(\d+) seg_total=([\d.]+) seg_primera=([\d.]+) cpu_s=([\d.]+) pico_mib=([\d.]+) "
                  r"final_mib=([\d.]+) hilos=(\d+)", t)
    h = re.search(r"HUELLA (\w+) filas=(\d+) suma_col2=(\d+)", t)
    if not m or not h:
        return None
    return dict(lineas=int(m.group(1)), seg=float(m.group(2)), primera=float(m.group(3)), cpu=float(m.group(4)),
                pico=float(m.group(5)), huella=h.group(1), suma=h.group(3))


def main():
    n = int(sys.argv[1])
    out = K.Salida(BASE / f"SALIDA-corrida-{n}.md")
    orden = ["telegraf", "vector"] if n % 2 else ["vector", "telegraf"]
    with K.candado(f"telegraf-vector corrida {n}"):
        out.w(f"# Telegraf (Go) contra Vector (Rust) · corrida {n}\n")
        out.w(f"> Corrida del {datetime.now():%Y-%m-%d %H:%M}, con el candado `_logs/banco.lock` tomado y sin `fab producir` "
              f"vivo. Regenerable con `python investigacion/pruebas_gra_telegraf_vector.py {n}`. Orden: {' → '.join(orden)}.\n")
        out.w("## 0 · Versiones y tamaños\n")
        v = K.sh(["docker", "run", "--name", f"{K.P}-ver", IMG, "sh", "-c",
                  "telegraf --version; vector --version; for f in /usr/local/bin/telegraf /usr/bin/vector; do "
                  "echo \"$f $(stat -c %s $f) bytes = $(awk -v b=$(stat -c %s $f) 'BEGIN{printf \"%.1f\", b/1048576}') MiB\"; done"],
                 timeout=120)
        for l in v.stdout.strip().splitlines():
            out.w(f"- `{l}`")
        for i in ("telegraf:1.40.4", "timberio/vector:0.62.0-debian"):
            s = K.sh(["docker", "image", "inspect", i, "--format", "{{.Size}}"]).stdout.strip()
            out.w(f"- imagen oficial `{i}`: {int(s) / 2**20:.1f} MiB")
        for f in ("telegraf-json.conf", "vector-json.yaml", "telegraf-json-stdout.conf", "vector-json-stdout.yaml",
                  "telegraf-nginx.conf", "vector-nginx.yaml"):
            ls = [l for l in (BASE / f).read_text(encoding="utf-8").splitlines() if l.strip() and not l.strip().startswith("#")]
            out.w(f"- `{f}`: {len(ls)} líneas sin comentarios")
        out.w(f"- contenedores ajenos vivos al empezar: {K.ajenos() or 'ninguno'} · CPU de Windows: {K.cpu_windows()}\n")

        tabla = {}
        for esc, esp, nom in (("json", 900000, "A · logs JSON, 1.000.000 de líneas → 900.000"),
                              ("json-stdout", 900000, "A2 · lo mismo que A, con la salida por stdout (redirigida a un archivo)"),
                              ("nginx", 1000000, "B · log de acceso de nginx, 1.000.000 de líneas")):
            out.w(f"## {nom}\n")
            for rep in range(1, REPS + 1):
                for h in orden:
                    t = correr(h, esc, esp)
                    out.w(f"### {h} · repetición {rep}\n```\n{t}\n```\n")
                    r = res(t)
                    if not r or r["lineas"] != esp:
                        out.w(f"⛔ {h} no llegó a {esp} líneas: se para.")
                        raise SystemExit(1)
                    tabla.setdefault((esc, h), []).append(r)
            hs = {tabla[(esc, h)][0]["huella"] for h in orden}
            out.w(f"**¿Hicieron el mismo trabajo?** huellas: " + ", ".join(f"{h} {tabla[(esc, h)][0]['huella']}" for h in orden)
                  + (" → ✅ iguales\n" if len(hs) == 1 else " → ⛔ DISTINTAS"))
            if len(hs) != 1:
                raise SystemExit(1)

        out.w("## La tabla (mediana de las repeticiones; eventos/s = líneas de entrada / segundos hasta la última línea escrita)\n")
        out.w("| pipeline | herramienta | segundos | eventos/s | CPU s | pico RAM MiB |")
        out.w("|---|---|---:|---:|---:|---:|")
        for esc in ("json", "json-stdout", "nginx"):
            for h in ("telegraf", "vector"):
                rs = tabla[(esc, h)]
                seg = K.mediana([r["seg"] for r in rs])
                cpu = K.mediana([r["cpu"] for r in rs])
                pico = K.mediana([r["pico"] for r in rs])
                out.w(f"| {esc} | {h} | {seg:.2f} | {1000000 / seg:,.0f} | {cpu:.2f} | {pico:.1f} |"
                      .replace(",", "."))
        out.w("")

        out.w("## La trampa: Telegraf con max_undelivered_lines por defecto (1000), pipeline A, 100.000 líneas → 90.000\n")
        for h, esc in (("telegraf", "json-defecto"), ("telegraf", "json"), ("vector", "json")):
            t = correr(h, esc, 90000, vol=f"{K.P}-logshumo", timeout=600)
            out.w(f"### {h} · {esc}\n```\n{t}\n```\n")

        out.w("## Memoria en reposo (pipeline A cargado, sin archivo de entrada, 15 s)\n")
        for h in ("telegraf", "vector"):
            t = correr(h, "json", 0, timeout=300)
            out.w(f"- {h}: `{t.splitlines()[-1] if t else '(nada)'}`")
        out.w(f"\n> Terminó {datetime.now():%Y-%m-%d %H:%M}. CPU de Windows al final: {K.cpu_windows()}.")


if __name__ == "__main__":
    main()
