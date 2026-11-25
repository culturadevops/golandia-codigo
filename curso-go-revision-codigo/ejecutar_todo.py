# -*- coding: utf-8 -*-
"""Corre los 20 fallos, con el código malo y con el arreglo, pasa los linters y regenera SALIDA.md.

La regla de la casa: nada de lo que sale en el video se escribe a mano. Sale de aquí.
Solo hace falta Docker y Python 3.10+. Go, staticcheck y golangci-lint van en contenedores
con la versión fija (abajo).

    python ejecutar_todo.py               # todo, y al terminar borra contenedor y caché
    python ejecutar_todo.py --cache       # deja el volumen de caché para la próxima vez
    python ejecutar_todo.py --sin-linters # solo las pruebas

Cada fallo vive en su carpeta, con dos archivos y una prueba:
    malo.go       el código del PR, tal como llegó   (//go:build malo)
    bien.go       el arreglo                          (//go:build !malo)
    fallo_test.go la prueba que lo demuestra: falla con -tags malo y pasa sin él
"""
import re
import subprocess
import sys
from datetime import datetime, timezone
from pathlib import Path

AQUI = Path(__file__).resolve().parent
IMAGEN_GO = "golang:1.27.2"
IMAGEN_LINT = "golangci/golangci-lint:v2.15.0"
STATICCHECK = "honnef.co/go/tools/cmd/staticcheck@2026.2.2"
CONTENEDOR = "g687-revision-go"
VOLUMEN = "g687-revision-cache"
ENTORNO = ["-e", "GOCACHE=/cache/build", "-e", "GOMODCACHE=/cache/mod", "-e", "GOBIN=/cache/bin",
           "-e", "GOLANGCI_LINT_CACHE=/cache/lint", "-e", "PATH=/cache/bin:/usr/local/go/bin:/go/bin:"
           "/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"]

FALLOS = [
    ("01", "Error ignorado", "01_error_ignorado"),
    ("02", "El nil que no es nil", "02_nil_no_es_nil"),
    ("03", "Errores comparados con ==", "03_errores_con_igual"),
    ("04", "Shadowing con :=", "04_shadowing"),
    ("05", "recover que se traga el pánico", "05_recover_traga"),
    ("06", "La goroutine que se fuga", "06_goroutine_fuga"),
    ("07", "El context que se ignora", "07_context_ignorado"),
    ("08", "defer dentro de un bucle", "08_defer_en_bucle"),
    ("09", "El slice que comparte memoria", "09_slice_compartido"),
    ("10", "range modifica una copia", "10_range_copia"),
    ("11", "Mapa sin candado", "11_mapa_sin_candado"),
    ("12", "sync.Mutex copiado", "12_mutex_copiado"),
    ("13", "WaitGroup.Add dentro de la goroutine", "13_waitgroup_add"),
    ("14", "La variable de bucle (Go 1.21 contra 1.22)", "14_variable_de_bucle"),
    ("15", "time.After dentro del bucle", "15_time_after_bucle"),
    ("16", "Body sin cerrar", "16_body_sin_cerrar"),
    ("17", "Cliente HTTP sin timeout", "17_cliente_sin_timeout"),
    ("18", "JSON con campos sin exportar", "18_json_sin_exportar"),
    ("19", "Conversión que desborda", "19_overflow_conversion"),
    ("20", "Concatenar strings en un bucle", "20_concatenar_en_bucle"),
]


def sh(cmd, timeout=900):
    r = subprocess.run(cmd, capture_output=True, text=True, encoding="utf-8", errors="replace",
                       timeout=timeout)
    return r.returncode, (r.stdout or "") + (r.stderr or "")


def go(orden, timeout=900):
    """Corre una orden de shell dentro del contenedor de Go."""
    return sh(["docker", "exec", CONTENEDOR, "sh", "-c", orden], timeout)


def lint(args):
    return sh(["docker", "run", "--rm", "--name", "g687-revision-lint", "-v", f"{AQUI}:/src",
               "-v", f"{VOLUMEN}:/cache", *ENTORNO, "-w", "/src", IMAGEN_LINT, "golangci-lint", "run",
               "--show-stats=false", *args, "./..."])


def levantar():
    sh(["docker", "rm", "-f", CONTENEDOR])
    c, s = sh(["docker", "run", "-d", "--name", CONTENEDOR, "-v", f"{AQUI}:/src", "-v",
               f"{VOLUMEN}:/cache", *ENTORNO, "-w", "/src", IMAGEN_GO, "sleep", "infinity"])
    if c:
        raise SystemExit("no pude arrancar el contenedor de Go:\n" + s)


RACE_CAB = ("Write at", "Read at", "Previous write at", "Previous read at")


def resumir(salida, tope=24):
    """Lo que importa de la salida de go test: los mensajes de la prueba, no la pila entera."""
    lineas, out, i, carrera = salida.splitlines(), [], 0, 0
    while i < len(lineas):
        l, s = lineas[i], lineas[i].strip()
        if s.startswith("WARNING: DATA RACE"):
            carrera += 1
            if carrera == 1:
                out.append(l)
        elif carrera == 1 and s.startswith(RACE_CAB):
            # de la pila, solo el primer marco que es código del curso (función y archivo)
            j = i + 1
            while j < len(lineas) and lineas[j].strip() and not re.search(r"/src/\d\d_", lineas[j]):
                j += 1
            out.append(l)
            if j < len(lineas) and re.search(r"/src/\d\d_", lineas[j]):
                out.extend([lineas[j - 1], lineas[j]])
        elif (re.match(r"^\s+fallo_test\.go:\d+: ", l) or s.startswith(("--- FAIL", "--- PASS", "fatal error", "panic:",
                                                    "FAIL\t", "ok ", "testing.go:", "Benchmark"))
              or re.match(r"^\d{4}/\d\d/\d\d ", s) or s.startswith(("goos", "cpu:"))):
            out.append(l)
        i += 1
    if carrera > 1:
        out.append(f"(... y {carrera - 1} carreras más del mismo tipo)")
    if len(out) > tope:
        out = out[:tope] + [f"(... {len(out) - tope} líneas más)"]
    return "\n".join(out)


HALLAZGO = re.compile(r"^(?:\./)?(\d\d)_[a-z0-9_]+/(malo|bien)\.go:\d+:\d+: (.*)$")


def hallazgos(salida):
    """{(fallo, archivo): [mensajes]} de la salida de vet, staticcheck o golangci-lint."""
    res = {}
    for l in salida.splitlines():
        m = HALLAZGO.match(l.strip())
        if m:
            res.setdefault((m.group(1), m.group(2)), []).append(m.group(3))
    return res


def detecta(h_malo, h_bien, num):
    """Detecta el fallo si avisa en malo.go con algo que ya no dice en bien.go."""
    norm = lambda m: re.sub(r"\d+", "#", m)
    bien = {norm(m) for m in h_bien.get((num, "bien"), [])}
    return [m for m in h_malo.get((num, "malo"), []) if norm(m) not in bien]


def pruebas(partes):
    fallos, carreras = 0, set()
    for num, titulo, d in FALLOS:
        print(f"  [{num}] {titulo} ...", end=" ", flush=True)
        cm, sm = go(f"go test -count=1 -race -v -tags malo ./{d}/")
        cb, sb = go(f"go test -count=1 -race -v ./{d}/")
        if "WARNING: DATA RACE" in sm:
            carreras.add(num)
        bien = cm != 0 and cb == 0
        fallos += not bien
        partes += [f"## Fallo {num} — {titulo}", "",
                   f"`go test -race -tags malo ./{d}/` · código de salida `{cm}` (el PR)", "",
                   "```", resumir(sm), "```", ""]
        if num in carreras:  # y sin el detector de carreras, que es como corre en producción
            cs, ss = go(f"go test -count=1 -v -tags malo ./{d}/")
            partes += [f"`go test -tags malo ./{d}/` (sin -race) · código de salida `{cs}`", "",
                       "```", resumir(ss, tope=8), "```", ""]
        partes += [f"`go test -race ./{d}/` · código de salida `{cb}` (con el arreglo)", "",
                   "```", resumir(sb), "```", ""]
        print("ok: falla con el malo, pasa con el arreglo" if bien else "NO REPRODUCE")
    return fallos, carreras


def benchmark(partes):
    d = "20_concatenar_en_bucle"
    partes += ["## Medida — el reporte CSV de 5000 pedidos, con benchmark", ""]
    for etiqueta, tags in (("el PR (+=)", "-tags malo"), ("el arreglo (strings.Builder)", "")):
        _, s = go(f"go test -run '^$' -bench CSV -benchmem -count 1 {tags} ./{d}/")
        partes += [f"{etiqueta}: `go test -bench CSV -benchmem {tags} ./{d}/`", "",
                   "```", resumir(s), "```", ""]


def linters(partes, carreras):
    herramientas = [
        ("go vet", lambda t: go(f"go vet {t} ./...")),
        ("staticcheck", lambda t: go(f"staticcheck {t} ./...")),
        ("golangci-lint (de serie)", lambda t: lint(["--no-config", "--max-issues-per-linter=0",
                                                     "--max-same-issues=0", *t.split()])),
        ("golangci-lint (.golangci.yml del curso)", lambda t: lint(t.split())),
    ]
    tabla = {}
    for nombre, correr in herramientas:
        print(f"  {nombre} ...", flush=True)
        _, sm = correr("-tags malo" if "golangci" not in nombre else "--build-tags malo")
        _, sb = correr("")
        hm, hb = hallazgos(sm), hallazgos(sb)
        tabla[nombre] = {n: detecta(hm, hb, n) for n, _, _ in FALLOS}
        en_bien = sum(len(v) for (n, f), v in hb.items() if f == "bien")
        partes += [f"### {nombre}", "", "Con el PR (`malo.go`), lo que avisa y desaparece con el arreglo:",
                   "", "```"]
        partes += [f"{n}  {m}" for n, _, _ in FALLOS for m in tabla[nombre][n]] or ["(nada)"]
        partes += ["```", "", f"Avisos que siguen en el arreglo (`bien.go`): {en_bien}", ""]
        if en_bien:
            partes += ["```"] + [f"{n}  {m}" for (n, f), v in sorted(hb.items()) if f == "bien"
                                 for m in v] + ["```", ""]
    cols = list(tabla)
    partes += ["### La tabla: quién ve cada fallo", "",
               "| # | fallo | " + " | ".join(cols) + " | go test -race |",
               "|---|---|" + "---|" * (len(cols) + 1)]
    escapan = []
    for num, titulo, _ in FALLOS:
        marcas = ["sí" if tabla[c][num] else "·" for c in cols] + ["sí" if num in carreras else "·"]
        if "sí" not in marcas:
            escapan.append(f"{num} {titulo}")
        partes.append(f"| {num} | {titulo} | " + " | ".join(marcas) + " |")
    totales = [sum(1 for n, _, _ in FALLOS if tabla[c][n]) for c in cols] + [len(carreras)]
    partes += ["| | **detectados, de 20** | " + " | ".join(f"**{t}**" for t in totales) + " |", "",
               f"Se escapan a todas las herramientas: **{len(escapan)} de 20**", ""]
    partes += [f"- {e}" for e in escapan] + [""]
    partes += ["«go test -race» solo cuenta si hay una prueba que ejecute ese código a la vez.", ""]


def main():
    sello = datetime.now(timezone.utc).strftime("%Y-%m-%d %H:%M:%SZ")
    print("  arrancando el contenedor de Go ...", flush=True)
    levantar()
    try:
        _, gov = go("go version")
        _, fmt_ = go("gofmt -l .")
        partes = ["# SALIDA — curso «Revisión de código en Go: 20 fallos reales que compilan»", "",
                  "> Generado por `ejecutar_todo.py`. **No se escribe a mano.**", "",
                  f"- fecha: `{sello}`", f"- {gov.strip()} (imagen `{IMAGEN_GO}`)",
                  f"- gofmt -l: `{fmt_.strip() or 'limpio'}`", ""]
        print("  compilando el PR ...", flush=True)
        cc, sc = go("go build -tags malo ./... && go build ./... && echo compila")
        partes += [f"El PR compila: `go build -tags malo ./...` · código `{cc}` · {sc.strip()}", ""]
        fallos, carreras = pruebas(partes)
        benchmark(partes)
        if "--sin-linters" not in sys.argv:
            _, scv = go(f"go install {STATICCHECK} >/dev/null 2>&1; staticcheck -version")
            _, glv = sh(["docker", "run", "--rm", IMAGEN_LINT, "golangci-lint", "version"])
            partes += ["## Los linters, contra el PR y contra el arreglo", "",
                       f"- {scv.strip()}", f"- {glv.strip()}", ""]
            linters(partes, carreras)
        (AQUI / "SALIDA.md").write_text("\n".join(partes) + "\n", encoding="utf-8", newline="\n")
    finally:
        sh(["docker", "rm", "-f", CONTENEDOR])
        if "--cache" not in sys.argv:
            sh(["docker", "volume", "rm", VOLUMEN])
    print(f"\n  -> SALIDA.md regenerado (20 fallos, {fallos} sin reproducir)")
    return 1 if fallos else 0


if __name__ == "__main__":
    sys.exit(main())
