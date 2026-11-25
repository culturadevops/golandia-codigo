# -*- coding: utf-8 -*-
"""demos.py — lo de «¿Go o Mojo?» (golandia 904) que no lleva cronómetro → SALIDA-demos.md

  1. versión estable y licencia que declara el paquete de pip (y la del repositorio, leída de GitHub);
  2. ¿hay servidor web? el ejemplo de TCP de la propia documentación de std.net, compilado con la estable 1.2.0 y con la
     nightly de hoy;
  3. lightbug_http, el framework web de la comunidad: ¿archivado?;
  4. ldd del binario de mojo build y lo que ocupa el paquete instalado.

Sin candado: no mide tiempos. Usa las imágenes gl-904-* de banco.py (o las construye).

    python recursos/pruebas-go-vs-mojo/demos.py
"""
import json
import subprocess
import urllib.request
from datetime import datetime
from pathlib import Path

AQUI = Path(__file__).resolve().parent
OUT = []
TCP = '''from std.net import SocketAddrV4, TCPListener, TCPStream

def main() raises:
    var listener = TCPListener(bind=.v4(SocketAddrV4(.localhost, 0)))
    var client = TCPStream(connect=listener.local_addr())
    var server, _ = listener.accept()
    client.write_all("ping".as_bytes())
    var buffer = List[Byte](length=4, fill=0)
    var n = server.read(buffer)
    print("leídos", n, "bytes")
'''


def sh(args, **kw):
    return subprocess.run(args, capture_output=True, text=True, encoding="utf-8", errors="replace", **kw)


def w(s=""):
    print(s, flush=True)
    OUT.append(s)


def bloque(cmd, texto, lineas=12):
    w("```")
    w(f"$ {cmd}")
    for l in texto.strip().splitlines()[:lineas]:
        w(l.rstrip())
    w("```\n")


def gh(url):
    with urllib.request.urlopen(urllib.request.Request(url, headers={"User-Agent": "banco"}), timeout=30) as r:
        return r.read().decode()


def main():
    for t in ("build", "final"):
        sh(["docker", "build", "-q", "--target", t, "-f", str(AQUI / "Dockerfile.mojo"), "-t",
            "gl-904-mojo-etapa:1" if t == "build" else "gl-904-mojo:1", str(AQUI)])
    w("# SALIDA-demos — ¿Go o Mojo? (golandia 904), sin cronómetro\n")
    w(f"> {datetime.now():%Y-%m-%d %H:%M}. Regenerable con `python recursos/pruebas-go-vs-mojo/demos.py`.\n")

    w("## 1 · Versión y licencia\n")
    r = sh(["docker", "run", "--rm", "--ulimit", "core=1", "--entrypoint", "sh", "gl-904-mojo-etapa:1", "-c",
            "mojo --version; pip show mojo mojo-compiler | grep -E '^(Name|Version|License)'"])
    bloque("mojo --version; pip show mojo mojo-compiler | grep -E '^(Name|Version|License)'", r.stdout + r.stderr)
    lic = gh("https://raw.githubusercontent.com/modular/modular/main/LICENSE")
    bloque("curl https://raw.githubusercontent.com/modular/modular/main/LICENSE | head -7", "\n".join(lic.splitlines()[:7]))

    w("## 2 · ¿Servidor web? El ejemplo de TCP de la documentación de std.net\n")
    (AQUI / "_tcp").mkdir(exist_ok=True)
    (AQUI / "_tcp" / "tcp.mojo").write_text(TCP, encoding="utf-8")
    vol = f"{AQUI / '_tcp'}:/t"
    r = sh(["docker", "run", "--rm", "--ulimit", "core=1", "-v", vol, "--entrypoint", "mojo", "gl-904-mojo-etapa:1", "run", "/t/tcp.mojo"])
    bloque("mojo run tcp.mojo   # Mojo 1.2.0, la estable", r.stdout + r.stderr, 8)
    r = sh(["docker", "run", "--rm", "--ulimit", "core=1", "-v", vol, "gl-904-mojo-etapa:1", "sh", "-c",
            "pip install -q -U --pre mojo --index-url https://whl.modular.com/nightly/simple/ --extra-index-url https://pypi.org/simple 2>&1 | grep -v -i notice | tail -2; mojo --version; "
            "mojo run /t/tcp.mojo 2>&1 | grep -v '^ ' | head -8"])
    bloque("pip install -U --pre mojo --index-url https://whl.modular.com/nightly/simple/ --extra-index-url https://pypi.org/simple; mojo --version; mojo run tcp.mojo   # nightly",
           r.stdout + r.stderr, 10)

    w("## 3 · lightbug_http, el framework web de la comunidad\n")
    d = json.loads(gh("https://api.github.com/repos/Lightbug-HQ/lightbug_http"))
    rel = json.loads(gh("https://api.github.com/repos/Lightbug-HQ/lightbug_http/releases?per_page=1"))[0]
    bloque("curl https://api.github.com/repos/Lightbug-HQ/lightbug_http",
           f"archived: {d['archived']}\npushed_at: {d['pushed_at']}\núltima versión: {rel['tag_name']} ({rel['published_at'][:10]})")

    w("## 4 · Lo que necesita el binario para correr\n")
    r = sh(["docker", "run", "--rm", "--ulimit", "core=1", "--entrypoint", "sh", "gl-904-mojo-etapa:1", "-c",
            "ldd /computo | sed 's/ (0x.*//'; du -sh /usr/local/lib/python3.13/site-packages/modular"])
    bloque("ldd computo; du -sh .../site-packages/modular", r.stdout + r.stderr, 14)
    (AQUI / "SALIDA-demos.md").write_text("\n".join(OUT) + "\n", encoding="utf-8")
    print("[ok] SALIDA-demos.md")


if __name__ == "__main__":
    main()
