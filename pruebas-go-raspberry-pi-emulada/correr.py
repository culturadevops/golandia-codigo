# -*- coding: utf-8 -*-
"""correr.py — «Go en una Raspberry Pi, EMULADA» (golandia). ⛔ No hay placa: todo bajo QEMU (el binfmt de Docker Desktop).

    python correr.py             → SALIDA.md (lo funcional: compilar, ejecutar, systemd, GPIO simulado)
    python correr.py --tiempos   → SALIDA-tiempos.md (el mismo cálculo nativo y emulado; con el candado banco.lock, sin renders)

Prefijo gl-prac-. Los contenedores con systemd son --privileged (systemd necesita escribir en su cgroup, como los nodos de kind)
y desechables; se borran al final. ⛔ Lo que diga el vídeo sale de la SALIDA de aquí.
"""
import json
import re
import statistics
import subprocess
import sys
import time
from datetime import datetime
from pathlib import Path

AQUI = Path(__file__).resolve().parent
BIN = AQUI / "_bin"
BUILD_IMG = "gl-prac-rpi-build:1"
SYSD_ARM, SYSD_AMD = "gl-prac-rpi-systemd:1", "gl-prac-rpi-systemd-amd64:1"
OUT = []
MIOS = []


def sh(args, check=False, **kw):
    r = subprocess.run(args, capture_output=True, text=True, encoding="utf-8", errors="replace", **kw)
    if check and r.returncode != 0:
        raise SystemExit(f"falló {' '.join(args)}\n{r.stdout}\n{r.stderr}")
    return r


def w(s=""):
    print(s)
    OUT.append(s)


ANSI = re.compile(r"\x1b\[[0-9;?]*[a-zA-Z]|\x1b\][^\x07]*\x07|\r")


def bloque(texto):
    w("```")
    for linea in ANSI.sub("", texto).rstrip("\n").splitlines():
        w(linea)
    w("```")
    w()


def build(cmd):
    """Un comando en el contenedor de compilación (amd64), con el código montado en /w."""
    return sh(["docker", "run", "--rm", "--ulimit", "core=1", "-v", f"{AQUI}:/w", "-v", "gl-prac-gomod:/go/pkg/mod",
               "-w", "/w/sensor", BUILD_IMG, "sh", "-c", cmd])


def plataforma(plat, imagen, cmd):
    return sh(["docker", "run", "--rm", "--ulimit", "core=1", "--platform", plat, "-v", f"{BIN}:/b", imagen, "sh", "-c", cmd])


def sistema(nombre, imagen, plat):
    sh(["docker", "rm", "-f", nombre])
    MIOS.append(nombre)
    # systemd como proceso 1: --privileged (escribe en su cgroup), -t (su consola), --ulimit core=1:-1 (con core=1 a secas
    # systemd muere al arrancar: «Failed to set RLIMIT_CORE: Operation not permitted»; el límite blando sigue en 1).
    sh(["docker", "run", "-d", "-t", "--name", nombre, "--platform", plat, "--privileged", "--cgroupns=private",
        "--ulimit", "core=1:-1", "--tmpfs", "/run", "--tmpfs", "/run/lock", "-e", "container=docker", imagen], check=True)
    t0 = time.time()
    while time.time() - t0 < 120:
        r = sh(["docker", "exec", nombre, "systemctl", "is-system-running"])
        if r.stdout.strip() in ("running", "degraded"):
            return r.stdout.strip()
        time.sleep(1)
    return "no terminó de arrancar"


def ex(nombre, cmd):
    r = sh(["docker", "exec", nombre, "sh", "-c", cmd])
    return r.stdout + r.stderr


def funcional():
    w("# SALIDA — Go en una Raspberry Pi, EMULADA con QEMU (golandia)\n")
    w(f"> Corrida del {datetime.now():%Y-%m-%d %H:%M}. Regenerable con `python correr.py`. ⛔ Sin placa: los contenedores ARM corren "
      "en un PC x86-64 con el QEMU que registra Docker Desktop (binfmt).\n")
    gv = build("go version").stdout.strip()
    mods = re.findall(r"(periph\.io/x/\S+ v[\d.]+)", (AQUI / "sensor" / "go.mod").read_text(encoding="utf-8"))
    dv = sh(["docker", "version", "--format", "{{.Server.Version}}"]).stdout.strip()
    w(f"- `{gv}` (imagen golang:1.27.2, en amd64) · {' · '.join(mods)} · Docker {dv} · Debian 13 (debian:trixie-slim)\n")

    w("## 1 · ¿Qué CPU ve cada contenedor? (`uname -m`)\n")
    filas = []
    for plat, img in (("linux/amd64", "debian:trixie-slim"), ("linux/arm64", "debian:trixie-slim"),
                      ("linux/arm/v7", "debian:trixie-slim"), ("linux/arm/v6", "alpine:3.22")):
        r = plataforma(plat, img, "uname -m")
        filas.append(f"--platform {plat:<13} ({img}) → {(r.stdout or r.stderr).strip()}")
    bloque("\n".join(filas))

    w("## 2 · Compilación cruzada desde amd64\n")
    r = build("echo '# sin variables:'; go env GOOS GOARCH CGO_ENABLED | paste -sd' '; "
              "echo '# GOOS=linux GOARCH=arm64:'; GOOS=linux GOARCH=arm64 go env GOOS GOARCH CGO_ENABLED | paste -sd' '")
    bloque(r.stdout + r.stderr)
    BIN.mkdir(exist_ok=True)
    cmds = [("amd64 (tu PC)", "GOOS=linux GOARCH=amd64 CGO_ENABLED=0", "sensor-amd64"),
            ("arm64 (Pi 3, 4, 5 y Zero 2 W con sistema de 64 bits)", "GOOS=linux GOARCH=arm64", "sensor-arm64"),
            ("armv7 (Pi 2, o una Pi con sistema de 32 bits)", "GOOS=linux GOARCH=arm GOARM=7", "sensor-armv7"),
            ("armv6 (Pi 1 y Pi Zero / Zero W)", "GOOS=linux GOARCH=arm GOARM=6", "sensor-armv6")]
    lineas = []
    for _, env, sal in cmds:
        c = f"{env} go build -trimpath -ldflags='-s -w' -o /w/_bin/{sal} ."
        r = build(c)
        assert r.returncode == 0, r.stderr
        lineas.append(f"$ {c}")
    bloque("\n".join(lineas))
    r = build("cd /w/_bin && for f in sensor-amd64 sensor-arm64 sensor-armv7 sensor-armv6; do "
              "printf '%s  %s bytes\\n' \"$(file -b $f | cut -d, -f1-2,4)\" \"$(stat -c %s $f)\"; done")
    w("`file` (recortado a arquitectura y enlazado) y tamaño:")
    bloque(r.stdout)
    r = build("cd /w/_bin && file -b sensor-arm64 | tr ',' '\\n' | sed 's/^ //' | grep -E 'ELF|ARM|statically'")
    bloque(r.stdout)
    w("Lo que el binario dice de sí mismo (`go version -m`, solo las líneas de la compilación):")
    r = build("cd /w/_bin && for f in sensor-arm64 sensor-armv7 sensor-armv6; do echo \"# $f\"; "
              "go version -m $f | grep -E 'build\\s+(GOARCH|GOARM|GOOS|CGO_ENABLED)=' | sed 's/^\\s*//'; done")
    bloque(r.stdout + r.stderr)

    w("## 3 · Ejecutar los binarios\n")
    filas = []
    for plat, img, b in (("linux/arm64", "debian:trixie-slim", "sensor-arm64"), ("linux/arm/v7", "debian:trixie-slim", "sensor-armv7"),
                         ("linux/arm/v6", "alpine:3.22", "sensor-armv6"), ("linux/amd64", "debian:trixie-slim", "sensor-arm64")):
        r = plataforma(plat, img, f"uname -m; /b/{b} -calculo 1000; echo rc=$?")
        filas.append(f"# {b} en --platform {plat} ({img})\n" + (r.stdout + r.stderr).strip())
    bloque("\n\n".join(filas))
    w("(La última: el binario ARM64 en un contenedor amd64 NO da «exec format error»: el kernel de Docker Desktop lo pasa a QEMU "
      "por binfmt, sin avisar.)\n")

    w("## 4 · La trampa de cgo\n")
    r = sh(["docker", "run", "--rm", "--ulimit", "core=1", "-v", f"{AQUI}:/w", "-w", "/w/concgo", BUILD_IMG, "sh", "-c",
            "echo '$ GOARCH=arm64 go build -o /w/_bin/concgo-arm64 .'; GOARCH=arm64 go build -o /w/_bin/concgo-arm64 . > /tmp/o 2>&1; "
            "echo rc=$?; head -4 /tmp/o; "
            "echo '$ CGO_ENABLED=1 GOARCH=arm64 go build -o /w/_bin/concgo-arm64 .'; "
            "CGO_ENABLED=1 GOARCH=arm64 go build -o /w/_bin/concgo-arm64 . > /tmp/o 2>&1; echo rc=$?; head -6 /tmp/o; "
            "echo '$ CGO_ENABLED=1 GOARCH=arm64 CC=aarch64-linux-gnu-gcc go build -o /w/_bin/concgo-arm64 .'; "
            "CGO_ENABLED=1 GOARCH=arm64 CC=aarch64-linux-gnu-gcc go build -o /w/_bin/concgo-arm64 . > /tmp/o 2>&1; echo rc=$?; "
            "head -4 /tmp/o; "
            "file -b /w/_bin/concgo-arm64 | tr ',' '\\n' | sed 's/^ //' | grep -E 'ARM|dynamically|interpreter'"])
    bloque(r.stdout + r.stderr)
    filas = []
    for plat, img in (("linux/arm64", "debian:trixie-slim"), ("linux/arm64", "alpine:3.22")):
        r = plataforma(plat, img, "/b/concgo-arm64; echo rc=$?")
        filas.append(f"# concgo-arm64 en {img} ({plat})\n" + (r.stdout + r.stderr).strip())
    bloque("\n\n".join(filas))

    w("## 5 · El servicio con systemd, en una «Pi» emulada (Debian 13 ARM64 con systemd como proceso 1)\n")
    w("La unidad (systemd/sensor.service):")
    bloque((AQUI / "systemd" / "sensor.service").read_text(encoding="utf-8"))
    est = sistema("gl-prac-rpi", SYSD_ARM, "linux/arm64")
    time.sleep(8)
    w(f"`systemctl is-system-running` → {est}. `systemctl status sensor` a los 8 s:")
    bloque(ex("gl-prac-rpi", "systemctl status sensor --no-pager | head -9; echo; systemctl show -p NRestarts sensor"))
    w("`systemctl --failed` (por qué el sistema está «degraded») y `journalctl -u sensor`:")
    bloque(ex("gl-prac-rpi", "systemctl --failed --no-pager --plain | head -8; echo; journalctl -u sensor --no-pager | tail -4"))
    w("La MISMA unidad y el MISMO binario ARM64, con el systemd nativo de un contenedor amd64 (el binario corre por QEMU):")
    sistema("gl-prac-rpi64", SYSD_AMD, "linux/amd64")
    time.sleep(5)
    bloque(ex("gl-prac-rpi64", "systemctl status sensor --no-pager | head -11"))
    sh(["docker", "rm", "-f", "gl-prac-rpi64"])
    w("En la «Pi» emulada, con la unidad sin las líneas de endurecimiento (systemd/sensor-simple.service: solo ExecStart, "
      "Restart=always y RestartSec=2):")
    bloque(ex("gl-prac-rpi", "systemctl disable --now sensor 2>&1; systemctl enable --now sensor-simple 2>&1; sleep 4; "
                             "systemctl status sensor-simple --no-pager | head -11"))
    w("`curl localhost:8080/estado` dentro de la «Pi»:")
    e = ex("gl-prac-rpi", "curl -s localhost:8080/estado")
    try:
        bloque(json.dumps(json.loads(e), indent=1, ensure_ascii=False))
    except ValueError:
        bloque(e)
    pid = ex("gl-prac-rpi", "systemctl show -p MainPID --value sensor-simple").strip()
    w(f"`kill -9 {pid}` (el proceso del sensor) y, 4 s después:")
    ex("gl-prac-rpi", f"kill -9 {pid}")
    time.sleep(4)
    bloque(ex("gl-prac-rpi", "systemctl status sensor-simple --no-pager | sed -n '3p;6p'; echo; "
                             "systemctl show -p MainPID -p NRestarts -p Result sensor-simple"))

    w("## 6 · GPIO: lo que el emulador no tiene\n")
    bloque(ex("gl-prac-rpi", "ls /dev/gpiochip* /dev/gpiomem 2>&1; cat /proc/device-tree/model 2>&1; "
                             "cat /sys/class/thermal/thermal_zone0/temp 2>&1"))
    w("`sensor -gpio=real` (periph.io/x/host: host.Init() y gpioreg.ByName(\"GPIO17\")) en la «Pi» emulada:")
    bloque(ex("gl-prac-rpi", "timeout 20 /usr/local/bin/sensor -gpio=real 2>&1 | tail -6; echo rc=$?"))
    sh(["docker", "rm", "-f", "gl-prac-rpi"])

    w("## 7 · Una imagen de contenedor para la Pi\n")
    (AQUI / "_imagen").mkdir(exist_ok=True)
    (AQUI / "_imagen" / "Dockerfile").write_text("FROM scratch\nARG TARGETARCH\nCOPY sensor-${TARGETARCH} /sensor\n"
                                                 "ENTRYPOINT [\"/sensor\"]\n", encoding="utf-8")
    import shutil
    for a in ("amd64", "arm64"):
        shutil.copy(BIN / f"sensor-{a}", AQUI / "_imagen" / f"sensor-{a}")
    r = sh(["docker", "buildx", "build", "-q", "--platform", "linux/amd64,linux/arm64", "-t", "gl-prac-rpi-sensor:1",
            str(AQUI / "_imagen")])
    w("`docker buildx build --platform linux/amd64,linux/arm64 -t gl-prac-rpi-sensor:1` (FROM scratch + el binario de cada "
      f"arquitectura): rc={r.returncode}")
    r = sh(["docker", "image", "inspect", "gl-prac-rpi-sensor:1", "--format", "{{.Size}}"])
    w(f"\n`docker image inspect .Size`: {int(r.stdout.strip() or 0) / 1048576:.1f} MiB")
    filas = []
    for plat in ("linux/arm64", "linux/amd64"):
        sh(["docker", "rm", "-f", "gl-prac-rpi-img"])
        MIOS.append("gl-prac-rpi-img")
        sh(["docker", "run", "-d", "--name", "gl-prac-rpi-img", "--ulimit", "core=1", "--platform", plat, "-p", "9306:8080",
            "gl-prac-rpi-sensor:1"], check=True)
        time.sleep(3)
        r = sh(["curl", "-s", "http://127.0.0.1:9306/estado"])
        try:
            d = json.loads(r.stdout)
            filas.append(f"--platform {plat} → goarch {d['goarch']}, pin {d['pin']}")
        except ValueError:
            filas.append(f"--platform {plat} → {r.stdout or r.stderr}")
        sh(["docker", "rm", "-f", "gl-prac-rpi-img"])
    bloque("\n".join(filas))
    (AQUI / "SALIDA.md").write_text("\n".join(OUT) + "\n", encoding="utf-8")


def tiempos():
    w("# SALIDA-tiempos — el mismo cálculo, nativo y emulado\n")
    w(f"> Corrida del {datetime.now():%Y-%m-%d %H:%M}, con el candado banco.lock y sin renders. `sensor -calculo 300000000` "
      "(multiplicar y módulo en un bucle), 5 veces cada uno, mediana.\n")
    w("| binario | plataforma | mediana | mín | máx |")
    w("|---|---|---|---|---|")
    for b, plat, img in (("sensor-amd64", "linux/amd64", "debian:trixie-slim"), ("sensor-arm64", "linux/arm64", "debian:trixie-slim"),
                         ("sensor-armv7", "linux/arm/v7", "debian:trixie-slim")):
        ts = []
        for _ in range(5):
            r = plataforma(plat, img, f"/b/{b} -calculo 300000000")
            m = re.search(r"en ([\d.]+)(ms|s)", r.stdout)
            ts.append(float(m.group(1)) / (1000 if m.group(2) == "ms" else 1))
        w(f"| {b} | {plat} | {statistics.median(ts):.2f} s | {min(ts):.2f} s | {max(ts):.2f} s |")
    w()
    (AQUI / "SALIDA-tiempos.md").write_text("\n".join(OUT) + "\n", encoding="utf-8")


def main():
    sh(["docker", "build", "-q", "-t", BUILD_IMG, "-f", str(AQUI / "Dockerfile.build"), str(AQUI)], check=True)
    for plat in ("linux/amd64", "linux/arm64", "linux/arm/v7", "linux/arm/v6"):  # que no salga el «Pulling…» en la SALIDA
        for img in ("debian:trixie-slim", "alpine:3.22"):
            sh(["docker", "pull", "-q", "--platform", plat, img])
    try:
        if "--tiempos" in sys.argv:
            tiempos()
        else:
            build("GOOS=linux GOARCH=arm64 go build -trimpath -ldflags='-s -w' -o /w/_bin/sensor-arm64 .")
            for img, plat in ((SYSD_ARM, "linux/arm64"), (SYSD_AMD, "linux/amd64")):
                sh(["docker", "build", "-q", "--platform", plat, "-t", img, "-f", str(AQUI / "systemd" / "Dockerfile"), str(AQUI)],
                   check=True)
            funcional()
    finally:
        for n in dict.fromkeys(MIOS):
            sh(["docker", "rm", "-f", n])
    print("\n[ok]")


if __name__ == "__main__":
    main()
