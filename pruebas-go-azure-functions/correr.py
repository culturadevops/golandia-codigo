# -*- coding: utf-8 -*-
"""correr.py — «Go en Azure Functions con custom handler», TODO EN LOCAL (golandia). ⛔ Nada en la nube: no hay permiso.

    python correr.py              → SALIDA.md (lo funcional: sin cronómetros)
    python correr.py --tiempos    → SALIDA-tiempos.md (arranque y latencia en local; con el candado banco.lock y sin renders)

Piezas (prefijo gl-prac-, red con --subnet porque Docker no tiene rangos libres):
  · Core Tools 4.17.0 (npm) en gl-prac-functools:1: `func init --worker-runtime custom --docker` y `func start`.
  · El host OFICIAL en contenedor: mcr.microsoft.com/azure-functions/dotnet:4.1036.0.1 + el handler de Go (gl-prac-azfunc:1).
  · Azurite 3.38.0 para la cola (disparador) y los blobs (binding de salida); cliente Python con azure-storage-* fijados.
⛔ Lo que diga el vídeo sale de la SALIDA de aquí.
"""
import json
import re
import statistics
import subprocess
import sys
import time
import urllib.request
from datetime import datetime
from pathlib import Path

AQUI = Path(__file__).resolve().parent
RED, SUBRED = "gl-prac-azfunc", "10.213.48.0/24"
HOST_IMG, SOLO_IMG = "gl-prac-azfunc:1", "gl-prac-azfunc-sologo:1"
TOOLS_IMG, CLI_IMG = "gl-prac-functools:1", "gl-prac-azcliente:1"
AZURITE_IMG = "mcr.microsoft.com/azure-storage/azurite:3.38.0"
BASE_HOST = "mcr.microsoft.com/azure-functions/dotnet:4.1036.0.1"
# La cuenta de desarrollo de Azurite: clave pública y documentada, la misma en todas las instalaciones.
CS = ("DefaultEndpointsProtocol=http;AccountName=devstoreaccount1;"
      "AccountKey=Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw==;"
      "BlobEndpoint=http://gl-prac-azurite:10000/devstoreaccount1;QueueEndpoint=http://gl-prac-azurite:10001/devstoreaccount1;"
      "TableEndpoint=http://gl-prac-azurite:10002/devstoreaccount1;")
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


def bloque(texto):
    w("```")
    for linea in texto.rstrip("\n").splitlines():
        w(linea)
    w("```")
    w()


ANSI = re.compile(r"\x1b\[[0-9;]*m")


def logs(nombre):
    r = sh(["docker", "logs", nombre])
    return ANSI.sub("", r.stdout + r.stderr)


def correr(nombre, imagen, *cmd, env=None, puerto=None, limites=True, extra=()):
    sh(["docker", "rm", "-f", nombre])
    args = ["docker", "run", "-d", "--name", nombre, "--network", RED, "--ulimit", "core=1", *extra]
    if limites:
        args += ["--cpus", "2", "--memory", "1g"]
    if puerto:
        args += ["-p", f"{puerto}:{80 if imagen == HOST_IMG else 8080}"]
    for k, v in (env or {}).items():
        args += ["-e", f"{k}={v}"]
    MIOS.append(nombre)
    return sh(args + [imagen, *cmd], check=True)


def cli(*args):
    return sh(["docker", "run", "--rm", "--network", RED, "--ulimit", "core=1", "-e", f"AzureWebJobsStorage={CS}", CLI_IMG, *args],
              check=True).stdout


def get(url, timeout=5):
    with urllib.request.urlopen(url, timeout=timeout) as r:
        return r.status, r.read().decode()


def esperar_200(url, seg=120):
    t0 = time.perf_counter()
    while time.perf_counter() - t0 < seg:
        try:
            if get(url, 1)[0] == 200:
                return (time.perf_counter() - t0) * 1000
        except Exception:
            pass
        time.sleep(0.01)
    raise SystemExit(f"no respondió {url}")


def procesos(nombre):
    """pid, comando y VmRSS de cada proceso del contenedor (la imagen del host no trae ps)."""
    r = sh(["docker", "exec", nombre, "sh", "-c",
            'for p in /proc/[0-9]*; do c=$(cat $p/comm 2>/dev/null); r=$(grep VmRSS $p/status 2>/dev/null | tr -s " \t" " "); '
            'echo "${p#/proc/} $c $r"; done'])
    return [x for x in r.stdout.splitlines() if x.strip()]


def preparar():
    sh(["docker", "network", "create", "--subnet", SUBRED, RED])
    for img, ctx, df in ((HOST_IMG, AQUI, "Dockerfile"), (SOLO_IMG, AQUI, "Dockerfile.solo-go"),
                         (TOOLS_IMG, AQUI / "functools", "Dockerfile"), (CLI_IMG, AQUI / "cliente", "Dockerfile")):
        sh(["docker", "build", "-q", "-t", img, "-f", str(Path(ctx) / df), str(ctx)], check=True)
    correr("gl-prac-azurite", AZURITE_IMG, "azurite", "--blobHost", "0.0.0.0", "--queueHost", "0.0.0.0", "--tableHost", "0.0.0.0",
           "--skipApiVersionCheck", "--loose", limites=False)
    time.sleep(3)


def funcional():
    w("# SALIDA — Go en Azure Functions con custom handler, en LOCAL (golandia)\n")
    w(f"> Corrida del {datetime.now():%Y-%m-%d %H:%M}. Regenerable con `python correr.py`. ⛔ Nada en la nube real: el host oficial "
      "de Functions en un contenedor, Core Tools y Azurite, todo en esta máquina.\n")
    fv = sh(["docker", "run", "--rm", TOOLS_IMG, "func", "--version"]).stdout.strip()
    hv = sh(["docker", "image", "inspect", "-f", "{{range .Config.Env}}{{println .}}{{end}}", BASE_HOST]).stdout
    hv = re.search(r"HOST_VERSION=(\S+)", hv).group(1)
    gv = sh(["docker", "run", "--rm", "golang:1.27.2", "go", "version"]).stdout.strip()
    av = sh(["docker", "run", "--rm", AZURITE_IMG, "azurite", "--version"]).stdout.strip()
    w(f"- Core Tools `func --version`: {fv} · host: `{BASE_HOST}` (HOST_VERSION={hv} en su entorno) · Azurite {av} · `{gv}`")
    w("- Handler: net/http y nada más (go/main.go), compilado estático (CGO_ENABLED=0).\n")

    # 1 · lo que genera func init
    w("## 1 · Lo que genera `func init proy --worker-runtime custom --docker` (Core Tools 4.17.0)\n")
    r = sh(["docker", "run", "--rm", TOOLS_IMG, "sh", "-c",
            "cd /tmp && func init proy --worker-runtime custom --docker >/dev/null 2>&1; cd proy && ls -A && echo '--- Dockerfile' && "
            "cat Dockerfile && echo && echo '--- host.json (customHandler)' && sed -n '/customHandler/,$p' host.json"])
    bloque(r.stdout)
    w("¿Construye tal cual? (`docker build` del Dockerfile generado, en una carpeta temporal del contenedor de Core Tools no hay "
      "Docker: se copia aquí a _generado/):\n")
    gen = AQUI / "_generado"
    gen.mkdir(exist_ok=True)
    r = sh(["docker", "run", "--rm", TOOLS_IMG, "sh", "-c",
            "cd /tmp && func init proy --worker-runtime custom --docker >/dev/null 2>&1 && cat proy/Dockerfile"])
    (gen / "Dockerfile").write_text(r.stdout, encoding="utf-8")
    b = sh(["docker", "build", "-t", "gl-prac-azfunc-generado:1", str(gen)])
    err = [x for x in (b.stdout + b.stderr).splitlines() if re.search(r"error|ERROR|failed", x)]
    bloque("\n".join(err[:4]) if err else "(construyó sin error)")
    sh(["docker", "rmi", "gl-prac-azfunc-generado:1"])

    # 2 · func start
    w("## 2 · `func start` (Core Tools) con el handler de Go\n")
    hb = sh(["docker", "run", "--rm", "--entrypoint", "cat", HOST_IMG, "/home/site/wwwroot/host.json"])
    w("host.json del proyecto (el mismo en `func start` y en la imagen):")
    bloque(hb.stdout)
    sh(["docker", "rm", "-f", "gl-prac-funcstart"])
    MIOS.append("gl-prac-funcstart")
    # el handler compilado sale de la imagen del host; func start lo arranca desde la carpeta del proyecto
    sh(["docker", "create", "--name", "gl-prac-tmpcopia", HOST_IMG], check=True)
    sh(["docker", "cp", "gl-prac-tmpcopia:/home/site/wwwroot", str(AQUI / "_wwwroot")], check=True)
    sh(["docker", "rm", "-f", "gl-prac-tmpcopia"])
    (AQUI / "_wwwroot" / "local.settings.json").write_text(json.dumps(
        {"IsEncrypted": False, "Values": {"FUNCTIONS_WORKER_RUNTIME": "custom", "AzureWebJobsStorage": CS}}, indent=2), encoding="utf-8")
    sh(["docker", "run", "-d", "--name", "gl-prac-funcstart", "--network", RED, "--ulimit", "core=1", "-p", "9303:7071",
        "-v", f"{AQUI / '_wwwroot'}:/proy", "-w", "/proy", TOOLS_IMG, "sh", "-c", "chmod +x handler && func start"], check=True)
    # ⚠️ func start escucha en localhost DENTRO del contenedor: desde fuera, el puerto publicado da «empty reply». Se le
    # pregunta desde dentro, con el fetch de node.
    def dentro():
        return sh(["docker", "exec", "gl-prac-funcstart", "node", "-e",
                   "fetch('http://127.0.0.1:7071/api/hola?nombre=func-start').then(r=>r.text()).then(t=>console.log(t))"
                   ".catch(()=>process.exit(1))"])
    t0 = time.time()
    while dentro().returncode != 0:
        if time.time() - t0 > 180:
            raise SystemExit("func start no respondió")
        time.sleep(1)
    l = logs("gl-prac-funcstart")
    w("Salida de `func start` (las líneas de las funciones y del handler):")
    bloque("\n".join(x.strip() for x in l.splitlines()
                     if re.search(r"Functions:|hola:|procesar:|Worker process started|\[handler\]|Host started|functions found|"
                                  r"Core Tools Version|Function Runtime Version", x)))
    w("`/api/hola?nombre=func-start`, pedido desde dentro del contenedor →")
    bloque(dentro().stdout)
    sh(["docker", "rm", "-f", "gl-prac-funcstart"])

    w("### 2b · `func start` con AzureWebJobsStorage vacío (solo la función HTTP importa… ¿o no?)\n")
    (AQUI / "_wwwroot" / "local.settings.json").write_text(json.dumps(
        {"IsEncrypted": False, "Values": {"FUNCTIONS_WORKER_RUNTIME": "custom", "AzureWebJobsStorage": ""}}, indent=2), encoding="utf-8")
    sh(["docker", "run", "-d", "--name", "gl-prac-funcstart", "--network", RED, "--ulimit", "core=1",
        "-v", f"{AQUI / '_wwwroot'}:/proy", "-w", "/proy", TOOLS_IMG, "sh", "-c", "chmod +x handler && func start"], check=True)
    time.sleep(40)
    l = logs("gl-prac-funcstart")
    ok = dentro().returncode == 0
    bloque("\n".join(x.strip() for x in l.splitlines() if re.search(r"error|Secret|Value cannot be null", x, re.I))[:1200]
           + f"\n\n/api/hola responde: {'sí' if ok else 'no'}")
    sh(["docker", "rm", "-f", "gl-prac-funcstart"])

    # 3 · el host oficial en contenedor
    w("## 3 · El host oficial en un contenedor (2 CPU, 1 GiB), con Azurite\n")
    correr("gl-prac-azfunc-host", HOST_IMG, env={"AzureWebJobsStorage": CS}, puerto=9302)
    esperar_200("http://127.0.0.1:9302/api/hola?nombre=a")
    w("Tres peticiones seguidas a /api/hola:")
    bloque("\n".join(get(f"http://127.0.0.1:9302/api/hola?nombre=golandia-{i}")[1].strip() for i in (1, 2, 3)))
    w("Procesos dentro del contenedor (pid, nombre, memoria residente):")
    bloque("\n".join(procesos("gl-prac-azfunc-host")))
    l = logs("gl-prac-azfunc-host")
    m = re.search(r"\[handler\] escuchando en :(\d+)", l)
    w(f"El handler escucha en el puerto que le da el host: {m.group(1) if m else '?'} (FUNCTIONS_CUSTOMHANDLER_PORT; el host "
      "escucha en el 80 y reenvía).\n")

    # 4 · cola → blob
    w("## 4 · Disparador de cola (Azurite) → handler de Go → blob (binding de salida)\n")
    w("```")
    w(cli("cola", "pedidos").strip())
    w(cli("mandar", "pedidos", '{"id":"p-42","total":1000}').strip())
    w("```\n")
    t0 = time.time()
    while time.time() - t0 < 60:
        b = cli("blobs", "facturas")
        if ".json" in b:
            break
        time.sleep(1)
    w("Lo que escribió el handler en su log (vía el host):")
    l = logs("gl-prac-azfunc-host")
    bloque("\n".join(x.strip() for x in l.splitlines() if "[handler]" in x and "escuchando" not in x
                     or re.search(r"Execut(ing|ed) 'Functions\.procesar'", x)))
    w("Blobs en el contenedor «facturas» (los escribió el HOST con lo que devolvió el handler en Outputs.factura):")
    bloque(b)

    # 5 · el handler muere
    w("## 5 · ¿Y si el proceso de Go muere? (`kill -9` al handler dentro del contenedor)\n")
    antes = get("http://127.0.0.1:9302/api/hola?nombre=antes")[1].strip()
    pid = json.loads(antes)["pid"]
    sh(["docker", "exec", "gl-prac-azfunc-host", "sh", "-c", f"kill -9 {pid}"], check=True)
    tk = time.time()
    resultados = []
    while time.time() - tk < 30:
        try:
            st, cuerpo = get("http://127.0.0.1:9302/api/hola?nombre=despues", 10)
            resultados.append(f"+{time.time() - tk:.2f} s → {st} {cuerpo.strip()}")
            if json.loads(cuerpo)["pid"] != pid:
                break
        except Exception as e:
            resultados.append(f"+{time.time() - tk:.2f} s → {type(e).__name__}: {e}")
        time.sleep(0.2)
    bloque(f"antes: {antes}\nkill -9 {pid}\n" + "\n".join(resultados[:12]))
    l = logs("gl-prac-azfunc-host")
    bloque("\n".join(x.strip() for x in l.splitlines() if re.search(r"exited|WorkerProcessExitException|escuchando", x))[:1500])

    # 6 · tamaños
    w("## 6 · Lo que pesa cada imagen (`docker image inspect .Size`)\n")
    w("| imagen | tamaño |")
    w("|---|---|")
    for n, im in (("host oficial + handler de Go", HOST_IMG), ("base del host sola", BASE_HOST), ("el handler solo (FROM scratch)", SOLO_IMG)):
        s = int(sh(["docker", "image", "inspect", "-f", "{{.Size}}", im]).stdout.strip())
        w(f"| {n} (`{im}`) | {s / 1048576:.1f} MiB |")
    hsz = sh(["docker", "run", "--rm", "--entrypoint", "ls", HOST_IMG, "-l", "/home/site/wwwroot/handler"]).stdout.strip()
    w(f"\nEl binario: `{hsz}`\n")
    (AQUI / "SALIDA.md").write_text("\n".join(OUT) + "\n", encoding="utf-8")


def hey_lat(url, seg=20):
    r = sh(["docker", "run", "--rm", "--network", RED, "--cpus", "4", "golandia-revancha-hey:1", "-z", f"{seg}s", "-c", "1", "-q", "5",
            url])
    t = r.stdout
    return {p: float(re.search(rf"\s{p}% in ([\d.]+) secs", t).group(1)) * 1000 for p in (50, 99)}


def tiempos():
    w("# SALIDA-tiempos — Go en Azure Functions con custom handler, en LOCAL\n")
    w(f"> Corrida del {datetime.now():%Y-%m-%d %H:%M}, con el candado banco.lock y sin renders. 2 CPU y 1 GiB por contenedor. ⛔ Es "
      "un arranque EN LOCAL, con la imagen ya descargada: no es el arranque en frío de Azure.\n")
    w("## T1 · Contenedor nuevo → primera respuesta 200 de /api/hola (mediana de 8) y memoria en reposo\n")
    w("| variante | 1ª respuesta (docker incluido) | mín | máx | memoria del contenedor en reposo |")
    w("|---|---|---|---|---|")
    for n, im, env in (("host oficial + handler de Go", HOST_IMG, {"AzureWebJobsStorage": CS}),
                       ("el handler solo, sin host", SOLO_IMG, {})):
        ts, mems = [], []
        for _ in range(8):
            sh(["docker", "rm", "-f", "gl-prac-azt"])
            t0 = time.perf_counter()
            correr("gl-prac-azt", im, env=env, puerto=9304)
            esperar_200("http://127.0.0.1:9304/api/hola?nombre=t")
            ts.append((time.perf_counter() - t0) * 1000)
            time.sleep(2)
            s = sh(["docker", "stats", "--no-stream", "--format", "{{.MemUsage}}", "gl-prac-azt"]).stdout.split("/")[0].strip()
            v = float(re.match(r"[\d.]+", s).group(0))
            mems.append(v / 1024 if "KiB" in s else v * 1024 if "GiB" in s else v)
        sh(["docker", "rm", "-f", "gl-prac-azt"])
        w(f"| {n} | {statistics.median(ts):.0f} ms | {min(ts):.0f} ms | {max(ts):.0f} ms | {statistics.median(mems):.1f} MiB |")
    w("\n(«1ª respuesta» cuenta desde que se lanza `docker run` hasta el primer 200: incluye crear el contenedor.)\n")
    w("## T2 · Latencia de una petición sola, en caliente: un cliente a 5 pet/s durante 20 s (`hey -c 1 -q 5`)\n")
    w("| variante | p50 | p99 |")
    w("|---|---|---|")
    for n, im, env, nombre in (("host oficial + handler de Go", HOST_IMG, {"AzureWebJobsStorage": CS}, "gl-prac-azt"),
                               ("el handler solo, sin host", SOLO_IMG, {}, "gl-prac-azt")):
        correr(nombre, im, env=env, puerto=9304)
        esperar_200("http://127.0.0.1:9304/api/hola?nombre=t")
        puerto = 80 if im == HOST_IMG else 8080
        url = f"http://{nombre}:{puerto}/api/hola?nombre=t"
        sh(["docker", "run", "--rm", "--network", RED, "golandia-revancha-hey:1", "-z", "5s", "-c", "4", url])  # calentar
        lat = hey_lat(url)
        sh(["docker", "rm", "-f", nombre])
        w(f"| {n} | {lat[50]:.2f} ms | {lat[99]:.2f} ms |")
    w()
    (AQUI / "SALIDA-tiempos.md").write_text("\n".join(OUT) + "\n", encoding="utf-8")


def contenido():
    """`--contenido` → SALIDA-contenido.md: de qué están hechos los 454 MiB de la imagen (du dentro de ella)."""
    w("# SALIDA-contenido — qué hay dentro de la imagen del host\n")
    w(f"> Corrida del {datetime.now():%Y-%m-%d %H:%M}. `du -sh` dentro de `{HOST_IMG}` (base `{BASE_HOST}`).\n")
    r = sh(["docker", "run", "--rm", "--entrypoint", "sh", HOST_IMG, "-c",
            "du -sh /azure-functions-host /FuncExtensionBundles /home/site/wwwroot /home/site/wwwroot/handler"])
    bloque(r.stdout + r.stderr)
    (AQUI / "SALIDA-contenido.md").write_text("\n".join(OUT) + "\n", encoding="utf-8")


def main():
    if "--contenido" in sys.argv:
        contenido()
        return
    preparar()
    try:
        if "--tiempos" in sys.argv:
            tiempos()
        else:
            funcional()
    finally:
        for n in dict.fromkeys(MIOS):
            sh(["docker", "rm", "-f", n])
        sh(["docker", "network", "rm", RED])
        import shutil
        shutil.rmtree(AQUI / "_wwwroot", ignore_errors=True)
    print("\n[ok]")


if __name__ == "__main__":
    main()
