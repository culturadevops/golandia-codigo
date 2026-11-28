# -*- coding: utf-8 -*-
"""pruebas_gra_comun.py — lo común de los bancos de la serie «Go contra Rust: herramientas populares, lo medí» (golandia
959-961, Lo usan pruebas_gra_restic_rustic.py, pruebas_gra_telegraf_vector.py y
pruebas_gra_proxy.py.

Reglas de operación:
  · TODO lo de Docker con prefijo gl-gra- (contenedores, volúmenes, red, imágenes);
  · cada `docker run` lleva --rm, --ulimit core=1, nombre gl-gra-… y un TIMEOUT duro: si vence se borra el contenedor
    por su nombre y la corrida se para (una corrida se colgó 4 h el 23-nov);
  · se mide SOLO con el candado _logs/banco.lock (en la raíz del repo) (mkdir atómico, con quien.txt dentro) y sin un
    `fab producir` vivo; el candado se suelta al terminar, también si falla.
"""
import os
import subprocess
import sys
import time
from contextlib import contextmanager
from datetime import datetime
from pathlib import Path

CANDADO = Path(__file__).resolve().parent.parent / "_logs" / "banco.lock"
REC = Path(__file__).resolve().parent.parent  # la raíz del repo: cada banco vive en su carpeta
P = "gl-gra"
_n = [0]


class Colgado(Exception):
    pass


def producir_vivo():
    r = subprocess.run(["powershell", "-NoProfile", "-Command",
                        "Get-CimInstance Win32_Process | Where-Object { $_.CommandLine -match 'fab.py\"?\\s+producir|"
                        "remotion.*render' -and $_.CommandLine -notmatch 'Get-CimInstance' } | "
                        "ForEach-Object { $_.ProcessId }"], capture_output=True, text=True, timeout=60)
    return [l for l in r.stdout.split() if l.strip()]


@contextmanager
def candado(que, espera_max=4 * 3600):
    t0 = time.time()
    while True:
        vivos = producir_vivo()
        if not vivos:
            try:
                CANDADO.mkdir()
                break
            except FileExistsError:
                pass
        if time.time() - t0 > espera_max:
            raise SystemExit(f"⛔ no conseguí el candado en {espera_max} s (producir vivo: {vivos})")
        time.sleep(2)
    (CANDADO / "quien.txt").write_text(f"{P} · {que} · pid {os.getpid()} · {datetime.now():%Y-%m-%d %H:%M:%S}\n",
                                       encoding="utf-8")
    print(f"[candado] tomado para {que} tras {time.time() - t0:.0f} s", flush=True)
    try:
        yield
    finally:
        try:
            (CANDADO / "quien.txt").unlink()
        except FileNotFoundError:
            pass
        CANDADO.rmdir()
        print("[candado] suelto", flush=True)


def sh(args, timeout=180, check=False, **kw):
    a = [str(x) for x in args]
    nombre = None
    if a[:2] == ["docker", "run"]:
        extra = []
        if "--rm" not in a:
            extra.append("--rm")
        if "--ulimit" not in a:
            extra += ["--ulimit", "core=1"]
        if "--name" in a:
            nombre = a[a.index("--name") + 1]
            assert nombre.startswith(P + "-"), nombre
        else:
            _n[0] += 1
            nombre = f"{P}-tmp-{os.getpid()}-{_n[0]}"
            extra += ["--name", nombre]
        a = a[:2] + extra + a[2:]
    try:
        r = subprocess.run(a, capture_output=True, text=True, encoding="utf-8", errors="replace", timeout=timeout, **kw)
    except subprocess.TimeoutExpired:
        if nombre:
            subprocess.run(["docker", "rm", "-f", nombre], capture_output=True, timeout=120)
        raise Colgado(f"⛔ timeout de {timeout} s en: {' '.join(a)[:300]}")
    if check and r.returncode != 0:
        raise SystemExit(f"⛔ rc={r.returncode} en {' '.join(a)[:300]}\n{r.stdout[-2000:]}\n{r.stderr[-2000:]}")
    return r


def rm(*nombres):
    for n in nombres:
        assert n.startswith(P + "-"), n
        subprocess.run(["docker", "rm", "-f", n], capture_output=True, timeout=120)


def ajenos():
    r = sh(["docker", "ps", "--format", "{{.Names}}"])
    return [n for n in r.stdout.split() if not n.startswith(P + "-")]


def cpu_windows():
    try:
        r = subprocess.run(["powershell", "-NoProfile", "-Command", "(Get-CimInstance Win32_Processor | Measure-Object "
                            "-Property LoadPercentage -Average).Average"], capture_output=True, text=True, timeout=60)
        return r.stdout.strip() + " %"
    except Exception as e:  # noqa: BLE001
        return f"(no se pudo leer: {e})"


class Salida:
    def __init__(self, ruta):
        self.ruta = Path(ruta)
        self.l = []

    def w(self, s=""):
        print(s, flush=True)
        self.l.append(s)
        self.ruta.write_text("\n".join(self.l) + "\n", encoding="utf-8")


def mediana(xs):
    xs = sorted(xs)
    n = len(xs)
    return xs[n // 2] if n % 2 else (xs[n // 2 - 1] + xs[n // 2]) / 2


if __name__ == "__main__":
    # python pruebas_gra_comun.py -- <comando...>   → corre el comando con el candado tomado (pruebas de humo)
    if "--" in sys.argv:
        cmdl = sys.argv[sys.argv.index("--") + 1:]
        with candado("humo: " + " ".join(cmdl)[:80]):
            rc = subprocess.run(cmdl, timeout=3600).returncode
        sys.exit(rc)
    print("producir vivo:", producir_vivo())
