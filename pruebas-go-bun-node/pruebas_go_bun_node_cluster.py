# -*- coding: utf-8 -*-
"""pruebas_go_bun_node_cluster.py — anexo de «¿Go, Bun o Node?»: lo justo con JavaScript.
  · Bun y Node ejecutan tu JavaScript en UN hilo: con 2 CPU usan medio contenedor. Aquí, con WORKERS=2 (node:cluster: dos procesos).
  · El cómputo de UNA petición, sin concurrencia (campo ns del propio servidor): separa «el lenguaje es lento» de «usa un solo núcleo».
Mismo método que pruebas_go_bun_node.py. Guarda `recursos/pruebas-go-bun-node/SALIDA-procesos.md`.
    python investigacion/pruebas_go_bun_node_cluster.py   (después de pruebas_go_bun_node.py: reconstruye sus imágenes)
"""
import json
import statistics
import sys
import urllib.request
from datetime import datetime
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
import pruebas_go_java_revancha as R  # noqa: E402
import pruebas_go_bun_node as G  # noqa: E402


def computo(imagen, env, ruta, n, veces=300):
    t0 = R.arrancar("x", imagen, env, puerto=9300)
    R.esperar_200(9300, t0)
    url = f"http://127.0.0.1:9300/{ruta}?n={n}"
    for _ in range(200):
        urllib.request.urlopen(url).read()
    v = sorted(int(json.loads(urllib.request.urlopen(url).read())["ns"]) / 1e6 for _ in range(veces))
    R.sh(["docker", "rm", "-f", "srv"])
    return statistics.median(v)


def main():
    R.OUT.clear()
    R.sh(["docker", "network", "create", R.RED])
    for k, df in G.DF.items():   # reconstruir: servidor.js ahora sabe usar WORKERS
        f = G.BASE / f"Dockerfile.{k}"
        r = R.sh(["docker", "build", "-q", "-f", str(f), "-t", G.IMG[k], str(G.BASE)])
        assert r.returncode == 0, r.stderr[-1500:]
    R.w("# SALIDA — ¿Go, Bun o Node? Anexo: un hilo contra dos procesos, y el cómputo de una petición\n")
    R.w(f"> Corrida del {datetime.now():%Y-%m-%d %H:%M}. Regenerable con `python investigacion/pruebas_go_bun_node_cluster.py`. "
        "Mismo método que SALIDA.md.\n")
    variantes = [("go", G.IMG["go"], {}), ("bun", G.IMG["bun"], {}), ("bun · 2 procesos", G.IMG["bun"], {"WORKERS": "2"}),
                 ("node", G.IMG["node"], {}), ("node · 2 procesos", G.IMG["node"], {"WORKERS": "2"})]
    R.w("## A · Cómputo de UNA petición (mediana del reloj del propio servidor, sin concurrencia)\n")
    R.w("| variante | /trabajo (20.000 objetos) | /suma (200.000 vueltas) |")
    R.w("|---|---|---|")
    for nombre, img, env in [variantes[0], variantes[1], variantes[3]]:
        R.w(f"| {nombre} | {computo(img, env, 'trabajo', R.N_TRABAJO):.2f} ms | {computo(img, env, 'suma', R.N_SUMA):.3f} ms |")
    R.w()
    R.VARIANTES[:] = variantes
    R.carga("trabajo", R.N_TRABAJO, "B · /trabajo con 2 procesos de JavaScript")
    R.carga("suma", R.N_SUMA, "C · /suma con 2 procesos de JavaScript")
    (G.BASE / "SALIDA-procesos.md").write_text("\n".join(R.OUT) + "\n", encoding="utf-8")
    print("\n[ok] SALIDA-procesos.md escrita")


if __name__ == "__main__":
    main()
