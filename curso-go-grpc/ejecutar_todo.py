# -*- coding: utf-8 -*-
"""ejecutar_todo.py — corre las lecciones del curso de gRPC y regenera SALIDA.md.

Todo corre en Docker (imagen curso-grpc:1, ver Dockerfile.curso): Go, protoc y grpcurl iguales en
cualquier máquina. La lección 7 levanta réplicas en una red de Docker y la orquesta este script.

    python ejecutar_todo.py            todas
    python ejecutar_todo.py 03 07      solo esas
"""
import datetime
import os
import subprocess
import sys
import time
from pathlib import Path

AQUI = Path(__file__).resolve().parent
IMAGEN = "curso-grpc:1"
RED = "curso-grpc-red"
VOL = ["-v", f"{AQUI.as_posix()}:/curso", "-v", "curso-grpc-gomod:/go/pkg/mod",
       "-v", "curso-grpc-gocache:/root/.cache/go-build"]
ENV = dict(os.environ, MSYS_NO_PATHCONV="1")

LECCIONES = [
    ("01", "01_protobuf", "protobuf: el número es el contrato"),
    ("02", "02_cliente", "El cliente: NewClient no conecta"),
    ("03", "03_plazos", "Plazos: sin deadline se espera para siempre"),
    ("04", "04_errores", "Errores: lo que cruza el cable, y el pánico"),
    ("05", "05_streaming", "Streaming: control de flujo y cancelar"),
    ("06", "06_tamano", "El límite de 4 MB"),
    ("07", "07_balanceo", "El balanceo: la conexión que se pega"),
    ("08", "08_apagar", "Apagar sin cortar: GracefulStop, health e interceptores"),
    ("09", "09_rest_o_grpc", "¿gRPC o REST?"),
]


def docker(*args, check=True, capturar=True):
    r = subprocess.run(["docker", *args], capture_output=capturar, text=True, encoding="utf-8",
                       errors="replace", env=ENV)
    if check and r.returncode != 0:
        raise SystemExit(f"[X] docker {' '.join(args[:3])}: {r.stderr.strip()}")
    return r


def go(*cmd, red=None):
    extra = ["--network", red] if red else []
    return docker("run", "--rm", *extra, *VOL, IMAGEN, *cmd, check=False)


def quitar(*nombres):
    for n in nombres:
        docker("rm", "-f", n, check=False)


def replica(nombre, edad=None):
    env = ["-e", f"MAX_CONNECTION_AGE={edad}"] if edad else []
    docker("run", "-d", "--name", nombre, "--hostname", nombre, "--network", RED,
           "--network-alias", "inventario", *env, *VOL, IMAGEN, "/curso/bin/l07", "servidor")


def cliente_en_vivo(fase, al_ver=None):
    """Corre el cliente leyendo su salida línea a línea; `al_ver` se llama con cada una."""
    p = subprocess.Popen(["docker", "run", "--rm", "--network", RED, *VOL, IMAGEN, "/curso/bin/l07",
                          "cliente", fase], stdout=subprocess.PIPE, stderr=subprocess.STDOUT,
                         text=True, encoding="utf-8", env=ENV)
    out = []
    for linea in p.stdout:
        linea = linea.rstrip("\n")
        if al_ver and al_ver(linea):
            continue
        out.append(linea)
    return p.wait(), "\n".join(out)


def leccion_07():
    """Las réplicas y el cliente, cada uno en su contenedor."""
    r = go("sh", "-c", "CGO_ENABLED=0 go build -o bin/l07 ./07_balanceo")
    if r.returncode:
        return r.returncode, r.stdout + r.stderr
    nombres = ["replica-a", "replica-b", "replica-c"]
    quitar(*nombres)
    docker("network", "rm", RED, check=False)
    docker("network", "create", RED)
    trozos, codigo = [], 0
    try:
        replica("replica-a")
        replica("replica-b")
        time.sleep(1)
        ip = docker("run", "--rm", "--network", RED, IMAGEN, "getent", "ahosts", "inventario").stdout
        ips = sorted({l.split()[0] for l in ip.splitlines() if l.strip()})
        trozos.append("== 7.1 · Dos réplicas con el mismo nombre: el DNS de Docker da las dos ==")
        trozos.append(f"  inventario -> {', '.join(ips)}")
        for titulo, fase in (("7.2 · La trampa: el cliente por defecto (pick_first)", "pick_first"),
                             ("7.3 · Lo correcto: round_robin en la configuración del servicio", "round_robin")):
            c, out = cliente_en_vivo(fase)
            codigo |= c
            trozos += ["", f"== {titulo} ==", out]

        for titulo, edad in (("7.4 · La segunda trampa: escalas a 3 réplicas y la nueva no recibe nada", None),
                             ("7.5 · Lo correcto: MaxConnectionAge en el servidor (aquí, 3 s)", "3s")):
            quitar(*nombres)
            replica("replica-a", edad)
            replica("replica-b", edad)
            time.sleep(1)

            def tercera(linea, edad=edad):
                if linea == "ESPERANDO_TERCERA":
                    replica("replica-c", edad)
                    return True
                return False

            c, out = cliente_en_vivo("escalar", tercera)
            codigo |= c
            trozos += ["", f"== {titulo} ==", out]
    finally:
        quitar(*nombres)
        docker("network", "rm", RED, check=False)
    return codigo, "\n".join(trozos)


def correr(directorio):
    if directorio == "07_balanceo":
        return leccion_07()
    r = go("go", "run", f"./{directorio}")
    return r.returncode, (r.stdout + (r.stderr if r.returncode else "")).strip("\n")


def main():
    pedidas = sys.argv[1:]
    elegidas = [l for l in LECCIONES if not pedidas or l[0] in pedidas]
    docker("image", "inspect", IMAGEN, check=False).returncode == 0 or docker(
        "build", "-t", IMAGEN, "-f", str(AQUI / "Dockerfile.curso"), str(AQUI))
    version = go("go", "version").stdout.strip()
    mods = go("go", "list", "-m", "google.golang.org/grpc", "google.golang.org/protobuf").stdout.split("\n")
    protoc = go("protoc", "--version").stdout.strip()
    partes = ["# SALIDA — curso «gRPC en Go: protobuf, streaming y las trampas que no dan error»", "",
              "> Generado por `ejecutar_todo.py`. **No se escribe a mano.**", "",
              f"- fecha: `{datetime.datetime.now(datetime.timezone.utc):%Y-%m-%d %H:%M:%SZ}`",
              f"- {version}", *[f"- `{m}`" for m in mods if m], f"- `{protoc}`",
              f"- Docker `{docker('version', '--format', '{{.Server.Version}}').stdout.strip()}`"]
    fallos = 0
    for num, d, titulo in elegidas:
        if not (AQUI / d).exists():
            continue
        t0 = time.time()
        codigo, out = correr(d)
        fallos += codigo != 0
        print(f"[{'ok' if codigo == 0 else 'X '}] {num} {d} ({time.time() - t0:.0f} s)")
        partes += ["", f"## Leccion {num} — {titulo}", "",
                   f"`go run ./{d}` · codigo de salida `{codigo}`", "", "```", out, "```"]
    if not pedidas:
        (AQUI / "SALIDA.md").write_bytes(("\n".join(partes) + "\n").encode("utf-8"))
        print("[ok] SALIDA.md")
    else:
        print("\n".join(partes))
    sys.exit(1 if fallos else 0)


if __name__ == "__main__":
    main()
