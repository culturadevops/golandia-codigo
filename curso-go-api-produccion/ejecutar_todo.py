# -*- coding: utf-8 -*-
"""Corre las 13 lecciones del curso «Una API en Go de cero a producción» y regenera SALIDA.md.

Regla de la casa: nada de lo que sale en el vídeo se escribe a mano. Sale de aquí.
Solo hace falta Docker y Python 3.10+: Go, Postgres, act y curl corren en contenedores.

    python ejecutar_todo.py           todo; al final borra contenedores, red, imágenes y volúmenes
    python ejecutar_todo.py --dejar   deja Postgres y la API corriendo para jugar
    python ejecutar_todo.py --limpiar solo limpia

Todo lo que crea lleva el prefijo g685 (contenedores, red, volúmenes, imágenes), salvo la
imagen tareas:ci que construye el workflow dentro de act, que también se borra.
"""
import os
import platform
import re
import shutil
import subprocess
import sys
import tempfile
import time
from datetime import datetime, timezone
from pathlib import Path
from types import SimpleNamespace

AQUI = Path(__file__).resolve().parent
ENV = {**os.environ, "MSYS_NO_PATHCONV": "1", "DOCKER_CLI_HINTS": "false"}

GO = "golang:1.27.2"
PG = "postgres:18.6-alpine"
ALPINE = "alpine:3.24.3"
ACT = "v0.2.90"
RUNNER = "catthehacker/ubuntu:act-latest"

P = "g685"
RED_, CPG, CAPI, CACT = P, f"{P}-pg", f"{P}-api", f"{P}-act"
IMG = f"{P}-tareas"
APP = f"postgres://app:app@{CPG}:5432/tareas?sslmode=disable"
MIG = f"postgres://migrador:migrador@{CPG}:5432/tareas?sslmode=disable"
TEST_APP = f"postgres://app:app@{CPG}:5432/tareas_test?sslmode=disable"
TEST_MIG = f"postgres://migrador:migrador@{CPG}:5432/tareas_test?sslmode=disable"

OUT = []
FALLOS = []


# ── utilidades ────────────────────────────────────────────────────────────────

def sh(*a, entrada=None, timeout=900):
    r = subprocess.run(list(a), capture_output=True, text=True, env=ENV, encoding="utf-8",
                       errors="replace", input=entrada, timeout=timeout)
    r.salida = ((r.stdout or "") + (r.stderr or "")).rstrip()
    return r


def golang(script, env=(), red=True):
    """Corre un script de sh en golang:1.27.2 con el repo en /src y la caché de módulos."""
    a = ["docker", "run", "--rm", "-v", f"{AQUI}:/src", "-v", f"{P}-gomod:/go/pkg/mod",
         "-v", f"{P}-gocache:/root/.cache/go-build", "-w", "/src"]
    if red:
        a += ["--network", RED_]
    for e in env:
        a += ["-e", e]
    return sh(*a, GO, "sh", "-c", script)


def psql(usuario, base, sql):
    return sh("docker", "exec", "-e", f"PGPASSWORD={usuario}", CPG, "psql", "-h", "127.0.0.1",
              "-U", usuario, "-d", base, "-v", "ON_ERROR_STOP=1", "-c", sql)


def curl(args):
    """curl desde un contenedor en la red (golang:1.27.2 trae curl)."""
    return sh("docker", "run", "--rm", "--network", RED_, GO, "sh", "-c", f"curl -s {args}")


def titulo(t):
    print(f"  {t} ...", flush=True)
    OUT.extend([f"## {t}", ""])


def nota(t):
    OUT.extend([t, ""])


def bloque(mostrado, r, esperado=0, codigo=True):
    OUT.append("```")
    OUT.append(f"$ {mostrado}")
    if r.salida:
        OUT.append(r.salida)
    if codigo:
        OUT.append(f"[código de salida {r.returncode}]")
    OUT.extend(["```", ""])
    if esperado is not None and r.returncode != esperado:
        FALLOS.append(f"{mostrado}: código {r.returncode}, esperaba {esperado}")
    return r


def limpiar():
    sh("docker", "rm", "-f", "-v", CPG, CAPI, CACT, f"{P}-carga")
    sh("docker", "network", "rm", RED_)
    for t in ("ingenuo", "final", "con-cgo"):
        sh("docker", "rmi", "-f", f"{IMG}:{t}")
    sh("docker", "rmi", "-f", "tareas:ci")


def limpiar_volumenes():
    sh("docker", "volume", "rm", f"{P}-gomod", f"{P}-gocache")
    sh("docker", "rmi", "-f", RUNNER)   # la imagen del runner de act pesa más de 1 GB


def esperar_pg():
    seguidas = 0
    for _ in range(90):
        r = psql("app", "tareas", "select 1")
        seguidas = seguidas + 1 if r.returncode == 0 else 0
        if seguidas == 2:
            return
        time.sleep(1)
    raise SystemExit("Postgres no arrancó:\n" + sh("docker", "logs", CPG).salida[-2000:])


def levantar_pg():
    sh("docker", "network", "create", RED_)
    r = sh("docker", "run", "-d", "--rm", "--name", CPG, "--network", RED_, "-e", "POSTGRES_PASSWORD=postgres",
           "-v", f"{AQUI / 'sql' / 'roles.sql'}:/docker-entrypoint-initdb.d/roles.sql:ro", PG)
    if r.returncode:
        raise SystemExit(r.salida)
    esperar_pg()


def esperar_salud(estado, limite=60):
    t0 = time.time()
    while time.time() - t0 < limite:
        r = sh("docker", "inspect", "-f", "{{.State.Health.Status}}", CAPI)
        if r.stdout.strip() == estado:
            return round(time.time() - t0)
        time.sleep(1)
    return None


# ── lecciones ─────────────────────────────────────────────────────────────────

def versiones():
    titulo("Versiones")
    bloque("docker version --format '{{.Server.Version}}'", sh("docker", "version", "--format", "{{.Server.Version}}"))
    bloque("go version", golang("go version", red=False))
    bloque("postgres --version", sh("docker", "exec", CPG, "postgres", "--version"))
    bloque("go list -m github.com/jackc/pgx/v5 github.com/pressly/goose/v3 golang.org/x/crypto",
           golang("go list -m github.com/jackc/pgx/v5 github.com/pressly/goose/v3 golang.org/x/crypto", red=False))


def arbol():
    fuera = {".git", "lecciones", "__pycache__"}
    lineas = []
    for raiz, dirs, archivos in os.walk(AQUI):
        dirs[:] = sorted(d for d in dirs if d not in fuera)
        rel = Path(raiz).relative_to(AQUI)
        nivel = len(rel.parts)
        if nivel:
            lineas.append("  " * (nivel - 1) + rel.name + "/")
        for a in sorted(archivos):
            if a in ("SALIDA.md", "ejecutar_todo.py"):
                continue
            n = len((Path(raiz) / a).read_text(encoding="utf-8", errors="replace").splitlines())
            lineas.append("  " * nivel + f"{a}  ({n})")
    return "\n".join(lineas)


def l01():
    titulo("Lección 1 — La estructura: main solo llama a run()")
    nota("La plantilla (sin `lecciones/`), con las líneas de cada archivo:")
    OUT.extend(["```", arbol(), "```", ""])
    bloque("go run ./lecciones/01_estructura -modo fatal", golang("go run ./lecciones/01_estructura -modo fatal 2>&1"), 1)
    bloque("go run ./lecciones/01_estructura -modo run", golang("go run ./lecciones/01_estructura -modo run 2>&1"), 1)


def l02():
    titulo("Lección 2 — El router de net/http (Go 1.22+)")
    bloque("go run ./lecciones/02_router", golang("go run ./lecciones/02_router"))
    bloque("GODEBUG=httpmuxgo121=1 go run ./lecciones/02_router",
           golang("go run ./lecciones/02_router", env=["GODEBUG=httpmuxgo121=1"]))


def l03():
    titulo("Lección 3 — Middlewares: request id, registro con slog, recover y plazo")
    bloque("go run ./lecciones/03_middlewares", golang("go run ./lecciones/03_middlewares"))


def l04():
    titulo("Lección 4 — Configuración por entorno, validada al arrancar")
    bloque("go run ./lecciones/04_config", golang("go run ./lecciones/04_config"))
    nota("El binario de verdad, con el entorno mal:")
    bloque("PUERTO=80a TIEMPO_PETICION=30 go run ./cmd/api",
           golang("go build -o /tmp/api ./cmd/api && /tmp/api", env=["PUERTO=80a", "TIEMPO_PETICION=30"]), 1)


def l05():
    titulo("Lección 5 — Postgres: migraciones embebidas y un usuario sin DDL")
    bloque("psql -U postgres -f sql/roles.sql   (lo hizo el contenedor al crearse; así quedó)",
           psql_su("SELECT rolname, rolsuper, rolcreatedb FROM pg_roles WHERE rolname IN ('app','migrador') ORDER BY 1"))
    bloque("BD_URL_MIGRAR=…migrador… api migrar", golang("go run ./cmd/api migrar", env=[f"BD_URL_MIGRAR={MIG}"]))
    bloque("BD_URL_MIGRAR=…migrador… api migrar   (otra vez)", golang("go run ./cmd/api migrar", env=[f"BD_URL_MIGRAR={MIG}"]))
    nota("Con el usuario de la app, que es con el que corre el servidor. Sobre `tareas` no hay nada que "
         "aplicar y sale bien; sobre una base vacía (`vacia`, de migrador) tiene que crear tablas:")
    bloque("BD_URL_MIGRAR=…app…/tareas api migrar", golang("go run ./cmd/api migrar", env=[f"BD_URL_MIGRAR={APP}"]))
    psql_su("CREATE DATABASE vacia OWNER migrador")
    bloque("BD_URL_MIGRAR=…app…/vacia api migrar", golang("go run ./cmd/api migrar",
           env=[f"BD_URL_MIGRAR=postgres://app:app@{CPG}:5432/vacia?sslmode=disable"]), 1)
    bloque("psql -U app -c 'DROP TABLE tareas'", psql("app", "tareas", "DROP TABLE tareas"), 1)
    bloque("psql -U app -c 'CREATE TABLE copia (x int)'", psql("app", "tareas", "CREATE TABLE copia (x int)"), 1)
    bloque("psql -U app -c 'TRUNCATE tareas'", psql("app", "tareas", "TRUNCATE tareas"), 1)
    nota("Editar una migración ya aplicada (base aparte, `demo05`):")
    psql_su("CREATE DATABASE demo05 OWNER migrador")
    bloque("go run ./lecciones/05_migraciones", golang(
        "go run ./lecciones/05_migraciones",
        env=[f"DEMO_BD_URL=postgres://migrador:migrador@{CPG}:5432/demo05?sslmode=disable"]))


def psql_su(sql):
    return sh("docker", "exec", CPG, "psql", "-U", "postgres", "-d", "tareas", "-c", sql)


def l06():
    titulo("Lección 6 — Autenticación: sesiones en Postgres y bcrypt")
    bloque("go run ./lecciones/06_auth", golang("go run ./lecciones/06_auth", env=[f"APP_BD_URL={APP}"]))


def l07():
    titulo("Lección 7 — Validación y errores JSON coherentes")
    bloque("go run ./lecciones/07_errores", golang("go run ./lecciones/07_errores", env=[f"APP_BD_URL={APP}"]))


def l08():
    titulo("Lección 8 — Tests unitarios y de integración contra Postgres")
    golang("go clean -testcache")
    nota("Sin las variables de la base:")
    bloque("go test ./...", golang("go test ./internal/... ./cmd/..."))
    bloque("go test -v ./... | grep -E 'SKIP|sin TEST_BD_URL'",
           golang("go test -v ./internal/... 2>&1 | grep -E -- '--- SKIP|sin TEST_BD_URL'"))
    bloque("CI=true go test ./...", golang("go test ./internal/... 2>&1 | grep -v '^?'", env=["CI=true"]), None)
    db = [f"TEST_BD_URL={TEST_APP}", f"TEST_BD_URL_MIGRAR={TEST_MIG}"]
    nota("Con la base de tests:")
    bloque("go test ./...", golang("go test ./internal/... 2>&1 | grep -v '^?'", env=db))
    bloque("go test ./...   (otra vez)", golang("go test ./internal/... 2>&1 | grep -v '^?'", env=db))
    nota("Ahora se rompe la base de tests: la columna titulo pasa a llamarse nombre.")
    bloque("psql -U migrador -d tareas_test -c 'ALTER TABLE tareas RENAME COLUMN titulo TO nombre'",
           psql("migrador", "tareas_test", "ALTER TABLE tareas RENAME COLUMN titulo TO nombre"))
    bloque("go test ./...", golang("go test ./internal/... 2>&1 | grep -v '^?'", env=db))
    bloque("go test -count=1 ./...", golang("go test -count=1 ./internal/... 2>&1 | grep -v '^?' | grep -E '^(ok|FAIL|---)|no existe|does not exist' | head -12", env=db), None)
    psql("migrador", "tareas_test", "ALTER TABLE tareas RENAME COLUMN nombre TO titulo")
    nota("Base arreglada otra vez, y con el detector de carreras:")
    bloque("go test -race -count=1 ./...", golang("go test -race -count=1 ./internal/... 2>&1 | grep -v '^?'", env=db))
    unit = integ = 0
    for f in AQUI.glob("internal/**/*_test.go"):
        n = len(re.findall(r"^func Test", f.read_text(encoding="utf-8"), re.M))
        if "pruebabd.Abrir" in f.read_text(encoding="utf-8"):
            integ += n
        else:
            unit += n
    nota(f"Funciones de test: {unit} unitarias y {integ} de integración (las que llaman a `pruebabd.Abrir`).")


def l09():
    titulo("Lección 9 — Dockerfile multietapa, medido")
    for t, f in (("ingenuo", "lecciones/09_docker/Dockerfile.ingenuo"), ("final", "Dockerfile"),
                 ("con-cgo", "lecciones/09_docker/Dockerfile.con-cgo")):
        t0 = time.time()
        r = sh("docker", "build", "-q", "-f", str(AQUI / f), "-t", f"{IMG}:{t}", str(AQUI))
        r.salida = f"(construida en {time.time() - t0:.0f} s)" if r.returncode == 0 else r.salida
        bloque(f"docker build -f {f} -t {IMG}:{t} .", r)
    bloque(f"docker image ls {IMG}", sh("docker", "image", "ls", IMG, "--format", "{{.Repository}}:{{.Tag}}  {{.Size}}"))
    for t in ("ingenuo", "final"):
        bloque(f"docker inspect -f 'usuario={{{{.Config.User}}}}' {IMG}:{t}",
               sh("docker", "inspect", "-f", "usuario={{.Config.User}}", f"{IMG}:{t}"))
    bloque(f"docker run --rm {IMG}:ingenuo id", sh("docker", "run", "--rm", "--entrypoint", "id", f"{IMG}:ingenuo"))
    bloque(f"docker run --rm {IMG}:con-cgo", sh("docker", "run", "--rm", f"{IMG}:con-cgo"), None)
    bloque(f"docker run --rm {IMG}:final   (sin entorno)", sh("docker", "run", "--rm", f"{IMG}:final"), 1)


def arrancar_api(extra=()):
    sh("docker", "rm", "-f", CAPI)
    r = sh("docker", "run", "-d", "--rm", "--name", CAPI, "--network", RED_, "-e", f"BD_URL={APP}",
           *[x for e in extra for x in ("-e", e)], f"{IMG}:final")
    if r.returncode:
        raise SystemExit(r.salida)


def l10():
    titulo("Lección 10 — Healthcheck: vivo y listo")
    arrancar_api()
    s = esperar_salud("healthy")
    nota(f"Contenedor `{CAPI}` (imagen final). Docker lo marcó healthy a los {s} s.")
    fmt = "-w '  [%{http_code} en %{time_total} s]\\n'"
    bloque("curl /salud/vivo ; curl /salud/listo",
           curl(f"{fmt} http://{CAPI}:8080/salud/vivo; curl -s {fmt} http://{CAPI}:8080/salud/listo"))

    nota("Se pausa Postgres (`docker pause`): no se cae, deja de contestar. Lo peor para una sonda.")
    demo = subprocess.Popen(
        ["docker", "run", "--rm", "-i", "--network", RED_, "-v", f"{AQUI}:/src", "-v", f"{P}-gomod:/go/pkg/mod",
         "-v", f"{P}-gocache:/root/.cache/go-build", "-w", "/src", "-e", f"APP_BD_URL={APP}", GO,
         "sh", "-c", "go build -o /tmp/d ./lecciones/10_salud && /tmp/d"],
        stdin=subprocess.PIPE, stdout=subprocess.PIPE, stderr=subprocess.STDOUT, text=True, env=ENV)
    demo.stdout.readline()
    sh("docker", "pause", CPG)
    bloque("docker pause g685-pg; curl /salud/vivo ; curl /salud/listo",
           curl(f"{fmt} http://{CAPI}:8080/salud/vivo; curl -s {fmt} http://{CAPI}:8080/salud/listo"))
    resto, _ = demo.communicate("\n", timeout=60)
    OUT.extend(["```", "$ go run ./lecciones/10_salud   (con Postgres pausado)", resto.rstrip(), "```", ""])
    sh("docker", "unpause", CPG)
    esperar_salud("healthy")

    nota("Se despliega un binario más nuevo que el esquema: se simula quitando la migración 2.")
    psql("migrador", "tareas", "BEGIN; DROP TABLE tareas; DELETE FROM goose_db_version WHERE version_id = 2; COMMIT;")
    bloque("curl /salud/listo", curl(f"{fmt} http://{CAPI}:8080/salud/listo"))
    s = esperar_salud("unhealthy")
    bloque("docker inspect -f '{{.State.Health.Status}}' g685-api",
           sh("docker", "inspect", "-f", "{{.State.Health.Status}}", CAPI))
    nota(f"(unhealthy a los {s} s: intervalo de 5 s y 3 reintentos)")
    bloque("api migrar", golang("go run ./cmd/api migrar", env=[f"BD_URL_MIGRAR={MIG}"]))
    s = esperar_salud("healthy")
    bloque("curl /salud/listo", curl(f"{fmt} http://{CAPI}:8080/salud/listo"))
    nota(f"(healthy otra vez a los {s} s)")


def l11():
    titulo("Lección 11 — Apagado ordenado, medido")
    nota("20 peticiones de 2 s en vuelo; a los 300 ms, SIGTERM. Tres servidores:")
    bloque("go run ./lecciones/11_apagado", golang("go build -o /tmp/l ./lecciones/11_apagado && /tmp/l"))

    nota("Y con la imagen de verdad y `docker stop`. Para tener peticiones en vuelo, una sesión de psql "
         "bloquea la tabla tareas 3 s; las 20 peticiones GET /tareas esperan al candado.")
    arrancar_api()
    esperar_salud("healthy")
    j = '{"email":"stop@ejemplo.com","clave":"una clave larga"}'
    curl(f"-X POST -d '{j}' http://{CAPI}:8080/usuarios")
    tok = re.search(r'"token":"([^"]+)"', curl(f"-X POST -d '{j}' http://{CAPI}:8080/sesiones").salida).group(1)
    candado = subprocess.Popen(["docker", "exec", "-e", "PGPASSWORD=migrador", CPG, "psql", "-h", "127.0.0.1",
                                "-U", "migrador", "-d", "tareas", "-c",
                                "BEGIN; LOCK TABLE tareas IN ACCESS EXCLUSIVE MODE; SELECT pg_sleep(3); COMMIT;"],
                               stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL, env=ENV)
    time.sleep(0.5)
    carga = subprocess.Popen(
        ["docker", "run", "--rm", "--name", f"{P}-carga", "--network", RED_, GO, "sh", "-c",
         f"for i in $(seq 20); do curl -s -o /dev/null -w '%{{http_code}}\\n' -H 'Authorization: Bearer {tok}' "
         f"http://{CAPI}:8080/tareas & done; wait"],
        stdout=subprocess.PIPE, stderr=subprocess.STDOUT, text=True, env=ENV)
    # La API corre con --rm: al pararse desaparece. Logs y código de salida se recogen en vivo.
    seguir = subprocess.Popen(["docker", "logs", "-f", CAPI], stdout=subprocess.PIPE,
                              stderr=subprocess.STDOUT, text=True, env=ENV, encoding="utf-8", errors="replace")
    espera = subprocess.Popen(["docker", "wait", CAPI], stdout=subprocess.PIPE, stderr=subprocess.STDOUT,
                              text=True, env=ENV)
    time.sleep(1.5)
    t0 = time.time()
    r = sh("docker", "stop", CAPI)
    parada = time.time() - t0
    codigos, _ = carga.communicate(timeout=60)
    candado.wait(timeout=30)
    salida_api, _ = espera.communicate(timeout=60)
    logs, _ = seguir.communicate(timeout=60)
    cuenta = {}
    for c in codigos.split():
        cuenta[c] = cuenta.get(c, 0) + 1
    r.salida = f"{r.salida}\n(docker stop tardó {parada:.1f} s)"
    bloque("docker stop g685-api   (con las 20 peticiones esperando)", r)
    OUT.extend(["```", "$ códigos HTTP que recibieron las 20 peticiones",
                "  ".join(f"{k}: {v}" for k, v in sorted(cuenta.items())), "```", ""])
    bloque("docker logs -f g685-api | grep -E 'apagando|apagado'",
           SimpleNamespace(salida="\n".join(l for l in logs.splitlines() if '"apagando"' in l or '"apagado"' in l),
                           returncode=0))
    bloque("docker wait g685-api   (código con el que salió)",
           SimpleNamespace(salida=salida_api.strip(), returncode=espera.returncode))
    if cuenta.get("200", 0) != 20:
        FALLOS.append(f"docker stop: {cuenta}")


def volumenes_act():
    # Solo los de este workflow (ci) y sus dos trabajos: otros proyectos de la máquina pueden usar act.
    return {v for v in sh("docker", "volume", "ls", "-q").stdout.split()
            if v.startswith(("act-ci-probar-", "act-ci-imagen-"))}


def limpiar_act(antes=frozenset()):
    """Lo que act deja: contenedores act-*, sus volúmenes y la imagen tareas:ci."""
    viejos = [x for f in ("name=^act-ci-probar", "name=^act-ci-imagen")
              for x in sh("docker", "ps", "-aq", "--filter", f).stdout.split()]
    if viejos:
        sh("docker", "rm", "-f", "-v", *viejos)
    nuevos = sorted(volumenes_act() - set(antes))
    if nuevos:
        sh("docker", "volume", "rm", "-f", *nuevos)
    sh("docker", "rmi", "-f", "tareas:ci")


def l12():
    titulo("Lección 12 — CI de GitHub Actions, corrido en local con act")
    antes = volumenes_act()
    try:
        _l12()
    finally:
        sh("docker", "rm", "-f", "-v", CACT)
        limpiar_act(antes)


def _l12():
    sh("docker", "rm", "-f", CACT)
    sh("docker", "run", "-d", "--rm", "--name", CACT, "-v", "//var/run/docker.sock:/var/run/docker.sock",
       "-v", f"{AQUI}:/repo:ro", ALPINE, "sleep", "infinity")
    prep = (f"apk add -q curl git docker-cli >/dev/null && curl -sL https://github.com/nektos/act/releases/download/"
            f"{ACT}/act_Linux_x86_64.tar.gz | tar xz -C /usr/local/bin act && rm -rf /work && cp -r /repo /work && "
            "cd /work && rm -rf .git && git init -q && git -c user.name=ci -c user.email=ci@ejemplo.com add -A && "
            "git -c user.name=ci -c user.email=ci@ejemplo.com commit -qm ci")
    sh("docker", "exec", CACT, "sh", "-c", prep)
    bloque("act --version", sh("docker", "exec", CACT, "act", "--version"))
    bloque("act -l", sh("docker", "exec", "-w", "/work", CACT, "act", "-l"))
    t0 = time.time()
    r = sh("docker", "exec", "-w", "/work", CACT, "sh", "-c",
           f"act push -P ubuntu-latest={RUNNER} 2>&1; echo \"[act terminó con código $?]\"")
    dur = time.time() - t0
    utiles = [re.sub(r"^\[ci/", "[", l).replace("|", "").rstrip() for l in r.salida.splitlines()
              if re.search(r"Success - Main|Failure - |Job succeeded|Job failed|^\S+\s*\|\s*(ok|FAIL|---)\s|"
                           r"\| [0-9.]+MB$|act terminó", l)]
    r.salida = "\n".join(utiles) + f"\n(act tardó {dur:.0f} s en total)"
    bloque(f"act push -P ubuntu-latest={RUNNER}", r)
    if "Job failed" in r.salida or "código 0" not in r.salida:
        FALLOS.append("act")
    bloque(f"docker image inspect {RUNNER} (digest)",
           sh("docker", "image", "inspect", "-f", "{{index .RepoDigests 0}}", RUNNER))


def l13():
    titulo("Lección 13 — Usarla de plantilla")
    nuevo = "github.com/tu-usuario/inventario"
    script = (
        "set -e; T=$(mktemp -d); cp -r /src/. $T/; cd $T; "
        "rm -rf .git lecciones ejecutar_todo.py SALIDA.md; "
        f"echo \"imports a cambiar: $(grep -rl '\"tareas/' --include=*.go . | wc -l) archivos\"; "
        f"go mod edit -module {nuevo}; "
        f"grep -rl '\"tareas/' --include=*.go . | xargs sed -i 's#\"tareas/#\"{nuevo}/#'; "
        "echo \"quedan con el nombre viejo: $(grep -rl '\"tareas/' --include=*.go . | wc -l)\"; "
        "head -1 go.mod; go build ./... && go vet ./... && go test ./... 2>&1 | grep -v '^?'; "
        "echo \"archivos .go: $(find . -name '*.go' ! -name '*_test.go' | wc -l) de código, "
        "$(find . -name '*_test.go' | wc -l) de tests\"; "
        "echo \"líneas de Go: $(cat $(find . -name '*.go' ! -name '*_test.go') | wc -l) de código, "
        "$(cat $(find . -name '*_test.go') | wc -l) de tests\"; rm -rf $T"
    )
    bloque("copiar, renombrar el módulo, compilar y probar (en un directorio de mktemp)", golang(script, red=False))
    nota("docker compose de la plantilla, en una copia temporal y con los puertos quitados (para no chocar "
         "con nada de la máquina):")
    tmp = Path(tempfile.mkdtemp(prefix="g685-compose-"))
    try:
        for x in ("Dockerfile", ".dockerignore", "docker-compose.yml", "go.mod", "go.sum"):
            shutil.copy(AQUI / x, tmp / x)
        for d in ("cmd", "internal", "migraciones", "sql"):
            shutil.copytree(AQUI / d, tmp / d)
        (tmp / "sin-puertos.yml").write_text(
            "services:\n  bd:\n    ports: !reset []\n  api:\n    ports: !reset []\n", encoding="utf-8")
        dc = ["docker", "compose", "-p", f"{P}c", "-f", str(tmp / "docker-compose.yml"), "-f", str(tmp / "sin-puertos.yml")]
        t0 = time.time()
        r = sh(*dc, "up", "-d", "--build", "--wait", timeout=600)
        r.salida = f"(levantado en {time.time() - t0:.0f} s)" if r.returncode == 0 else r.salida[-1500:]
        bloque("docker compose up -d --build --wait", r)
        bloque("docker compose ps -a", sh(*dc, "ps", "-a", "--format", "{{.Service}}  {{.Status}}"))
        bloque("docker compose logs migrar", sh(*dc, "logs", "--no-log-prefix", "migrar"))
        bloque("curl http://api:8080/salud/listo", sh("docker", "run", "--rm", "--network", f"{P}c_default", GO,
                                                       "curl", "-s", "http://api:8080/salud/listo"))
    finally:
        sh(*dc, "down", "-v", "--rmi", "local", timeout=300)
        shutil.rmtree(tmp, ignore_errors=True)


def main():
    if hasattr(sys.stdout, "reconfigure"):
        sys.stdout.reconfigure(encoding="utf-8", errors="replace")
    if "--limpiar" in sys.argv:
        limpiar()
        limpiar_act()
        limpiar_volumenes()
        return 0
    sello = datetime.now(timezone.utc).strftime("%Y-%m-%d %H:%M UTC")
    limpiar()
    print("  levantando Postgres limpio ...", flush=True)
    levantar_pg()
    golang("go mod download", red=False)  # que «go: downloading …» no se cuele en las lecciones
    OUT.extend(["# SALIDA — curso «Una API en Go de cero a producción» (golandia 685)", "",
                "> Generado por `python ejecutar_todo.py`. **No se escribe a mano.**", "",
                f"- fecha: `{sello}` · sistema anfitrión: `{platform.system()} {platform.release()}`",
                f"- imágenes: `{GO}`, `{PG}`, `{ALPINE}`, act `{ACT}` con `{RUNNER}`",
                "- los tiempos dependen de la máquina: Docker Desktop sobre Windows", ""])
    try:
        for f in (versiones, l01, l02, l03, l04, l05, l06, l07, l08, l09, l10, l11, l12, l13):
            f()
        OUT.extend(["## Resultado", "", f"Fallos inesperados: {len(FALLOS)}", ""] + [f"- {x}" for x in FALLOS])
        (AQUI / "SALIDA.md").write_text("\n".join(OUT) + "\n", encoding="utf-8", newline="\n")
    finally:
        # Siempre, también si una lección revienta: no quedan contenedores, volúmenes ni imágenes g685.
        if "--dejar" not in sys.argv:
            limpiar()
            limpiar_volumenes()
    print(f"\n  -> SALIDA.md regenerado ({len(FALLOS)} fallos inesperados)")
    for x in FALLOS:
        print("     ·", x)
    return 1 if FALLOS else 0


if __name__ == "__main__":
    sys.exit(main())
