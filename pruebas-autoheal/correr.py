# -*- coding: utf-8 -*-
"""Demo del vídeo «Autoheal en Go» (golandia): escenarios medidos con tiempos.

    python correr.py [escenario ...]     escenarios: sin-vigia eventos sondeo freno etiqueta  (por defecto, todos)

Solo crea y borra contenedores con nombre gh-* y la etiqueta golandia.autoheal; nunca toca otros.
Cada escenario deja su salida en salidas/<escenario>.txt; SALIDA.md se escribe a mano a partir de ellas.
"""
import os
import subprocess
import sys
import time
from pathlib import Path

AQUI = Path(__file__).parent
SAL = AQUI / "salidas"
SAL.mkdir(exist_ok=True)
ENV = {**os.environ, "MSYS_NO_PATHCONV": "1"}
IMG = "nginx:1.30.5-alpine"
ETQ = "golandia.autoheal=demo"
VIGIA_DIR = subprocess.run(["cygpath", "-w", str(AQUI / "vigia")], capture_output=True, text=True).stdout.strip() or str(AQUI / "vigia")
# Healthcheck honesto: pide la página de verdad. Cada 2 s, 2 s de espera, 3 fallos seguidos = unhealthy.
SALUD = ["--health-cmd", "wget -q -T 1 -O /dev/null http://127.0.0.1/ || exit 1",
         "--health-interval", "2s", "--health-timeout", "2s", "--health-retries", "3"]


def sh(*a, check=True, crudo=False):
    r = subprocess.run(list(a), capture_output=True, text=True, env=ENV, encoding="utf-8", errors="replace")
    if check and r.returncode != 0:
        raise SystemExit(f"falló {' '.join(a)}:\n{r.stdout}{r.stderr}")
    return r.stdout.rstrip() if crudo else r.stdout.strip()


def limpiar():
    for n in ("gh-app", "gh-roto", "gh-ajeno", "gh-vigia"):
        sh("docker", "rm", "-f", n, check=False)


def app(nombre, etiqueta=True, roto=False):
    args = ["docker", "run", "-d", "--name", nombre, "--init", "--restart", "unless-stopped", *SALUD]
    if etiqueta:
        args += ["--label", ETQ]
    if roto:
        # Arranca, pero escucha en otro puerto: el healthcheck falla siempre, también después de cada reinicio.
        args += [IMG, "sh", "-c", "sed -i 's/listen  *80;/listen 8080;/' /etc/nginx/conf.d/default.conf && exec nginx -g 'daemon off;'"]
    else:
        args += [IMG]
    sh(*args)


def estado(nombre):
    return sh("docker", "inspect", "-f", "{{.State.Status}} {{.State.Health.Status}} {{.RestartCount}} {{.State.StartedAt}}", nombre).split()


def esperar_sano(nombre, tope=30):
    t = time.time()
    while time.time() - t < tope:
        if estado(nombre)[1] == "healthy":
            return
        time.sleep(0.5)
    raise SystemExit(f"{nombre} no llegó a healthy")


def colgar(nombre):
    # El proceso sigue vivo pero ya no responde: SIGSTOP a nginx. Por eso va con --init (nginx no es el PID 1).
    sh("docker", "exec", nombre, "sh", "-c", "kill -STOP $(pidof nginx)")


def vigia(*flags):
    sh("docker", "run", "-d", "--name", "gh-vigia", "-v", "/var/run/docker.sock:/var/run/docker.sock",
       "-v", f"{VIGIA_DIR}:/app", "--entrypoint", "/app/vigia", IMG, "-etiqueta", ETQ, *flags)


def observar(nombres, segundos, t0, out, hasta=None):
    """Mira el estado cada 0,5 s y apunta solo los cambios, con el tiempo desde t0."""
    previo = {}
    fin = time.time() + segundos
    while time.time() < fin:
        for n in nombres:
            s = estado(n)
            clave = (s[0], s[1], s[3])
            if previo.get(n) != clave:
                # RestartCount solo cuenta los reinicios de la política de Docker; StartedAt cambia con cualquiera.
                out.append(f"{time.time() - t0:6.1f}s  {n:9} {s[0]:8} {s[1]:10} RestartCount={s[2]}  arrancó {s[3][11:19]}")
                previo[n] = clave
        if hasta and hasta():
            break
        time.sleep(0.5)


def escribir(nombre, lineas):
    (SAL / f"{nombre}.txt").write_text("\n".join(lineas) + "\n", encoding="utf-8", newline="\n")
    print("\n".join(lineas))
    print(f"[ok] salidas/{nombre}.txt\n")


def sin_vigia():
    limpiar()
    app("gh-app")
    esperar_sano("gh-app")
    out = [f"$ docker run -d --name gh-app --restart unless-stopped --init {' '.join(SALUD[:2])} … {IMG}", "$ (cuelgo nginx: kill -STOP)"]
    t0 = time.time()
    colgar("gh-app")
    observar(["gh-app"], 60, t0, out)
    out.append(f"{time.time() - t0:6.1f}s  fin de la observación · {' '.join(estado('gh-app')[:3])}")
    limpiar()
    escribir("sin-vigia", out)


def con_vigia(nombre, flags):
    limpiar()
    app("gh-app")
    esperar_sano("gh-app")
    vigia(*flags)
    time.sleep(2)
    out = [f"$ vigia {' '.join(flags)} -etiqueta {ETQ}", "$ (cuelgo nginx: kill -STOP)"]
    t0 = time.time()
    colgar("gh-app")
    arranque = estado("gh-app")[3]
    curado = lambda: estado("gh-app")[1] == "healthy" and estado("gh-app")[3] != arranque  # noqa: E731
    observar(["gh-app"], 60, t0, out, hasta=curado)
    out.append(f"{time.time() - t0:6.1f}s  sano otra vez")
    out.append("--- log del vigía (su reloj empieza cuando arranca él, ~2 s antes del cuelgue) ---")
    out += sh("docker", "logs", "gh-vigia", crudo=True).splitlines()
    limpiar()
    escribir(nombre, out)


def freno():
    limpiar()
    app("gh-roto", roto=True)
    vigia("-modo", "eventos", "-max", "3", "-ventana", "2m")
    out = ["$ gh-roto: nginx escucha en el 8080 y el healthcheck pide el 80 → enfermo siempre",
           "$ vigia -modo eventos -max 3 -ventana 2m"]
    t0 = time.time()
    observar(["gh-roto"], 70, t0, out)
    out.append(f"{time.time() - t0:6.1f}s  fin · {' '.join(estado('gh-roto')[:3])}")
    out.append("--- log del vigía ---")
    out += sh("docker", "logs", "gh-vigia", crudo=True).splitlines()
    limpiar()
    escribir("freno", out)


def etiqueta():
    limpiar()
    app("gh-app")
    app("gh-ajeno", etiqueta=False)
    esperar_sano("gh-app")
    esperar_sano("gh-ajeno")
    vigia("-modo", "eventos")
    time.sleep(2)
    out = ["$ gh-app (con la etiqueta) y gh-ajeno (sin ella): cuelgo los dos"]
    t0 = time.time()
    colgar("gh-app")
    colgar("gh-ajeno")
    observar(["gh-app", "gh-ajeno"], 40, t0, out)
    out.append(f"{time.time() - t0:6.1f}s  fin · gh-app {' '.join(estado('gh-app')[:3])} · gh-ajeno {' '.join(estado('gh-ajeno')[:3])}")
    out.append("--- log del vigía ---")
    out += sh("docker", "logs", "gh-vigia", crudo=True).splitlines()
    limpiar()
    escribir("etiqueta", out)


ESC = {"sin-vigia": sin_vigia,
       "eventos": lambda: con_vigia("eventos", ["-modo", "eventos"]),
       "sondeo": lambda: con_vigia("sondeo", ["-modo", "sondeo", "-cada", "10s"]),
       "freno": freno, "etiqueta": etiqueta}

if __name__ == "__main__":
    pedidos = sys.argv[1:] or list(ESC)
    try:
        for e in pedidos:
            print(f"== {e}")
            ESC[e]()
    finally:
        limpiar()
