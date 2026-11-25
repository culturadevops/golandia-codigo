# -*- coding: utf-8 -*-
"""Corre las 5 partes del curso «Go práctico 2027» en Docker y escribe parteN/SALIDA.md.

La regla de la casa: nada de lo que sale en el vídeo se escribe a mano. Sale de aquí.
Solo hace falta Docker y Python 3.10+ (Go va dentro de la imagen oficial golang:1.27.2,
Postgres en postgres:18.6-alpine).

    python ejecutar_todo.py              # las 5 partes
    python ejecutar_todo.py 2 3          # solo esas partes

Todo lo que crea lleva el prefijo gl-curso- y se borra por nombre al terminar
(los volúmenes de caché de Go gl-curso-mod y gl-curso-cache se quedan para la siguiente corrida;
`python ejecutar_todo.py --limpiar` los borra también).
"""
import subprocess
import sys
import time
from datetime import datetime, timezone
from pathlib import Path

AQUI = Path(__file__).resolve().parent
SRC = AQUI.as_posix()
GO = "golang:1.27.2"
PG = "postgres:18.6-alpine"
RED = "gl-curso-red"
PGN = "gl-curso-pg"
URL = f"postgres://tareas:tareas@{PGN}:5432/tareas?sslmode=disable"
CACHE = ["-v", "gl-curso-mod:/go/pkg/mod", "-v", "gl-curso-cache:/root/.cache/go-build"]
BASE = ["docker", "run", "--rm", "--ulimit", "core=1"]


def run(args, timeout=1800):
    r = subprocess.run(args, capture_output=True, text=True, encoding="utf-8", errors="replace", timeout=timeout)
    return r.returncode, (r.stdout + r.stderr).replace("\r", "")


def en_go(parte, red=False, env=None, nombre=None):
    """Corre parteN/demo.sh dentro de golang:1.27.2."""
    args = BASE + ["--name", nombre or f"gl-curso-go-p{parte}", "-v", f"{SRC}:/src:ro", *CACHE]
    if red:
        args += ["--network", RED]
    for k, v in (env or {}).items():
        args += ["-e", f"{k}={v}"]
    args += [GO, "bash", f"/src/parte{parte}/demo.sh"]
    rc, out = run(args)
    return out + f"\n[demo.sh terminó con código {rc}]\n"


def host(texto, args, timeout=1800):
    """Un comando del anfitrión: se imprime como «$ texto» y su salida real."""
    rc, out = run(args, timeout)
    return f"$ {texto}\n{out.rstrip()}\n" + (f"[código de salida: {rc}]\n" if rc else "")


def levantar_postgres():
    run(["docker", "network", "create", "--subnet", "10.231.0.0/24", RED])
    run(BASE + ["-d", "--name", PGN, "--network", RED, "-e", "POSTGRES_USER=tareas",
                "-e", "POSTGRES_PASSWORD=tareas", "-e", "POSTGRES_DB=tareas", PG])
    for _ in range(120):
        rc, _ = run(["docker", "exec", PGN, "psql", "-U", "tareas", "-h", "127.0.0.1", "-c", "select 1"])
        if rc == 0:
            return
        time.sleep(0.5)
    raise SystemExit("Postgres no arrancó")


def bajar_postgres():
    run(["docker", "rm", "-f", PGN])
    run(["docker", "network", "rm", RED])


def parte1():
    return en_go(1)


def con_postgres(parte, env=None):
    levantar_postgres()
    try:
        return en_go(parte, red=True, env={"DATABASE_URL": URL, **(env or {})})
    finally:
        bajar_postgres()


def parte2():
    return con_postgres(2)


def escribir(parte, salida):
    _, versiones = run(["docker", "version", "--format", "Docker {{.Server.Version}}"])
    sello = datetime.now(timezone.utc).strftime("%Y-%m-%d %H:%M:%SZ")
    md = [f"# SALIDA — Go práctico 2027 · parte {parte}", "",
          "> Generado por `ejecutar_todo.py`. **No se escribe a mano.** Ids, tiempos y tamaños son los de esta corrida.", "",
          f"- fecha: `{sello}`", f"- imagen Go: `{GO}` · Postgres: `{PG}` · `{versiones.strip()}`", ""]
    trozos = salida.split("@@PASO ")
    if trozos[0].strip():
        md += ["```", trozos[0].strip("\n"), "```", ""]
    for t in trozos[1:]:
        titulo, _, cuerpo = t.partition("\n")
        md += [f"## {titulo.strip()}", "", "```", cuerpo.strip("\n"), "```", ""]
    (AQUI / f"parte{parte}" / "SALIDA.md").write_text("\n".join(md), encoding="utf-8", newline="\n")
    print(f"  parte{parte}/SALIDA.md ({len(trozos) - 1} pasos)")


def parte3():
    return con_postgres(3, {"BD_PRUEBAS": URL})


def cliente(red, texto):
    """curl desde un contenedor en la misma red (la API no publica puertos en el anfitrión)."""
    rc, out = run(BASE + ["--network", red, GO, "bash", "-c", texto])
    return f"$ {texto}\n{out.rstrip()}\n" + (f"[código de salida: {rc}]\n" if rc else "")


def esperar_http(red, url, segundos=60):
    for _ in range(segundos * 2):
        rc, _ = run(BASE + ["--network", red, GO, "curl", "-sf", "-o", "/dev/null", url])
        if rc == 0:
            return True
        time.sleep(0.5)
    return False


def parte4():
    salida = con_postgres(4)
    ctx = (AQUI / "parte4" / "tareas").as_posix()
    s = ["@@PASO imagenes"]
    s.append(host("docker build -q -f Dockerfile.una-etapa -t gl-curso-tareas:una-etapa .",
                  ["docker", "build", "-q", "-f", f"{ctx}/Dockerfile.una-etapa", "-t", "gl-curso-tareas:una-etapa", ctx]))
    s.append(host("docker build -q -t gl-curso-tareas:p4 .",
                  ["docker", "build", "-q", "-f", f"{ctx}/Dockerfile", "-t", "gl-curso-tareas:p4", ctx]))
    fmt = "{{.Size}} bytes · usuario: '{{.Config.User}}'"
    for tag in ("una-etapa", "p4"):
        s.append(host(f'docker image inspect gl-curso-tareas:{tag} --format "{fmt}"',
                      ["docker", "image", "inspect", f"gl-curso-tareas:{tag}", "--format", fmt]))
    s.append(host('docker image ls gl-curso-tareas --format "{{.Tag}}  {{.Size}}"',
                  ["docker", "image", "ls", "gl-curso-tareas", "--format", "{{.Tag}}  {{.Size}}"]))
    s.append(host('docker history gl-curso-tareas:p4 --format "{{.Size}}  {{.CreatedBy}}"',
                  ["docker", "history", "gl-curso-tareas:p4", "--format", "{{.Size}}  {{.CreatedBy}}"]))
    s.append(host("docker run --rm gl-curso-tareas:p4",
                  BASE + ["gl-curso-tareas:p4"]))
    s.append(host("docker run --rm --entrypoint sh gl-curso-tareas:p4",
                  BASE + ["--entrypoint", "sh", "gl-curso-tareas:p4"]))

    s.append("@@PASO contenedor")
    levantar_postgres()
    try:
        s.append(f"$ docker run -d --rm --name gl-curso-api-p4 --network {RED} -e DATABASE_URL=... gl-curso-tareas:p4\n")
        run(BASE + ["-d", "--name", "gl-curso-api-p4", "--network", RED, "-e", f"DATABASE_URL={URL}", "gl-curso-tareas:p4"])
        esperar_http(RED, "http://gl-curso-api-p4:8080/salud")
        s.append(cliente(RED, "curl -s -X POST gl-curso-api-p4:8080/tareas -d '{\"titulo\":\"desde Docker\"}'"))
        logs = subprocess.Popen(["docker", "logs", "-f", "gl-curso-api-p4"], stdout=subprocess.PIPE,
                                stderr=subprocess.STDOUT, text=True, encoding="utf-8", errors="replace")
        espera = subprocess.Popen(["docker", "wait", "gl-curso-api-p4"], stdout=subprocess.PIPE,
                                  stderr=subprocess.STDOUT, text=True)
        time.sleep(1)
        s.append(host("docker stop gl-curso-api-p4", ["docker", "stop", "gl-curso-api-p4"]))
        codigo = espera.communicate(timeout=60)[0].strip()
        s.append("$ docker logs -f gl-curso-api-p4\n" + logs.communicate(timeout=60)[0].rstrip() + "\n")
        s.append(f"[docker wait: código de salida {codigo}]\n")
    finally:
        run(["docker", "rm", "-f", "gl-curso-api-p4"])
        bajar_postgres()
        run(["docker", "image", "rm", "gl-curso-tareas:una-etapa", "gl-curso-tareas:p4"])
    return salida + "\n".join(s)


def parte5():
    """compose con la BD, health checks y un fallo provocado: la BD se cae y vuelve."""
    dir5 = AQUI / "parte5" / "tareas"
    red = "gl-curso-p5_default"
    api = "gl-curso-p5-api-1"

    def dc(texto, *args, timeout=900):
        rc, out = 0, ""
        r = subprocess.run(["docker", "compose", *args], cwd=dir5, capture_output=True, text=True,
                           encoding="utf-8", errors="replace", timeout=timeout)
        out = (r.stdout + r.stderr).replace("\r", "")
        return f"$ docker compose {texto}\n{out.rstrip()}\n" + (f"[código de salida: {r.returncode}]\n" if r.returncode else "")

    def salud_api():
        return run(["docker", "inspect", api, "--format", "{{.State.Health.Status}}"])[1].strip()

    def esperar_salud(estado, segundos=90):
        for _ in range(segundos * 2):
            if salud_api() == estado:
                return True
            time.sleep(0.5)
        return False

    ps = ("ps -a --format \"table {{.Service}}\\t{{.Status}}\"", "ps", "-a", "--format", "table {{.Service}}\t{{.Status}}")
    previo = en_go(5)
    s = ["@@PASO levantar"]
    try:
        s.append(dc("build -q", "build", "-q"))
        s.append(dc("up -d --wait", "up", "-d", "--wait"))
        s.append(dc(*ps))
        s.append(cliente(red, "curl -s -X POST api:8080/tareas -d '{\"titulo\":\"comprar pan\"}'"))
        s.append(cliente(red, "curl -s -X POST api:8080/tareas -d '{\"titulo\":\"llamar al dentista\"}'"))
        s.append(cliente(red, "curl -s -w 'HTTP %{http_code}\\n' api:8080/salud"))
        s.append(host(f"docker inspect {api} --format \"{{{{.State.Health.Status}}}}\"",
                      ["docker", "inspect", api, "--format", "{{.State.Health.Status}}"]))
        s.append(host(f"docker exec {api} /tareas salud", ["docker", "exec", api, "/tareas", "salud"]))

        s.append("@@PASO bd-caida")
        s.append(dc("stop db", "stop", "db"))
        s.append(cliente(red, "curl -s -w 'HTTP %{http_code} en %{time_total} s\\n' api:8080/tareas"))
        s.append(cliente(red, "curl -s -w 'HTTP %{http_code}\\n' -X POST api:8080/tareas -d '{\"titulo\":\"pagar la luz\"}'"))
        s.append(cliente(red, "curl -s -w 'HTTP %{http_code}\\n' api:8080/tareas/abc"))
        s.append(cliente(red, "curl -s -w 'HTTP %{http_code}\\n' api:8080/salud"))
        ok = esperar_salud("unhealthy")
        s.append(f"[esperé a que Docker marcara la API: {'unhealthy' if ok else 'no cambió en 90 s'}]\n")
        s.append(dc(*ps))
        s.append(host(f"docker exec {api} /tareas salud", ["docker", "exec", api, "/tareas", "salud"]))
        s.append(dc("logs api --no-log-prefix --tail 4", "logs", "api", "--no-log-prefix", "--tail", "4"))
        r = subprocess.run(["docker", "compose", "logs", "api", "--no-log-prefix"], cwd=dir5, capture_output=True,
                           text=True, encoding="utf-8", errors="replace")
        errores = [x for x in r.stdout.splitlines() if '"msg":"almacén"' in x]
        s.append("$ docker compose logs api --no-log-prefix | grep almacén\n" + "\n".join(errores) + "\n")
        s.append(host(f"docker inspect {api} --format \"{{{{.RestartCount}}}}\"",
                      ["docker", "inspect", api, "--format", "{{.RestartCount}}"]))

        s.append("@@PASO bd-vuelve")
        s.append(dc("start db", "start", "db"))
        ok = esperar_salud("healthy")
        s.append(f"[esperé a que Docker marcara la API: {'healthy' if ok else 'no cambió en 90 s'}]\n")
        s.append(dc(*ps))
        s.append(cliente(red, "curl -s -w 'HTTP %{http_code}\\n' api:8080/tareas"))
        s.append(host(f"docker inspect {api} --format \"{{{{.RestartCount}}}}\"",
                      ["docker", "inspect", api, "--format", "{{.RestartCount}}"]))

        s.append("@@PASO api-muere")
        # docker kill / compose kill cuentan como parada manual y la política de reinicio no actúa.
        # Para simular que el proceso se cae solo (como si lo matara el kernel por memoria), se le
        # manda SIGKILL desde el espacio de procesos del anfitrión de Docker. Desde dentro de su
        # propio espacio no se puede: el PID 1 ignora las señales que no maneja.
        s.append(host(f"docker inspect {api} --format \"{{{{.State.Pid}}}}\"",
                      ["docker", "inspect", api, "--format", "{{.State.Pid}}"]))
        pid = run(["docker", "inspect", api, "--format", "{{.State.Pid}}"])[1].strip()
        s.append(host(f"docker run --rm --pid host alpine:3.24.3 kill -KILL {pid}",
                      BASE + ["--pid", "host", "alpine:3.24.3", "kill", "-KILL", pid]))
        time.sleep(1)
        s.append(dc(*ps))
        s.append(host(f"docker inspect {api} --format \"{{{{.State.ExitCode}}}} {{{{.RestartCount}}}}\"",
                      ["docker", "inspect", api, "--format", "{{.State.ExitCode}} {{.RestartCount}}"]))
        esperar_salud("healthy", 60)
        s.append(dc(*ps))
        s.append(host(f"docker inspect {api} --format \"{{{{.RestartCount}}}}\"",
                      ["docker", "inspect", api, "--format", "{{.RestartCount}}"]))
        s.append(cliente(red, "curl -s -w 'HTTP %{http_code}\\n' api:8080/tareas"))

        s.append("@@PASO apagar")
        s.append(dc("stop api", "stop", "api"))
        s.append(dc("logs api --no-log-prefix --tail 3", "logs", "api", "--no-log-prefix", "--tail", "3"))
        s.append(host(f"docker inspect {api} --format \"{{{{.State.ExitCode}}}}\"",
                      ["docker", "inspect", api, "--format", "{{.State.ExitCode}}"]))
    finally:
        s.append(dc("down -v", "down", "-v"))
        s.append(host("docker image rm gl-curso-tareas:p5", ["docker", "image", "rm", "gl-curso-tareas:p5"]))
    return previo + "\n".join(s)


PARTES = {1: parte1, 2: parte2, 3: parte3, 4: parte4, 5: parte5}


def main():
    if "--limpiar" in sys.argv:
        bajar_postgres()
        for v in ("gl-curso-mod", "gl-curso-cache"):
            print(host(f"docker volume rm {v}", ["docker", "volume", "rm", v]))
        return 0
    pedidas = [int(a) for a in sys.argv[1:]] or sorted(PARTES)
    for p in pedidas:
        print(f"parte {p} ...")
        escribir(p, PARTES[p]())
    return 0


if __name__ == "__main__":
    sys.exit(main())
