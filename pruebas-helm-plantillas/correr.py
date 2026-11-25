# -*- coding: utf-8 -*-
"""Demo del vídeo «Las plantillas de Helm son de Go» (golandia 676).

    python correr.py            → todo: Go, Helm, las trampas, helm lint y kubeconform

Imágenes fijas: golang:1.27.2 · alpine/helm:4.3.1 · ghcr.io/yannh/kubeconform:v0.7.0.
Cada caso se renderiza en un chart temporal (tempfile) con su plantilla y sus values; nada se borra fuera de él.
Las salidas quedan en salidas/*.txt sin editar; SALIDA.md se escribe a partir de ellas.
"""
import os
import shutil
import subprocess
import tempfile
from pathlib import Path

AQUI = Path(__file__).resolve().parent
SAL = AQUI / "salidas"
SAL.mkdir(exist_ok=True)
ENV = {**os.environ, "MSYS_NO_PATHCONV": "1"}
GO = "golang:1.27.2"
HELM = "alpine/helm:4.3.1"
KC = "ghcr.io/yannh/kubeconform:v0.7.0"
CHART_YAML = "apiVersion: v2\nname: caso\nversion: 0.1.0\n"

# (salida, carpeta del caso, plantilla, values)
CASOS = [
    ("01-sin-guion", "01-guiones", "sin-guion.yaml", "values.yaml"),
    ("01-guiones-mal", "01-guiones", "mal.yaml", "values.yaml"),
    ("01-guiones-bien", "01-guiones", "bien.yaml", "values.yaml"),
    ("02-indent-mal", "02-nindent", "mal.yaml", "values.yaml"),
    ("02-nindent-bien", "02-nindent", "bien.yaml", "values.yaml"),
    ("03-toyaml-mal", "03-toyaml", "mal.yaml", "values.yaml"),
    ("03-toyaml-bien", "03-toyaml", "bien.yaml", "values.yaml"),
    ("04-default-mal", "04-default", "mal.yaml", "values.yaml"),
    ("04-default-bien", "04-default", "bien.yaml", "values.yaml"),
    ("05-quote-mal", "05-quote", "mal.yaml", "values.yaml"),
    ("05-quote-bien", "05-quote", "bien.yaml", "values.yaml"),
    ("05-quote-texto", "05-quote", "bien.yaml", "values-texto.yaml"),
    ("06-required-mal", "06-required", "mal.yaml", "values.yaml"),
    ("06-required-sin-imagen", "06-required", "mal.yaml", "values-sin-imagen.yaml"),
    ("06-required-bien", "06-required", "bien.yaml", "values.yaml"),
    ("06-required-con-repo", "06-required", "bien.yaml", "values-con-repo.yaml"),
    ("07-range-mal", "07-scope", "mal.yaml", "values.yaml"),
    ("07-range-bien", "07-scope", "bien.yaml", "values.yaml"),
    ("07-with-mal", "07-scope", "with-mal.yaml", "values.yaml"),
    ("07-with-bien", "07-scope", "with-bien.yaml", "values.yaml"),
    ("08-tpl-mal", "08-tpl", "mal.yaml", "values.yaml"),
    ("08-tpl-bien", "08-tpl", "bien.yaml", "values.yaml"),
    ("09-template-tubo", "09-include", "template-tubo.yaml", "values.yaml"),
    ("09-template-solo", "09-include", "template-solo.yaml", "values.yaml"),
    ("09-include-bien", "09-include", "bien.yaml", "values.yaml"),
    ("10-env", "10-env", "mal.yaml", "values.yaml"),
]
# Los que además pasan por kubeconform (helm template no valida tipos)
KUBECONFORM = {"03-toyaml-mal", "03-toyaml-bien", "05-quote-mal", "05-quote-bien", "04-default-mal",
               "09-template-solo", "06-required-mal"}
# Los que además pasan por helm lint (para ver qué atrapa y qué no)
LINT = {"03-toyaml-mal", "05-quote-mal", "04-default-mal", "09-template-solo", "01-guiones-mal", "06-required-mal"}


def docker(*args, entrada=None):
    r = subprocess.run(["docker", "run", "--rm", "-i", *args], input=entrada, capture_output=True, text=True,
                       encoding="utf-8", errors="replace", env=ENV)
    return r.returncode, r.stdout + r.stderr


def guardar(nombre, cabecera, codigo, texto):
    (SAL / f"{nombre}.txt").write_text(f"$ {cabecera}\n{texto.rstrip()}\n[código de salida: {codigo}]\n", encoding="utf-8")
    print(f"== {nombre} (salida {codigo})")


def go(*args):
    return docker("-v", "gomodcache-g676:/go/pkg/mod", "-v", f"{AQUI}:/w", "-w", "/w/go", GO, "go", "run", ".", *args)


def helm_template(chart_dir):
    return docker("-v", f"{chart_dir}:/chart", HELM, "template", "demo", "/chart")


def kubeconform(yaml_txt):
    return docker("-v", "kubeconform-cache-g676:/cache", KC, "-strict", "-summary", "-cache", "/cache", "-", entrada=yaml_txt)


def parte_go():
    for modo in ("cuenta", "guiones", "vacios"):
        guardar(f"go-{modo}", f"go run . {modo}", *go(modo))
    for capa in ("solo-go", "sprig", "helm"):
        guardar(f"go-{capa}", f"go run . {capa} ../mini", *go(capa, "../mini"))


def comparar():
    c, h = docker("-v", f"{AQUI / 'mini'}:/chart", HELM, "template", "demo", "/chart")
    guardar("helm-mini", "helm template demo mini", c, h)
    _, g = go("helm", "../mini")
    cuerpo_helm = h.split("\n", 2)[2]  # quita «---» y «# Source: …»
    igual = cuerpo_helm.rstrip("\n") == g.rstrip("\n")
    texto = (f"helm template (sin las 2 líneas de cabecera): {len(cuerpo_helm.encode())} bytes\n"
             f"go run . helm: {len(g.encode())} bytes\n¿idénticos? {'sí' if igual else 'NO'}")
    guardar("comparar", "comparar helm template con go run . helm", 0 if igual else 1, texto)


def un_caso(nombre, carpeta, plantilla, values):
    origen = AQUI / "casos" / carpeta
    with tempfile.TemporaryDirectory(prefix="g676-") as tmp:
        t = Path(tmp)
        (t / "templates").mkdir()
        (t / "Chart.yaml").write_text(CHART_YAML, encoding="utf-8")
        shutil.copy(origen / values, t / "values.yaml")
        shutil.copy(origen / plantilla, t / "templates" / "caso.yaml")
        for tpl in origen.glob("_*.tpl"):
            shutil.copy(tpl, t / "templates" / tpl.name)
        c, out = helm_template(t)
        if nombre in LINT:
            guardar(f"{nombre}-lint", f"helm lint {carpeta}  # {plantilla}", *docker("-v", f"{t}:/chart", HELM, "lint", "/chart"))
    guardar(nombre, f"helm template demo {carpeta}  # {plantilla} + {values}", c, out)
    if nombre in KUBECONFORM and c == 0:
        guardar(f"{nombre}-kubeconform", "helm template … | kubeconform -strict -summary -", *kubeconform(out))


def parte_chart():
    montaje = ["-v", f"{AQUI / 'chart'}:/chart"]
    guardar("chart-lint", "helm lint chart", *docker(*montaje, HELM, "lint", "/chart"))
    c, out = docker(*montaje, HELM, "template", "demo", "/chart")
    guardar("chart-template", "helm template demo chart", c, out)
    guardar("chart-kubeconform", "helm template demo chart | kubeconform -strict -summary -", *kubeconform(out))


def versiones():
    _, g = docker(GO, "go", "version")
    _, h = docker(HELM, "version")
    _, k = docker(KC, "-v")
    d = subprocess.run(["docker", "version", "--format", "{{.Server.Version}}"], capture_output=True, text=True).stdout
    guardar("versiones", "versiones", 0, f"docker {d.strip()}\n{g.strip()}\nhelm {h.strip()}\nkubeconform {k.strip()}")


if __name__ == "__main__":
    versiones()
    parte_go()
    comparar()
    for caso in CASOS:
        un_caso(*caso)
    parte_chart()
