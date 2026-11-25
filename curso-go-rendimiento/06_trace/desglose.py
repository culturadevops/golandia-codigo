"""Lección 6: el desglose por goroutine de `go tool trace`, en texto.

La interfaz de `go tool trace` es una web. Este script la levanta, lee la página de las
goroutines del trabajo (las que arranca wg.Go) y la imprime como tabla: cuánto tiempo pasó cada
una ejecutando, bloqueada en un candado y esperando a que el planificador le diera una CPU.

    python3 06_trace/desglose.py traza.out
"""
import html
import re
import subprocess
import sys
import time
import urllib.parse
import urllib.request

PUERTO = 6061
BASE = f"http://127.0.0.1:{PUERTO}"


def leer(ruta):
    with urllib.request.urlopen(BASE + ruta, timeout=30) as r:
        return r.read().decode("utf-8", "replace")


def a_ms(texto):
    t = texto.strip().replace("µ", "u").replace("�", "u")
    m = re.fullmatch(r"([\d.]+)(ns|us|ms|s)", t)
    if not m:
        return 0.0
    return float(m.group(1)) * {"ns": 1e-6, "us": 1e-3, "ms": 1.0, "s": 1000.0}[m.group(2)]


def main():
    traza = sys.argv[1]
    srv = subprocess.Popen(["go", "tool", "trace", f"-http=127.0.0.1:{PUERTO}", traza],
                           stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
    try:
        for _ in range(300):
            try:
                portada = leer("/goroutines")
                break
            except OSError:
                time.sleep(0.1)
        else:
            raise SystemExit("go tool trace no arrancó")
        enlaces = re.findall(r'href="(/goroutine\?name=[^"]+)"', portada)
        grupo = next(e for e in enlaces if "WaitGroup" in urllib.parse.unquote(e))
        pagina = leer(html.unescape(grupo))
        tabla = pagina[pagina.index('class="details"'):]
        tabla = tabla[:tabla.index("</table>")]
        cols = [html.unescape(re.sub(r"<[^>]+>", "", c)).strip()
                for c in re.findall(r"<th[^>]*>(.*?)</th>", tabla, re.S)]
        filas = []
        for tr in re.findall(r"<tr>(.*?)</tr>", tabla, re.S)[1:]:
            celdas = [html.unescape(re.sub(r"<[^>]+>", "", c)).strip()
                      for c in re.findall(r"<td[^>]*>(.*?)</td>", tr, re.S)]
            filas.append(dict(zip(cols, celdas)))
    finally:
        srv.terminate()
        srv.wait()

    quiero = [("Execution time", "ejecutando"), ("Block time (sync)", "en un candado"),
              ("Sched wait time", "esperando CPU")]
    print(f"go tool trace · {len(filas)} goroutines de trabajo (wg.Go), en ms")
    print(f"  {'goroutine':>9}  {'total':>7}" + "".join(f"  {n:>13}" for _, n in quiero))
    sumas = [0.0] * len(quiero)
    for f in filas:
        vals = [a_ms(f.get(k, "0s")) for k, _ in quiero]
        sumas = [a + b for a, b in zip(sumas, vals)]
        print(f"  {f['Goroutine']:>9}  {a_ms(f['Total']):7.1f}" + "".join(f"  {v:13.1f}" for v in vals))
    print(f"  {'suma':>9}  {'':>7}" + "".join(f"  {v:13.1f}" for v in sumas))


if __name__ == "__main__":
    main()
