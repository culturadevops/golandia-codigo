# -*- coding: utf-8 -*-
"""grb_comun.py — lo común de los bancos de la serie «Go contra Rust: herramientas populares, lo medí» (golandia,

Reglas de operación:
  · TODO lo de Docker con prefijo gl-grb- (imágenes, contenedores, volúmenes);
  · cada `docker run` lleva --rm, --ulimit core=1, nombre gl-grb-… y un TIMEOUT duro: si vence, se borra el contenedor
    por su nombre y la medición se para (una corrida se colgó 4 h el 23-nov);
  · el candado _logs/banco.lock (mkdir atómico) se toma para CADA medición y se suelta entre mediciones, para que otro
    agente pueda medir; y no se mide si hay un `fab.py producir` vivo.
"""
import os
import statistics
import subprocess
import sys
import time
from contextlib import contextmanager
from pathlib import Path

CANDADO = Path(__file__).resolve().parent.parent / "_logs" / "banco.lock"
REC = Path(__file__).resolve().parent.parent  # la raíz del repo: cada banco vive en su carpeta
P = "gl-grb"
_n = [0]


class Colgado(Exception):
    pass


def producir_vivo():
    r = subprocess.run(["powershell", "-NoProfile", "-Command",
                        "Get-CimInstance Win32_Process | Where-Object { $_.Name -match 'python' -and $_.CommandLine -match 'fab.py' -and "
                        "$_.CommandLine -match 'producir' } | Measure-Object | Select-Object -ExpandProperty Count"],
                       capture_output=True, text=True, timeout=60)
    try:
        return int(r.stdout.strip() or "0") > 0
    except ValueError:
        return False


@contextmanager
def candado(que, espera_max=3 * 3600):
    """Toma el candado (mkdir atómico) para UNA medición; lo suelta al salir, pase lo que pase."""
    t0 = time.time()
    avisado = False
    while True:
        if not producir_vivo():
            try:
                os.mkdir(CANDADO)
                (CANDADO / "quien.txt").write_text(f"gl-grb · {que} · pid {os.getpid()} · {time.strftime('%Y-%m-%d %H:%M:%S')}\n",
                                                   encoding="utf-8")
                break
            except FileExistsError:
                pass
        if not avisado:
            print(f"   … esperando candado / producir para «{que}»", flush=True)
            avisado = True
        if time.time() - t0 > espera_max:
            raise SystemExit(f"⛔ no conseguí el candado en {espera_max} s para «{que}»")
        time.sleep(1)
    try:
        yield
    finally:
        try:
            (CANDADO / "quien.txt").unlink(missing_ok=True)
            os.rmdir(CANDADO)
        except OSError as e:
            print(f"⚠️ no pude soltar el candado: {e}", flush=True)


def docker_run(args, timeout, nombre=None, check=True):
    """docker run --rm --ulimit core=1 --name gl-grb-… con timeout duro. args = lo que va después de «docker run»."""
    if nombre is None:
        _n[0] += 1
        nombre = f"{P}-tmp-{os.getpid()}-{_n[0]}"
    assert nombre.startswith(P + "-"), nombre
    a = ["docker", "run", "--rm", "--ulimit", "core=1", "--name", nombre] + [str(x) for x in args]
    try:
        r = subprocess.run(a, capture_output=True, text=True, encoding="utf-8", errors="replace", timeout=timeout)
    except subprocess.TimeoutExpired:
        subprocess.run(["docker", "rm", "-f", nombre], capture_output=True, timeout=120)
        raise Colgado(f"⛔ timeout de {timeout} s en: {' '.join(a)[:300]}")
    if check and r.returncode != 0:
        raise RuntimeError(f"⛔ código {r.returncode} en: {' '.join(a)[:300]}\n--- stdout\n{r.stdout[-3000:]}\n--- stderr\n{r.stderr[-3000:]}")
    return r


def mediana(xs):
    return statistics.median(xs)


class Salida:
    """Escribe la SALIDA.md a medida que avanza (si algo se cuelga, lo medido queda)."""

    def __init__(self, ruta):
        self.ruta = Path(ruta)
        self.ruta.parent.mkdir(parents=True, exist_ok=True)
        self.ruta.write_text("", encoding="utf-8")

    def w(self, s=""):
        print(s, flush=True)
        with self.ruta.open("a", encoding="utf-8") as f:
            f.write(s + "\n")


def host():
    try:
        r = subprocess.run(["docker", "info", "--format", "{{.NCPU}} CPU · {{.MemTotal}} B · {{.KernelVersion}} · {{.ServerVersion}}"],
                           capture_output=True, text=True, timeout=60)
        return r.stdout.strip()
    except Exception as e:  # noqa: BLE001
        return f"(sin docker info: {e})"


if __name__ == "__main__":
    print("producir vivo:", producir_vivo())
    sys.exit(0)
