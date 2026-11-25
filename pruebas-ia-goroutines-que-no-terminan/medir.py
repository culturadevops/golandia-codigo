# -*- coding: utf-8 -*-
"""Mide las respuestas del vídeo 705 (y los dos controles) en golang:1.27.2 y deja resultados.json.

Por muestra (módulo con go.mod «go 1.27», el tarea.go de la IA y la prueba fija de su tarea):
  go build .                          → compila (sin las pruebas)
  go vet .                            → vet limpio (sin las pruebas: solo el código de la IA)
  go test -vet=off -count=1 -v .      → termina, correcto y fugas (perfil goroutineleak) → línea RESULTADO
  go test -vet=off -race -count=1 .   → data races (WARNING: DATA RACE)
Controles: control/bien (no debe dar nada) y control/mal (debe cazar cada fallo).
Un contenedor (--rm, --cpus 2, nombre v705-medir); caché de Go en un directorio temporal.
"""
import glob
import json
import os
import re
import shutil
import subprocess
import sys
import tempfile

AQUI = os.path.dirname(os.path.abspath(__file__))
IMAGEN = "golang:1.27.2"
TAREAS = [t["id"] for t in json.load(open(os.path.join(AQUI, "tareas.json"), encoding="utf-8"))["tareas"]]
EN_MAL = ["c02-con-timeout", "c04-el-mas-rapido", "c06-generador", "c09-primer-error", "c10-cache-aciertos", "c11-contar-primos"]

SH = r"""
set -u
uno() {
  d="$1"; cd "/w/m/$d"
  go build . > build.txt 2>&1; echo $? > build.rc
  go vet . > vet.txt 2>&1; echo $? > vet.rc
  cp _pruebas/*_test.go .
  timeout 60 go test -vet=off -count=1 -v . > test.txt 2>&1; echo $? > test.rc
  timeout 180 go test -vet=off -race -count=1 . > race.txt 2>&1; echo $? > race.rc
}
export -f uno
ls /w/m | xargs -P 2 -I{} bash -c 'uno {}'
echo LISTO
"""


def preparar(tmp, k, codigo, tarea):
    d = os.path.join(tmp, "m", k)
    os.makedirs(os.path.join(d, "_pruebas"))
    open(os.path.join(d, "go.mod"), "w", newline="\n").write("module tarea\n\ngo 1.27\n")
    open(os.path.join(d, "tarea.go"), "w", encoding="utf-8", newline="\n").write(open(codigo, encoding="utf-8").read())
    for n in ("ayuda_test.go", tarea + "_test.go"):
        shutil.copy(os.path.join(AQUI, "pruebas", n), os.path.join(d, "_pruebas", n))


def main():
    tmp = tempfile.mkdtemp(prefix="v705-medir-")
    os.makedirs(os.path.join(tmp, "m"))
    os.makedirs(os.path.join(tmp, "cache"))
    claves = {}
    i = 0
    for p in sorted(glob.glob(os.path.join(AQUI, "respuestas", "*", "*", "*", "s*", "tarea.go"))):
        rel = os.path.relpath(os.path.dirname(p), os.path.join(AQUI, "respuestas")).replace("\\", "/")
        claves["m%04d" % i] = rel
        preparar(tmp, "m%04d" % i, p, rel.split("/")[2])
        i += 1
    extra = [("control-bien", "control/bien", TAREAS), ("control-mal", "control/mal", EN_MAL),
             ("caso-ia", "demos/caso/ia", ["c04-el-mas-rapido"]), ("caso-arreglado", "demos/caso/arreglado", ["c04-el-mas-rapido"])]
    for nombre, carpeta, tareas in extra:
        for t in tareas:
            rel = "%s/-/%s/s1" % (nombre, t)
            claves["m%04d" % i] = rel
            preparar(tmp, "m%04d" % i, os.path.join(AQUI, carpeta, "tarea.go"), t)
            i += 1
    open(os.path.join(tmp, "medir.sh"), "w", newline="\n").write(SH)
    r = subprocess.run(["docker", "run", "--rm", "--name", "v705-medir", "--cpus", "2", "-v", tmp + ":/w",
                        "-e", "GOCACHE=/w/cache", "-e", "GOTOOLCHAIN=local", IMAGEN, "bash", "/w/medir.sh"],
                       capture_output=True, text=True)
    print(r.stdout[-300:], r.stderr[-2000:])
    res = []
    for k, rel in claves.items():
        d = os.path.join(tmp, "m", k)

        def leer(n):
            try:
                return open(os.path.join(d, n), encoding="utf-8", errors="replace").read()
            except FileNotFoundError:
                return ""
        partes = rel.split("/")
        test = leer("test.txt")
        m = re.search(r"RESULTADO termina=(\w+) correcto=(\w+) fugas=(\d+)", test)
        pilas = re.findall(r"PILAS\n(.*?)\nFIN-PILAS", test, flags=re.S)
        race = leer("race.txt")
        res.append({"muestra": rel, "modelo": partes[0], "variante": partes[1], "tarea": partes[2], "seed": partes[3],
                    "compila": leer("build.rc").strip() == "0", "build": leer("build.txt")[:1500],
                    "vet_ok": leer("vet.rc").strip() == "0", "vet": leer("vet.txt")[:1500],
                    "prueba_compila": bool(m) or "RESULTADO" in test,
                    "termina": m.group(1) == "true" if m else None,
                    "correcto": m.group(2) == "true" if m else None,
                    "fugas": int(m.group(3)) if m else None,
                    "pilas": pilas[0][:3000] if pilas else "",
                    "races": race.count("WARNING: DATA RACE"),
                    "race_txt": race[:3000] if "DATA RACE" in race else "",
                    "test_txt": test[-1500:]})
    json.dump(res, open(os.path.join(AQUI, sys.argv[1] if len(sys.argv) > 1 else "resultados.json"), "w", encoding="utf-8"), ensure_ascii=False, indent=1)
    shutil.rmtree(tmp, ignore_errors=True)
    print("muestras:", len(res))


if __name__ == "__main__":
    sys.exit(main())
