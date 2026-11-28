# -*- coding: utf-8 -*-
"""Genera un repositorio git sintético y determinista con `git fast-import`: N commits sobre ARCH archivos repartidos en
carpetas; cada commit cambia un archivo. Asunto de cada commit: «commit 000123: …», así se reconoce en pantalla.
Al final deja el árbol de trabajo «sucio» como el de un día normal: 20 archivos modificados y 10 sin seguimiento.

    python3 generar_repo.py <N> <ARCH> <destino>
"""
import os
import subprocess
import sys


def main():
    n, arch, destino = int(sys.argv[1]), int(sys.argv[2]), sys.argv[3]
    os.makedirs(destino, exist_ok=True)
    subprocess.run(["git", "init", "-q", destino], check=True)
    p = subprocess.Popen(["git", "fast-import", "--quiet"], cwd=destino, stdin=subprocess.PIPE)
    w = p.stdin.write
    t0 = 1600000000
    for i in range(1, n + 1):
        msg = f"commit {i:06d}: cambia el archivo {i % arch:05d}\n".encode()
        w(b"commit refs/heads/main\n")
        w(f"committer banco <banco@example.com> {t0 + i * 60} +0000\n".encode())
        w(b"data %d\n" % len(msg) + msg)
        if i == 1:
            for k in range(arch):
                cont = f"archivo {k}\nversión inicial\n".encode()
                w(f"M 644 inline src/carpeta-{k % 50:02d}/archivo-{k:05d}.txt\n".encode())
                w(b"data %d\n" % len(cont) + cont)
        else:
            k = i % arch
            cont = f"archivo {k}\nversión del commit {i}\n".encode()
            w(f"M 644 inline src/carpeta-{k % 50:02d}/archivo-{k:05d}.txt\n".encode())
            w(b"data %d\n" % len(cont) + cont)
        w(b"\n")
    p.stdin.close()
    if p.wait() != 0:
        raise SystemExit("fast-import falló")
    subprocess.run(["git", "checkout", "-q", "main"], cwd=destino, check=True)
    subprocess.run(["git", "reset", "-q", "--hard"], cwd=destino, check=True)
    for k in range(20):
        ruta = os.path.join(destino, f"src/carpeta-{(k * 7) % 50:02d}/archivo-{(k * 7):05d}.txt")
        with open(ruta, "a", encoding="utf-8") as f:
            f.write("cambio sin commit\n")
    for k in range(10):
        with open(os.path.join(destino, f"nuevo-{k:02d}.txt"), "w", encoding="utf-8") as f:
            f.write("sin seguimiento\n")
    subprocess.run(["git", "gc", "-q", "--aggressive"] if os.environ.get("GC") else ["git", "repack", "-adq"], cwd=destino, check=True)


if __name__ == "__main__":
    main()
