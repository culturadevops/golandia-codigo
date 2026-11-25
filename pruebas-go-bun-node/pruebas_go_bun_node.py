# -*- coding: utf-8 -*-
"""pruebas_go_bun_node.py — ¿Go, Bun o Node? El MISMO banco que la revancha de Java (pruebas_go_java_revancha.py): mismo servidor,
mismas cargas (/trabajo con objetos y /suma de CPU), hey -c 16 desde otro contenedor, servidores con 2 CPU y 1 GiB, y la comprobación
de que los tres devuelven exactamente lo mismo antes de medir. Pedido en un comentario de «Go vs Node vs Python 2026».
  · Go 1.26 (la imagen de la revancha)   · Bun 1.4.3 (oven/bun:1.4.3-slim, Bun.serve)   · Node 24 LTS (node:24-slim, node:http)
Guarda `recursos/pruebas-go-bun-node/SALIDA.md`.
    python investigacion/pruebas_go_bun_node.py   (después de pruebas_go_java_revancha.py: reutiliza la imagen de Go)
"""
import sys
from datetime import datetime
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
import pruebas_go_java_revancha as R  # noqa: E402

BASE = Path(__file__).resolve().parent.parent / "recursos" / "pruebas-go-bun-node"
IMG = {"go": R.IMG["go"], "bun": "golandia-gobunnode-bun:1", "node": "golandia-gobunnode-node:1"}
DF = {
    "bun": "FROM oven/bun:1.4.3-slim\nWORKDIR /app\nCOPY servidor.js /app/servidor.js\nEXPOSE 8080\nENTRYPOINT [\"bun\", \"servidor.js\"]\n",
    "node": "FROM node:24-slim\nWORKDIR /app\nCOPY servidor.js /app/servidor.js\nEXPOSE 8080\nENTRYPOINT [\"node\", \"servidor.js\"]\n",
}


def main():
    R.OUT.clear()
    R.sh(["docker", "network", "create", R.RED])
    for k, df in DF.items():
        f = BASE / f"Dockerfile.{k}"
        f.write_text(df, encoding="utf-8")
        r = R.sh(["docker", "build", "-q", "-f", str(f), "-t", IMG[k], str(BASE)])
        assert r.returncode == 0, r.stderr[-1500:]
    ver = {
        "go": R.sh(["docker", "run", "--rm", "golang:1.26", "go", "version"]).stdout.strip(),
        "bun": R.sh(["docker", "run", "--rm", "--entrypoint", "bun", IMG["bun"], "--version"]).stdout.strip(),
        "node": R.sh(["docker", "run", "--rm", "--entrypoint", "node", IMG["node"], "--version"]).stdout.strip(),
    }
    R.w("# SALIDA — ¿Go, Bun o Node? (golandia), con el banco de la revancha\n")
    R.w(f"> Corrida del {datetime.now():%Y-%m-%d %H:%M}. Regenerable con `python investigacion/pruebas_go_bun_node.py`. "
        "⛔ Ninguna cifra de blog.\n")
    for k, v in ver.items():
        R.w(f"- {k}: `{v}`")
    R.w(f"- Servidores con {R.CPUS} CPU y {R.MEM}; `hey -c {R.C}` con 4 CPU desde otro contenedor. Cada runtime con su API nativa.\n")
    R.w("## 1 · Lo que despliegas\n")
    R.w("| imagen | base | descarga (comprimida) |")
    R.w("|---|---|---|")
    for k, base in [("go", "scratch"), ("bun", "oven/bun:1.4.3-slim"), ("node", "node:24-slim")]:
        b = int(R.sh(["docker", "image", "inspect", "-f", "{{.Size}}", IMG[k]]).stdout.strip())
        R.w(f"| {k} | {base} | {b / 1048576:.1f} MiB |")
    R.w()
    R.VARIANTES[:] = [("go", IMG["go"], {}), ("bun", IMG["bun"], {}), ("node", IMG["node"], {})]
    R.mismo_resultado()
    R.arranque_y_memoria()
    R.carga("trabajo", R.N_TRABAJO, "4 · Carga con memoria: /trabajo (objetos, mapa, orden, JSON)")
    R.carga("suma", R.N_SUMA, "5 · Carga de CPU pura: /suma")
    (BASE / "SALIDA.md").write_text("\n".join(R.OUT) + "\n", encoding="utf-8")
    print("\n[ok] SALIDA.md escrita")


if __name__ == "__main__":
    main()
