# -*- coding: utf-8 -*-
"""Mide las respuestas del vídeo 703 en golang:1.27.2 y deja resultados.json.

Por muestra (un módulo con go.mod «go 1.27» y el tarea.go de la IA):
  go build .                                   → compila
  go vet .                                     → vet limpio
  go vet -vettool=$(go env GOTOOLDIR)/fix -json .  → avisos de go fix, por modernizer
  go fix -diff .                               → el diff que propone (se guarda)
Un contenedor (--rm, --cpus 2, nombre v703-medir) y la caché de Go en un directorio temporal.
"""
import glob
import json
import os
import shutil
import subprocess
import sys
import tempfile

AQUI = os.path.dirname(os.path.abspath(__file__))
IMAGEN = "golang:1.27.2"

SH = r"""
set -u
FIX=$(go env GOTOOLDIR)/fix
uno() {
  d="$1"; cd "/w/m/$d"
  go build . > build.txt 2>&1; echo $? > build.rc
  go vet . > vet.txt 2>&1; echo $? > vet.rc
  go vet -vettool=$FIX -json . > fix.json 2> fix.err; echo $? > fix.rc
  go fix -diff . > fix.diff 2>&1; echo $? > fixdiff.rc
}
export -f uno; export FIX
ls /w/m | xargs -P 2 -I{} bash -c 'uno {}'
echo LISTO
"""


def main():
    muestras = sorted(glob.glob(os.path.join(AQUI, "respuestas", "*", "*", "*", "s*", "tarea.go")))
    muestras += [os.path.join(AQUI, "control", c, "tarea.go") for c in ("viejo", "moderno")]
    tmp = tempfile.mkdtemp(prefix="v703-medir-")
    os.makedirs(os.path.join(tmp, "m"))
    claves = {}
    for i, p in enumerate(muestras):
        rel = os.path.relpath(os.path.dirname(p), os.path.join(AQUI, "respuestas")).replace("\\", "/")
        if "control" in p.split(os.sep):
            rel = "control-%s/-/control/s1" % os.path.basename(os.path.dirname(p))
        k = "m%04d" % i
        claves[k] = rel
        d = os.path.join(tmp, "m", k)
        os.makedirs(d)
        open(os.path.join(d, "go.mod"), "w", newline="\n").write("module tarea\n\ngo 1.27\n")
        open(os.path.join(d, "tarea.go"), "w", encoding="utf-8", newline="\n").write(open(p, encoding="utf-8").read())
    open(os.path.join(tmp, "medir.sh"), "w", newline="\n").write(SH)
    os.makedirs(os.path.join(tmp, "cache"))
    r = subprocess.run(["docker", "run", "--rm", "--name", "v703-medir", "--cpus", "2", "-v", tmp + ":/w",
                        "-e", "GOCACHE=/w/cache", "-e", "GOFLAGS=-mod=mod", "-e", "GOTOOLCHAIN=local",
                        IMAGEN, "bash", "/w/medir.sh"], capture_output=True, text=True)
    print(r.stdout[-300:], r.stderr[-2000:])
    res = []
    for k, rel in claves.items():
        d = os.path.join(tmp, "m", k)
        leer = lambda n: open(os.path.join(d, n), encoding="utf-8", errors="replace").read()
        partes = rel.split("/")
        fix = {}
        try:
            j = json.loads(leer("fix.json") or "{}")
            for pkg in j.values():
                for analizador, diags in pkg.items():
                    if isinstance(diags, list):
                        fix[analizador] = len(diags)
        except json.JSONDecodeError:
            pass
        codigo = leer("tarea.go")
        res.append({"muestra": rel, "modelo": partes[0], "variante": partes[1], "tarea": partes[2], "seed": partes[3],
                    "vacio": not codigo.strip(),
                    "compila": leer("build.rc").strip() == "0", "build": leer("build.txt")[:1500],
                    "vet_ok": leer("vet.rc").strip() == "0", "vet": leer("vet.txt")[:1500],
                    "fix": fix, "fix_diff": leer("fix.diff")[:4000]})
    json.dump(res, open(os.path.join(AQUI, "resultados.json"), "w", encoding="utf-8"), ensure_ascii=False, indent=1)
    shutil.rmtree(tmp, ignore_errors=True)
    print("muestras:", len(res))


if __name__ == "__main__":
    sys.exit(main())
