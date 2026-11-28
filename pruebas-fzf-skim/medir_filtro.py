# -*- coding: utf-8 -*-
"""Se ejecuta DENTRO del contenedor (2 CPU, 1 GiB). Imprime una línea JSON por medición.

    python3 medir_filtro.py tiempos <lista> <reps> <consulta>...      tiempo, RAM máx. y CPU de cada filtro, alternando el orden
    python3 medir_filtro.py iguales <lista> <consulta>...             ¿las dos devuelven el MISMO conjunto de líneas?
    python3 medir_filtro.py calidad <lista> <n>                       ¿en qué puesto sale la ruta que buscabas?
"""
import hashlib
import json
import random
import subprocess
import sys

HERR = {
    "fzf": lambda q: ["fzf", f"--filter={q}"],
    "sk": lambda q: ["sk", f"--filter={q}"],
    "sk-skim_v2": lambda q: ["sk", "--algo", "skim_v2", f"--filter={q}"],
}


def correr(nombre, q, lista, salida="/tmp/salida"):
    cmd = ["/usr/bin/time", "-f", "%e %M %U %S"] + HERR[nombre](q)
    with open(lista, "rb") as fin, open(salida, "wb") as fout:
        r = subprocess.run(cmd, stdin=fin, stdout=fout, stderr=subprocess.PIPE, timeout=600)
    ult = r.stderr.decode().strip().splitlines()[-1].split()
    seg, kb, u, s = float(ult[0]), int(ult[1]), float(ult[2]), float(ult[3])
    return dict(herr=nombre, q=q, seg=seg, rss_mib=round(kb / 1024, 1), cpu=round(u + s, 2), rc=r.returncode)


def lineas(ruta):
    with open(ruta, "rb") as f:
        return f.read().splitlines()


def main():
    modo, lista = sys.argv[1], sys.argv[2]
    subprocess.run(["sh", "-c", f"cat {lista} > /dev/null"])   # la lista en la caché de páginas, igual para todos
    if modo == "tiempos":
        reps, consultas = int(sys.argv[3]), sys.argv[4:]
        herrs = ["fzf", "sk", "sk-skim_v2"]
        for q in consultas:
            for i in range(reps):
                orden = herrs if i % 2 == 0 else herrs[::-1]
                for h in orden:
                    d = correr(h, q, lista)
                    d["rep"] = i
                    print(json.dumps(d), flush=True)
    elif modo == "iguales":
        for q in sys.argv[3:]:
            res = {}
            for h in HERR:
                correr(h, q, lista)
                ls = lineas("/tmp/salida")
                res[h] = dict(n=len(ls), conjunto=hashlib.sha256(b"\n".join(sorted(ls))).hexdigest()[:16],
                              top10=[x.decode() for x in ls[:10]])
            print(json.dumps(dict(q=q, res=res)), flush=True)
    elif modo == "calidad":
        n = int(sys.argv[3])
        todas = lineas(lista)
        r = random.Random(7)
        for k in range(n):
            obj = r.choice(todas).decode()
            partes = obj.split("/")
            arch = partes[-1]
            # cómo lo escribe una persona: el proyecto, el principio de la última carpeta y el nombre del archivo
            q1 = f"{partes[0]} {partes[-2][:4]} {arch}"
            # y abreviado, sin espacios: tres letras del proyecto, tres de la carpeta, el nombre sin extensión
            q2 = partes[0][:3] + partes[-2][:3] + arch.split(".")[0]
            for estilo, q in (("fragmentos", q1), ("abreviado", q2)):
                fila = dict(k=k, objetivo=obj, estilo=estilo, q=q)
                for h in HERR:
                    correr(h, q, lista)
                    ls = lineas("/tmp/salida")
                    pos = next((i + 1 for i, x in enumerate(ls) if x.decode() == obj), None)
                    fila[h] = dict(puesto=pos, n=len(ls))
                print(json.dumps(fila), flush=True)


if __name__ == "__main__":
    main()
