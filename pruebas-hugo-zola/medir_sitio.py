# -*- coding: utf-8 -*-
"""Se ejecuta DENTRO del contenedor (2 CPU, 1 GiB). Imprime una línea JSON por medición.

    python3 medir_sitio.py iguales <origen>             build de los dos y comparación de lo generado
    python3 medir_sitio.py build <origen> <reps>        build completo: primera (sin public/) y repetida (con public/)
    python3 medir_sitio.py serve <origen> <ediciones>   servidor de desarrollo: listo, RAM, editar una página y una plantilla
<origen> = carpeta con hugo/ y zola/ (de generar_sitio.py). Se copia a /tmp/w antes de cada cosa: el origen no se toca.
"""
import html
import json
import os
import re
import shutil
import signal
import subprocess
import sys
import time
import urllib.request

W = "/tmp/w"
CMD = {"hugo": ["hugo", "--quiet"], "zola": ["zola", "build"]}


def copiar(origen):
    shutil.rmtree(W, ignore_errors=True)
    shutil.copytree(origen, W)


def build(h, env=None, tope=1500):
    """h = «hugo», «zola» o «hugo:VAR=valor» (la misma herramienta con una variable de entorno)."""
    var = None
    if ":" in h:
        h, var = h.split(":", 1)
    d = f"{W}/{h}"
    cmd = ["/usr/bin/time", "-f", "%e %M %U %S"] + CMD[h]
    entorno = dict(os.environ)
    if var:
        k, v = var.split("=", 1)
        entorno[k] = v
    try:
        r = subprocess.run(cmd, cwd=d, capture_output=True, text=True, timeout=tope, env=entorno)
    except subprocess.TimeoutExpired:
        subprocess.run(["pkill", "-9", "-f", CMD[h][0]])
        return dict(herr=h + (f" [{var}]" if var else ""), seg=None, rss_mib=None, cpu=None, rc="timeout", html=0, err=f"más de {tope} s")
    h = h + (f" [{var}]" if var else "")
    ult = r.stderr.strip().splitlines()[-1].split() if r.stderr.strip() else ["0", "0", "0", "0"]
    try:
        seg, kb, u, s = float(ult[0]), int(ult[1]), float(ult[2]), float(ult[3])
    except (ValueError, IndexError):
        seg = kb = u = s = -1
    n = sum(1 for _, _, fs in os.walk(f"{d}/public") for f in fs if f.endswith(".html")) if os.path.isdir(f"{d}/public") else 0
    return dict(herr=h, seg=seg, rss_mib=round(kb / 1024, 1), cpu=round(u + s, 2), rc=r.returncode, html=n,
                err=r.stderr[-600:] if r.returncode else "")


def archivos(d):
    out = []
    for raiz, _, fs in os.walk(d):
        for f in fs:
            out.append(os.path.relpath(os.path.join(raiz, f), d))
    return sorted(out)


def texto(ruta, clase):
    s = open(ruta, encoding="utf-8").read()
    m = re.search(rf'<div class="{clase}">(.*?)</div></article>', s, re.S)
    t = m.group(1) if m else ""
    t = re.sub(r"<[^>]+>", " ", t)
    return " ".join(html.unescape(t).split())


def iguales(origen):
    copiar(origen)
    res = {h: build(h) for h in ("hugo", "zola")}
    fa, fz = archivos(f"{W}/hugo/public"), archivos(f"{W}/zola/public")
    solo_h, solo_z = sorted(set(fa) - set(fz)), sorted(set(fz) - set(fa))
    paginas = [f for f in fa if f.startswith("posts/nota-")]
    muestra = paginas[:: max(1, len(paginas) // 25)]
    dif = [p for p in muestra if texto(f"{W}/hugo/public/{p}", "contenido") != texto(f"{W}/zola/public/{p}", "contenido")]
    print(json.dumps(dict(res=res, archivos_hugo=len(fa), archivos_zola=len(fz), solo_hugo=solo_h[:10], solo_zola=solo_z[:10],
                          muestra=len(muestra), texto_distinto=dif[:5],
                          ejemplo=texto(f"{W}/hugo/public/{muestra[0]}", "contenido")[:160] if muestra else "")), flush=True)


def esperar(url, contiene=None, tope=600):
    t0 = time.perf_counter()
    while time.perf_counter() - t0 < tope:
        try:
            with urllib.request.urlopen(url, timeout=5) as r:
                cuerpo = r.read().decode("utf-8", "replace")
                if contiene is None or contiene in cuerpo:
                    return time.perf_counter() - t0
        except Exception:  # noqa: BLE001
            pass
        time.sleep(0.005)
    return None


def rss(pid):
    d = {}
    for l in open(f"/proc/{pid}/status"):
        if l.startswith(("VmRSS", "VmHWM")):
            k, v = l.split(":")
            d[k] = round(int(v.split()[0]) / 1024, 1)
    return d


def serve(origen, ediciones):
    for h in ("hugo", "zola"):
        copiar(origen)
        d = f"{W}/{h}"
        if h == "hugo":
            cmd, base = ["hugo", "server", "--bind", "127.0.0.1", "--port", "1313", "--disableLiveReload", "--logLevel", "error"], \
                "http://127.0.0.1:1313"
        else:
            cmd, base = ["zola", "serve", "--interface", "127.0.0.1", "--port", "1111"], "http://127.0.0.1:1111"
        t0 = time.perf_counter()
        p = subprocess.Popen(cmd, cwd=d, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL, start_new_session=True)
        try:
            listo = esperar(f"{base}/posts/nota-000001/", "Nota 000001")
            listo = None if listo is None else round(time.perf_counter() - t0, 3)
            time.sleep(2)
            mem = rss(p.pid)
            print(json.dumps(dict(herr=h, que="listo", seg=listo, **mem)), flush=True)
            paginas = sorted(f for f in os.listdir(f"{d}/content/posts") if f.startswith("nota-"))
            paso = max(1, len(paginas) // (ediciones + 1))
            for k in range(ediciones):
                f = paginas[(k + 1) * paso]
                marca = f"MARCA-{h}-{k}-{time.time_ns()}"
                url = f"{base}/posts/{f[:-3]}/"
                esperar(url)                       # visitada antes de editar, como en el navegador
                time.sleep(1.0)
                t1 = time.perf_counter()
                with open(f"{d}/content/posts/{f}", "a", encoding="utf-8") as fh:
                    fh.write(f"\n\nPárrafo nuevo {marca}.\n")
                seg = esperar(url, marca, tope=300)
                print(json.dumps(dict(herr=h, que="editar-pagina", k=k, seg=None if seg is None else round(seg, 3))), flush=True)
            plantilla = f"{d}/layouts/page.html" if h == "hugo" else f"{d}/templates/page.html"
            for k in range(3):
                f = paginas[-(k + 2)]
                marca = f"PLANTILLA-{h}-{k}-{time.time_ns()}"
                url = f"{base}/posts/{f[:-3]}/"
                esperar(url)
                time.sleep(1.5)
                s = open(plantilla, encoding="utf-8").read()
                # texto visible: html/template de Go quita los comentarios HTML, así que la marca va en un <b>
                s = s.replace("<article>", f'<article><b class="v">{marca}</b>', 1) if '<b class="v">' not in s else \
                    re.sub(r'<b class="v">[^<]*</b>', f'<b class="v">{marca}</b>', s, count=1)
                with open(plantilla, "w", encoding="utf-8") as fh:
                    fh.write(s)
                seg = esperar(url, marca, tope=600)
                print(json.dumps(dict(herr=h, que="editar-plantilla", k=k, seg=None if seg is None else round(seg, 3))), flush=True)
            mem = rss(p.pid)
            print(json.dumps(dict(herr=h, que="memoria-al-final", **mem)), flush=True)
        finally:
            os.killpg(p.pid, signal.SIGTERM)
            try:
                p.wait(timeout=20)
            except subprocess.TimeoutExpired:
                os.killpg(p.pid, signal.SIGKILL)


def main():
    modo, origen = sys.argv[1], sys.argv[2]
    if modo == "iguales":
        iguales(origen)
    elif modo == "build":
        # build <origen> <reps> <herr,herr,...> [repetida] [tope]
        reps, herrs = int(sys.argv[3]), sys.argv[4].split(",")
        repetida = len(sys.argv) > 5 and sys.argv[5] == "repetida"
        tope = int(sys.argv[6]) if len(sys.argv) > 6 else 1500
        for i in range(reps):
            orden = herrs if i % 2 == 0 else herrs[::-1]
            for h in orden:
                copiar(origen)          # sin public/ ni caché: la primera build
                d = build(h, tope=tope)
                d.update(rep=i, que="primera")
                print(json.dumps(d, ensure_ascii=False), flush=True)
                if repetida and d["rc"] == 0:
                    d = build(h, tope=tope)            # la misma, otra vez, con public/ ya escrito
                    d.update(rep=i, que="repetida")
                    print(json.dumps(d, ensure_ascii=False), flush=True)
    elif modo == "serve":
        serve(origen, int(sys.argv[3]))


if __name__ == "__main__":
    main()
