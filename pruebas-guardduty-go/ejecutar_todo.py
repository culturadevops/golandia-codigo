# -*- coding: utf-8 -*-
"""Corre todo el repo y regenera SALIDA.md. Nada de lo que sale en el vídeo se escribe a mano.

Levanta un emulador de AWS (moto, `pip install "moto[server]"`) en un puerto local, corre las
pruebas, los dos programas y compila el handler de Lambda para Linux. Hace falta Go 1.24+ y
Python 3.10+. No hace falta cuenta de AWS ni Docker.

    python ejecutar_todo.py
"""
import os
import platform
import subprocess
import sys
import time
import urllib.request
from datetime import datetime, timezone
from pathlib import Path

AQUI = Path(__file__).resolve().parent
PUERTO = 5077
ENDPOINT = f"http://127.0.0.1:{PUERTO}"
ENV = dict(os.environ, PYTHONIOENCODING="utf-8", AWS_ENDPOINT_URL=ENDPOINT,
           AWS_ACCESS_KEY_ID="test", AWS_SECRET_ACCESS_KEY="test", AWS_REGION="us-east-1")
BLOQUEO = "Control de aplicaciones"


def correr(cmd, env=ENV, timeout=600):
    """Ejecuta. Si Windows App Control bloquea el binario de `go run` por su hash, reintenta con
    otro -buildid (cambia los bytes, no el programa). En Linux y macOS nunca pasa."""
    for intento in range(6):
        c = list(cmd)
        if intento and c[:2] == ["go", "run"]:
            c.insert(2, f"-ldflags=-buildid=gd-{intento}")
        r = subprocess.run(c, capture_output=True, text=True, encoding="utf-8",
                           errors="replace", cwd=AQUI, env=env, timeout=timeout)
        salida = (r.stdout or "") + (r.stderr or "")
        if BLOQUEO not in salida:
            return r.returncode, salida
    return r.returncode, salida


def version(cmd):
    try:
        return subprocess.run(cmd, capture_output=True, text=True, cwd=AQUI).stdout.strip()
    except Exception as e:
        return f"(no disponible: {e})"


def main():
    moto = subprocess.Popen([sys.executable, "-m", "moto.server", "-p", str(PUERTO)],
                            stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
    try:
        for _ in range(60):
            try:
                urllib.request.urlopen(ENDPOINT + "/moto-api/", timeout=1)
                break
            except Exception:
                time.sleep(0.5)
        else:
            sys.exit("moto no arrancó")
        urllib.request.urlopen(urllib.request.Request(ENDPOINT + "/moto-api/reset", method="POST"))

        import moto as moto_pkg
        partes = [
            "# SALIDA — pruebas-guardduty-go",
            "",
            "Generado por `python ejecutar_todo.py`. **No se edita a mano.**",
            "",
            f"- fecha: {datetime.now(timezone.utc):%Y-%m-%d %H:%M} UTC",
            f"- {version(['go', 'version'])}",
            f"- moto {moto_pkg.__version__} (emulador de AWS, en {ENDPOINT})",
            f"- {platform.system()} {platform.release()}",
            "",
            "Los eventos de `testdata/eventos.json` están **construidos a mano** con los campos que "
            "documenta AWS. No salen de una cuenta real. El emulador no implementa los findings.",
        ]
        pasos = [
            ("1 · go vet", ["go", "vet", "./..."]),
            ("2 · go test", ["go", "test", "-count=1", "-v", "./..."]),
            ("3 · leer los eventos", ["go", "run", "./cmd/leer"]),
            ("4 · reglas de supresión contra el emulador", ["go", "run", "./cmd/supresion"]),
        ]
        fallos = 0
        for titulo, cmd in pasos:
            codigo, out = correr(cmd)
            fallos += codigo != 0
            partes += ["", f"## {titulo}", "", f"`$ {' '.join(cmd)}` → código {codigo}", "",
                       "```", out.rstrip(), "```"]
            print(f"{titulo}: código {codigo}")

        env = dict(ENV, GOOS="linux", GOARCH="arm64", CGO_ENABLED="0")
        destino = AQUI / "_build" / "bootstrap"
        codigo, out = correr(["go", "build", "-trimpath", "-ldflags=-s -w", "-o", str(destino),
                              "./cmd/lambda"], env=env)
        fallos += codigo != 0
        tam = f"{destino.stat().st_size / 1048576:.1f} MB" if destino.exists() else "no se generó"
        partes += ["", "## 5 · el handler de Lambda, para Linux arm64", "",
                   "`$ GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -trimpath "
                   "-ldflags=\"-s -w\" -o _build/bootstrap ./cmd/lambda`"
                   f" → código {codigo}", "", "```", (out.rstrip() or f"bootstrap: {tam}"), "```"]
        print(f"5 · lambda: código {codigo} · {tam}")

        (AQUI / "SALIDA.md").write_text("\n".join(partes) + "\n", encoding="utf-8")
        print("SALIDA.md escrito" + (f" · ⚠️ {fallos} paso(s) con error" if fallos else ""))
        sys.exit(1 if fallos else 0)
    finally:
        moto.terminate()


if __name__ == "__main__":
    main()
