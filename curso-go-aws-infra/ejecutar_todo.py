# -*- coding: utf-8 -*-
"""Levanta un emulador de AWS limpio (moto) en Docker, corre las 13 lecciones y regenera SALIDA.md.

La regla de la casa: nada de lo que sale en el video se escribe a mano. Sale de aqui.
Hace falta Go 1.24+, Docker, Node + el CLI de CDK (`npm i -g aws-cdk`) y Python 3.10+.

    python ejecutar_todo.py            # y al terminar borra el contenedor
    python ejecutar_todo.py --dejar    # lo deja corriendo, para jugar con la CLI de AWS
"""
import os
import platform
import subprocess
import sys
import tempfile
import time
import urllib.request
from datetime import datetime, timezone
from pathlib import Path

AQUI = Path(__file__).resolve().parent
CONTENEDOR = "curso-aws-moto"
IMAGEN = "motoserver/moto:latest"
ENDPOINT = "http://localhost:5000"
ENV = dict(os.environ, PYTHONIOENCODING="utf-8", AWS_ENDPOINT_URL=ENDPOINT,
           AWS_ACCESS_KEY_ID="test", AWS_SECRET_ACCESS_KEY="test", AWS_REGION="us-east-1",
           AWS_DEFAULT_REGION="us-east-1", CDK_DISABLE_VERSION_CHECK="1")
NPX = "npx.cmd" if os.name == "nt" else "npx"

LECCIONES = [
    ("01", "En que cuenta y region estoy", "01_donde_estoy"),
    ("02", "Inventario: reservas, instancias y paginas", "02_paginacion"),
    ("03", "Filtros y etiquetas", "03_filtros"),
    ("04", "Crear sin huerfanos ni duplicados", "04_crear"),
    ("05", "Grupos de seguridad y errores de la API", "05_grupos_seguridad"),
    ("06", "Una red desde cero, y el orden para borrarla", "06_vpc"),
    ("07", "DryRun: el error que significa adelante", "07_dry_run"),
    ("08", "El filtro vacio que selecciona TODO", "08_filtro_vacio"),
    ("09", "CDK en Go: lo que escribe una linea", "09_cdk_synth"),
    ("10", "cdk diff: cambiar un nombre no es renombrar", "10_cdk_diff"),
    ("11", "cdk destroy no borra todo: RemovalPolicy", "11_cdk_destroy"),
    ("12", "La deriva que cdk diff no ve", "12_deriva"),
]


def _una(cmd, env, timeout):
    r = subprocess.run(cmd, capture_output=True, text=True, encoding="utf-8",
                       errors="replace", cwd=AQUI, env=env, timeout=timeout)
    return r.returncode, (r.stdout or "") + (r.stderr or "")


def correr(cmd, env=ENV, timeout=900):
    """Ejecuta un comando. Si es `go run`/`go build` y Windows lo bloquea, reintenta.

    En Windows con App Control («Una directiva de Control de aplicaciones bloqueó este
    archivo»), un binario sin firmar puede quedar bloqueado por su hash, y cuál se bloquea
    cambia de una compilación a otra. -buildid cambia los bytes del binario, no el programa:
    con otro hash, se reintenta. En Linux y macOS nunca pasa.
    """
    codigo, salida = _una(cmd, env, timeout)
    for intento in range(1, 6):
        if "Control de aplicaciones" not in salida and "Application Control" not in salida:
            break
        if cmd[:2] not in (["go", "run"], ["go", "build"]):
            break
        codigo, salida = _una([*cmd[:2], f"-ldflags=-buildid=curso-{intento}", *cmd[2:]], env, timeout)
    return codigo, salida


def levantar_moto():
    subprocess.run(["docker", "rm", "-f", CONTENEDOR], capture_output=True)
    subprocess.run(["docker", "run", "-d", "--name", CONTENEDOR, "-p", "5000:5000", IMAGEN],
                   check=True, capture_output=True)
    for _ in range(60):
        try:
            urllib.request.urlopen(ENDPOINT + "/moto-api/", timeout=2)
            return
        except Exception:
            time.sleep(1)
    raise SystemExit("moto no arranco")


def bloque(titulo, cmd_txt, codigo, salida):
    return [f"### {titulo}" if titulo else "", "", f"`{cmd_txt}` · codigo de salida `{codigo}`",
            "", "```", salida.strip("\n"), "```", ""]


def proyecto():
    """La lección 13: inventario contra una cuenta con problemas, una región limpia y una cuenta ajena."""
    tmp = Path(tempfile.mkdtemp(prefix="curso-aws-"))
    exe = tmp / ("inventario.exe" if os.name == "nt" else "inventario")
    subprocess.run(["go", "build", "-trimpath", "-o", str(exe), "./13_proyecto"], cwd=AQUI, check=True)
    c, s = correr(["go", "run", "-trimpath", "./13_proyecto/escenario"])
    partes = bloque("El escenario", "go run ./13_proyecto/escenario", c, s)
    fallos = c != 0
    for titulo, args, esperado in [
        ("Con hallazgos", ["-cuenta", "123456789012"], 1),
        ("Lo mismo, en JSON", ["-cuenta", "123456789012", "-json"], 1),
        ("Una region en orden", ["-regiones", "us-west-2"], 0),
        ("La cuenta no es la esperada", ["-cuenta", "987654321098"], 2),
    ]:
        c, s = correr([str(exe), *args])
        fallos += c != esperado
        partes += bloque(titulo, "inventario " + " ".join(args), c, s)
    return partes, fallos


def main():
    sello = datetime.now(timezone.utc).strftime("%Y-%m-%d %H:%M:%SZ")
    print("  levantando moto limpio ...", flush=True)
    levantar_moto()
    _, gov = correr(["go", "version"])
    _, motov = correr(["docker", "exec", CONTENEDOR, "python", "-c", "import moto; print(moto.__version__)"])
    _, cdkv = correr([NPX, "--no-install", "cdk", "--version"])
    _, nodev = correr(["node", "--version"])
    libv = next((l.split()[1] for l in (AQUI / "cdk" / "go.mod").read_text().splitlines()
                 if "awscdk/v2 " in l), "?")
    sdkv = next((l.split()[1] for l in (AQUI / "go.mod").read_text().splitlines()
                 if l.strip().startswith("github.com/aws/aws-sdk-go-v2/service/ec2 ")), "?")
    partes = ["# SALIDA — curso «Infraestructura de AWS desde Go: SDK y CDK»", "",
              "> Generado por `ejecutar_todo.py`. **No se escribe a mano.**", "",
              f"- fecha: `{sello}`", f"- {gov.strip()}",
              f"- emulador: moto `{motov.strip()}` (imagen `{IMAGEN}`), en `{ENDPOINT}`",
              f"- SDK: `aws-sdk-go-v2/service/ec2 {sdkv}` · CDK: CLI `{cdkv.strip()}`, "
              f"`aws-cdk-go/awscdk/v2 {libv}`, node `{nodev.strip()}`",
              f"- sistema: `{platform.system()} {platform.release()}`", ""]
    fallos = 0
    for num, titulo, d in LECCIONES:
        print(f"  [{num}] {titulo} ...", end=" ", flush=True)
        codigo, salida = correr(["go", "run", "-trimpath", f"./{d}"])
        fallos += codigo != 0
        partes += [f"## Leccion {num} — {titulo}", "",
                   f"`go run ./{d}` · codigo de salida `{codigo}`", "", "```", salida.strip("\n"), "```", ""]
        if num == "01":
            sin = dict(ENV, AWS_ENDPOINT_URL="")
            c2, s2 = correr(["go", "run", "-trimpath", f"./{d}"], env=sin)
            fallos += c2 == 0
            partes += bloque("Y sin el emulador", "AWS_ENDPOINT_URL= go run ./01_donde_estoy", c2, s2)
        print("ok" if codigo == 0 else f"FALLO (codigo {codigo})")

    print("  [13] El proyecto: inventario ...", end=" ", flush=True)
    p, f = proyecto()
    fallos += f
    partes += ["## Leccion 13 — El proyecto: inventario", "", *p]
    print("ok" if not f else f"FALLO ({f})")

    (AQUI / "SALIDA.md").write_text("\n".join(partes), encoding="utf-8")
    if "--dejar" not in sys.argv:
        subprocess.run(["docker", "rm", "-f", CONTENEDOR], capture_output=True)
    print(f"\n  -> SALIDA.md regenerado (13 lecciones, {fallos} fallos)")
    return 1 if fallos else 0


if __name__ == "__main__":
    sys.exit(main())
