# -*- coding: utf-8 -*-
"""Corre las 10 lecciones del curso de Bubble Tea y regenera SALIDA.md.

La regla de la casa: nada de lo que sale en el vídeo se escribe a mano. Sale de aquí.
Hace falta Go 1.26.8+ (con GOTOOLCHAIN=auto, Go lo baja solo), Python 3.10+, el código de nubecita (para la lección 10) y, para la
prueba de carreras de la lección 4, Docker (en Windows `-race` necesita cgo; en el contenedor
de Go viene de serie).

    python ejecutar_todo.py
    python ejecutar_todo.py --nubecita <nubecita>
"""
import os
import platform
import shutil
import socket
import subprocess
import sys
import tempfile
import time
from datetime import datetime, timezone
from pathlib import Path

AQUI = Path(__file__).resolve().parent
ENV = dict(os.environ, PYTHONIOENCODING="utf-8")
IMAGEN_GO = "golang:1.26"
EXE = ".exe" if os.name == "nt" else ""

LECCIONES = [
    ("01", "El bucle: mensaje, Update, View", ["go", "run", "./01_el_bucle"], 0),
    ("02", "Las teclas", ["go", "run", "./02_teclas"], 0),
    ("03", "Los comandos: el trabajo lento va fuera de Update", ["go", "run", "./03_comandos"], 0),
    ("04", "No se toca el estado desde otra gorrutina", ["go", "run", "./04_gorrutinas"], 0),
    ("05", "tea.Batch no tiene orden", ["go", "run", "./05_batch_sequence"], 0),
    ("06", "Componer modelos", ["go", "run", "./06_componer"], 0),
    ("07", "El ancho de lo que se ve no es len()", ["go", "run", "./07_ancho"], 0),
    ("08", "El modelo es un valor, sus slices no", ["go", "run", "./08_slices"], 0),
    # 09 FALLA a propósito: una de sus pruebas caza el fallo que la otra deja pasar.
    ("09", "Probar el modelo sin terminal", ["go", "test", "-v", "-count=1", "./09_probar"], 1),
]


def correr(cmd, cwd=AQUI, timeout=900, env=ENV):
    r = subprocess.run(cmd, capture_output=True, text=True, encoding="utf-8",
                       errors="replace", cwd=cwd, env=env, timeout=timeout)
    return r.returncode, (r.stdout or "") + (r.stderr or "")


def bloque(titulo, cmd_txt, codigo, salida):
    return [f"### {titulo}", "", f"`{cmd_txt}` · codigo de salida `{codigo}`", "",
            "```", salida.strip("\n"), "```", ""]


def carrera():
    """go run -race de la versión mala de la lección 4, dentro del contenedor oficial de Go."""
    if not shutil.which("docker"):
        return None
    c, _ = correr(["docker", "info"], timeout=60)
    if c != 0:
        return None
    return correr(["docker", "run", "--rm", "-v", f"{AQUI}:/src", "-w", "/src",
                   "-v", "curso-tui-gomod:/go/pkg/mod", IMAGEN_GO, "sh", "-c",
                   "go mod download && go run -race ./04_gorrutinas -carrera"], timeout=900)


def puerto_libre():
    s = socket.socket()
    s.bind(("127.0.0.1", 0))
    p = s.getsockname()[1]
    s.close()
    return p


def proyecto(nubecita_src):
    tmp = Path(tempfile.mkdtemp(prefix="curso-tui-"))
    nube, panel = tmp / f"nubecita{EXE}", tmp / f"nubepanel{EXE}"
    subprocess.run(["go", "build", "-o", str(nube), "."], cwd=nubecita_src, check=True)
    subprocess.run(["go", "build", "-o", str(panel), "./10_proyecto"], cwd=AQUI, check=True)
    partes, fallos = [], 0

    def paso(titulo, cmd, esperado, cmd_txt):
        nonlocal fallos
        c, s = correr(cmd)
        fallos += c != esperado
        partes.extend(bloque(titulo, cmd_txt, c, s))

    paso("El panel contra nubecita, conducido sin teclado",
         ["go", "run", "./10_proyecto/demo", "-nubecita", str(nube)], 0,
         "go run ./10_proyecto/demo -nubecita nubecita")

    # Por ssh: una nubecita sembrada y el panel servido con wish.
    pn = puerto_libre()
    endpoint = f"http://127.0.0.1:{pn}"
    pn_proc = subprocess.Popen([str(nube), "-puerto", str(pn), "-callado"],
                               stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
    time.sleep(1)
    correr(["go", "run", "./10_proyecto/demo", "-sembrar", endpoint])
    procs = []

    def servir(direccion, *extra):
        p = subprocess.Popen([str(panel), "-endpoint", endpoint, "-ssh", direccion,
                              "-clave", str(tmp / "host_ed25519"), *extra], cwd=tmp,
                             stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
        procs.append(p)
        return p

    try:
        a = f"127.0.0.1:{puerto_libre()}"
        servir(a)
        paso("Servido por ssh, sin -autorizadas", ["go", "run", "./10_proyecto/cliente_ssh", "-a", a], 0,
             "go run ./10_proyecto/cliente_ssh   # sin ninguna credencial")

        for nombre in ("buena", "otra"):
            subprocess.run(["ssh-keygen", "-q", "-t", "ed25519", "-N", "", "-f", str(tmp / nombre)], check=True)
        shutil.copy(tmp / "buena.pub", tmp / "authorized_keys")
        b = f"127.0.0.1:{puerto_libre()}"
        servir(b, "-autorizadas", str(tmp / "authorized_keys"))
        paso("Con -autorizadas: sin credenciales", ["go", "run", "./10_proyecto/cliente_ssh", "-a", b], 1,
             "go run ./10_proyecto/cliente_ssh")
        paso("Con -autorizadas: otra clave", ["go", "run", "./10_proyecto/cliente_ssh", "-a", b,
                                              "-clave", str(tmp / "otra")], 1,
             "go run ./10_proyecto/cliente_ssh -clave otra")
        paso("Con -autorizadas: la clave autorizada", ["go", "run", "./10_proyecto/cliente_ssh", "-a", b,
                                                       "-clave", str(tmp / "buena")], 0,
             "go run ./10_proyecto/cliente_ssh -clave buena")
        paso("Escuchar en 0.0.0.0 sin -autorizadas", [str(panel), "-ssh", "0.0.0.0:23299"], 2,
             "nubepanel -ssh 0.0.0.0:23299")
    finally:
        for p in procs:
            p.kill()
        pn_proc.kill()

    # El binario: tamaño normal, sin símbolos, y para Linux desde aquí.
    tam = []
    for etiqueta, extra, goos in [("normal", [], None), ("-ldflags=\"-s -w\"", ["-ldflags=-s -w"], None),
                                  ("linux/amd64 -s -w", ["-ldflags=-s -w"], "linux")]:
        out = tmp / f"np-{len(tam)}"
        env = dict(ENV, CGO_ENABLED="0")
        if goos:
            env.update(GOOS=goos, GOARCH="amd64")
        subprocess.run(["go", "build", *extra, "-o", str(out), "./10_proyecto"], cwd=AQUI, env=env, check=True)
        tam.append(f"  {etiqueta:22s} {out.stat().st_size / 1024 / 1024:5.1f} MiB")
    partes.extend(bloque("El binario", "go build ./10_proyecto  (CGO_ENABLED=0)", 0, "\n".join(tam)))
    return partes, fallos


def main():
    nubecita_src = Path(sys.argv[sys.argv.index("--nubecita") + 1]) if "--nubecita" in sys.argv \
        else Path(r"<nubecita>")
    sello = datetime.now(timezone.utc).strftime("%Y-%m-%d %H:%M:%SZ")
    _, gov = correr(["go", "version"])
    _, mods = correr(["go", "list", "-m", "charm.land/bubbletea/v2", "charm.land/bubbles/v2",
                      "charm.land/lipgloss/v2", "charm.land/wish/v2"])
    partes = ["# SALIDA — curso «La otra pantalla: interfaces de terminal en Go con Bubble Tea»", "",
              "> Generado por `ejecutar_todo.py`. **No se escribe a mano.**", "",
              f"- fecha: `{sello}`", f"- {gov.strip()}",
              *[f"- `{m}`" for m in mods.strip().splitlines()],
              f"- sistema: `{platform.system()} {platform.release()}`", ""]
    fallos = 0
    for num, titulo, cmd, esperado in LECCIONES:
        print(f"  [{num}] {titulo} ...", end=" ", flush=True)
        codigo, salida = correr(cmd)
        fallos += codigo != esperado
        partes += [f"## Leccion {num} — {titulo}", "",
                   f"`{' '.join(cmd)}` · codigo de salida `{codigo}`", "", "```", salida.strip("\n"), "```", ""]
        if num == "04":
            r = carrera()
            if r is None:
                partes += ["### La carrera, con -race", "", "_(sin Docker: no se pudo correr)_", ""]
            else:
                c, s = r
                partes += bloque("La carrera, con -race (contenedor golang:1.26)",
                                 "go run -race ./04_gorrutinas -carrera", c, s)
        print("ok" if codigo == esperado else f"FALLO (codigo {codigo}, esperaba {esperado})")

    print("  [10] El proyecto: nubepanel ...", end=" ", flush=True)
    p, f = proyecto(nubecita_src)
    fallos += f
    partes += ["## Leccion 10 — El proyecto: nubepanel", "", *p]
    print("ok" if not f else f"FALLO ({f})")

    (AQUI / "SALIDA.md").write_text("\n".join(partes), encoding="utf-8")
    print(f"\n  -> SALIDA.md regenerado (10 lecciones, {fallos} fallos)")
    return 1 if fallos else 0


if __name__ == "__main__":
    sys.exit(main())
